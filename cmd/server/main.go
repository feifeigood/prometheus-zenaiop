package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/feifeigood/prometheus-zenaiop/pkg/converter"
	"github.com/feifeigood/prometheus-zenaiop/pkg/log"
	"github.com/feifeigood/prometheus-zenaiop/pkg/service"
	"github.com/feifeigood/prometheus-zenaiop/pkg/version"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/notify/webhook"
	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var (
		printVersion = flag.Bool("version", false, "show program version")
		listenAddr   = flag.String("web.listen-address", ":9299", "address on which the server will listen on")
		logLevel     = flag.String("log.level", "debug", "log message output level")
		webhookURL   = flag.String("aiop.webhook", "", "aiop webhook url")
	)

	flag.Parse()

	if *printVersion {
		fmt.Println(version.VERSION)
		os.Exit(0)
	}

	if err := log.Init(log.Options{Level: *logLevel}); err != nil {
		panic(err)
	}

	if _, err := url.ParseRequestURI(*webhookURL); err != nil {
		panic(err)
	}

	zap.S().Infof("starting prometheus-zenaiop version %s build_date %s", version.VERSION, version.BUILDDATE)

	r := gin.New()
	r.Use(ginzap.RecoveryWithZap(zap.L(), true))
	r.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))

	// creates Alertmanager webhook message converter chains
	cvt := converter.New()
	// build PostMessage service
	svc := service.NewSimpleService(cvt, *webhookURL)

	r.POST("/api/v1/zenlayer/aiop", func(c *gin.Context) {
		var wm webhook.Message
		if err := c.ShouldBindJSON(&wm); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if _, err := svc.Post(wm); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	var (
		term = make(chan os.Signal, 1)
		srvc = make(chan struct{})
	)

	srv := http.Server{
		Addr:    *listenAddr,
		Handler: r,
	}

	go func() {
		zap.S().Infof("httpserver listening on %s", *listenAddr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.S().Errorf("httpserver listen err: %v", err)
			close(srvc)
		}

		defer func() {
			if err := srv.Close(); err != nil {
				zap.S().Errorf("error on closing the httpserver: %v", err)
			}
		}()
	}()

	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			zap.S().Info("received SIGTERM, exiting gracefully...")
			os.Exit(0)
		case <-srvc:
			os.Exit(1)
		}
	}
}
