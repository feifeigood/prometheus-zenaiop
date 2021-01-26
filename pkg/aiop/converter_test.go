package aiop

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
)

const promAlerts = `
{
        "receiver": "cdn-web-teams",
        "status": "firing",
        "alerts": [
            {
                "status": "resolved",
                "labels": {
                    "address": "45.40.58.70",
                    "alertname": "网卡进剧增",
                    "cdnclass": "网页加速",
                    "datacenter": "印度马哈拉施特拉邦孟买机房",
                    "device": "lan0",
                    "environment": "prod",
                    "event": "网卡(lan0)进带宽剧增",
                    "instance": "45.40.58.70:9100",
                    "job": "node",
                    "layer": "C",
                    "monitor_cluster": "lax3",
                    "platform": "ZEN",
                    "region": "AP2",
                    "severity": "warning"
                },
                "annotations": {
                    "description": "节点: 45.40.58.70, 网卡(lan0)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 280.24%",
                    "summary": "节点45.40.58.70网卡(lan0)进带宽剧增"
                },
                "startsAt": "2020-08-13T07:35:08.299083471Z",
                "endsAt": "2020-08-13T07:37:08.299083471Z",
                "generatorURL": "http://prometheus-1:9090/graph?g0.expr=%28%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D%29+%2F+1000+%2F+1000+%2A+8%29+%2F+%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D+offset+5m%29+%2F+1000+%2F+1000+%2A+8%29+-+1%29+%2A+100+%3E+100&g0.tab=1",
                "fingerprint": "fc2fd639a684b991"
            },
            {
                "status": "resolved",
                "labels": {
                    "address": "45.40.58.71",
                    "alertname": "网卡进剧增",
                    "cdnclass": "网页加速",
                    "datacenter": "印度马哈拉施特拉邦孟买机房",
                    "device": "lo",
                    "environment": "prod",
                    "event": "网卡(lo)进带宽剧增",
                    "instance": "45.40.58.71:9100",
                    "job": "node",
                    "layer": "C",
                    "monitor_cluster": "lax3",
                    "platform": "ZEN",
                    "region": "AP2",
                    "severity": "warning"
                },
                "annotations": {
                    "description": "节点: 45.40.58.71, 网卡(lo)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 101.81%",
                    "summary": "节点45.40.58.71网卡(lo)进带宽剧增"
                },
                "startsAt": "2020-08-13T07:35:08.299083471Z",
                "endsAt": "2020-08-13T07:37:08.299083471Z",
                "generatorURL": "http://prometheus-1:9090/graph?g0.expr=%28%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D%29+%2F+1000+%2F+1000+%2A+8%29+%2F+%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D+offset+5m%29+%2F+1000+%2F+1000+%2A+8%29+-+1%29+%2A+100+%3E+100&g0.tab=1",
                "fingerprint": "7a926722de2132a7"
            },
            {
                "status": "firing",
                "labels": {
                    "address": "45.40.58.71",
                    "alertname": "网卡进剧增",
                    "cdnclass": "网页加速",
                    "datacenter": "印度马哈拉施特拉邦孟买机房",
                    "device": "wan0",
                    "environment": "prod",
                    "event": "网卡(wan0)进带宽剧增",
                    "instance": "45.40.58.71:9100",
                    "job": "node",
                    "layer": "C",
                    "monitor_cluster": "lax3",
                    "platform": "ZEN",
                    "region": "AP2",
                    "severity": "warning"
                },
                "annotations": {
                    "description": "节点: 45.40.58.71, 网卡(wan0)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 221.61%",
                    "summary": "节点45.40.58.71网卡(wan0)进带宽剧增"
                },
                "startsAt": "2020-08-13T07:36:08.299083471Z",
                "endsAt": "0001-01-01T00:00:00Z",
                "generatorURL": "http://prometheus-1:9090/graph?g0.expr=%28%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D%29+%2F+1000+%2F+1000+%2A+8%29+%2F+%28rate%28node_network_receive_bytes_total%7Bplatform%3D%22ZEN%22%7D%5B5m%5D+offset+5m%29+%2F+1000+%2F+1000+%2A+8%29+-+1%29+%2A+100+%3E+100&g0.tab=1",
                "fingerprint": "30c1bc9c795935f5"
            }
        ],
        "groupLabels": {
            "alertname": "网卡进剧增"
        },
        "commonLabels": {
            "alertname": "网卡进剧增",
            "cdnclass": "网页加速",
            "datacenter": "印度马哈拉施特拉邦孟买机房",
            "environment": "prod",
            "job": "node",
            "layer": "C",
            "monitor_cluster": "lax3",
            "platform": "ZEN",
            "region": "AP2",
            "severity": "warning"
        },
        "commonAnnotations": {},
        "externalURL": "http://alertmanager-1:9093",
        "version": "4",
        "groupKey": "{}/{platform=~\"^(?:ZEN)$\"}:{alertname=\"网卡进剧增\"}"
}
`

func TestAIOPCreator(t *testing.T) {

	want := []Alert{
		{
			ID:      757611078,
			Level:   1,
			Type:    "网卡进剧增",
			Message: "节点: 45.40.58.70, 网卡(lan0)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 280.24%",
			Time:    "2020.08.13.15:35:08",
			Status:  "RESOLVED",
			Owner:   "",
			Infor:   "45.40.58.70",
		},
		{
			ID:      757611079,
			Level:   1,
			Type:    "网卡进剧增",
			Message: "节点: 45.40.58.71, 网卡(lo)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 101.81%",
			Time:    "2020.08.13.15:35:08",
			Status:  "RESOLVED",
			Owner:   "",
			Infor:   "45.40.58.71",
		},
		{
			ID:      757611079,
			Level:   1,
			Type:    "网卡进剧增",
			Message: "节点: 45.40.58.71, 网卡(wan0)进带宽持续5分钟剧增(对比前5分钟数据) \\u003e 100%, 当前值: 221.61%",
			Time:    "2020.08.13.15:36:08",
			Status:  "PROBLEM",
			Owner:   "",
			Infor:   "45.40.58.71",
		},
	}

	var wm webhook.Message
	err := json.Unmarshal([]byte(promAlerts), &wm)
	if err != nil {
		t.Error(err)
	}

	creator := NewAIOPAlertCreator()
	actual, _ := creator.Convert(wm)
	if !reflect.DeepEqual(want, actual) {
		t.Errorf("expected %v, but got %v", want, actual)
	}
}

func BenchmarkConvert(b *testing.B) {
	var wm webhook.Message
	json.Unmarshal([]byte(promAlerts), &wm)

	creator := NewAIOPAlertCreator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		creator.Convert(wm)
	}
}
