package converter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"
)

// AIOPAlerts a list of AIOPAlert
type AIOPAlerts []AIOPAlert

// AIOPAlert represents Zenlayer AIOP Alert structure
// http://10.64.13.30:5006/api/v1.0/recive_alert/inter_alarm/SRE
type AIOPAlert struct {
	// ID is identifier for device using int type
	ID uint32 `json:"id"`
	// Level is message priority level, the level order is 1 < 2 < 3
	Level int `json:"level"`
	// Type is message name for alert
	Type string `json:"type"`
	// Message is deatil for aiop
	Message string `json:"message"`
	// Infor represents resource of report message
	Infor string `json:"infor"`
	// Time message startAt, date format is 2020.08.07 15:22:12
	Time string `json:"time"`
	// Status message current status  PROBLEM/RESOLVED
	Status string `json:"status"`
	//Owner the people on call
	Owner string `json:"owner"`
}

// Converter converts an alert manager webhook message to AIOP format
type Converter interface {
	Convert(*AIOPAlerts, webhook.Message) error
	SetNext(Converter)
}

// New creates a new Alertmanager webhook message converter object.
func New() Converter {

	unknow := &unknowAlert{}

	biz := &bizAlert{}
	biz.SetNext(unknow)

	node := &nodeAlert{}
	node.SetNext(biz)

	return node
}

var (
	// ShanghaiTZ timezone shanghai
	ShanghaiTZ = "Asia/Shanghai"
)

// TimeIn returns the time in UTC if the name is "" or "UTC".
// It returns the local time if the name is "Local".
// Otherwise, the name is taken to be a location name in
// the IANA Time Zone database, such as "Africa/Lagos".
func TimeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}

	return t, err
}

const (
	offset32 = 2166136261
	prime32  = 16777619
)

// hashNew initializies a new fnv32a hash value.
func hashNew() uint32 {
	return offset32
}

// hashAdd adds a string to a fnv32a hash value, returning the updated hash.
func hashAdd(h uint32, s string) uint32 {
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}

// hashAddByte adds a byte to a fnv32a hash value, returning the updated hash.
func hashAddByte(h uint32, b byte) uint32 {
	h ^= uint32(b)
	h *= prime32
	return h
}

// A LabelName is a key for a LabelSet or Metric.  It has a value associated
// therewith.
type LabelName string

// LabelNames is a sortable LabelName slice. In implements sort.Interface.
type LabelNames []LabelName

func (l LabelNames) Len() int {
	return len(l)
}

func (l LabelNames) Less(i, j int) bool {
	return l[i] < l[j]
}

func (l LabelNames) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l LabelNames) String() string {
	labelStrings := make([]string, 0, len(l))
	for _, label := range l {
		labelStrings = append(labelStrings, string(label))
	}
	return strings.Join(labelStrings, ", ")
}

// SeparatorByte is a byte that cannot occur in valid UTF-8 sequences and is
// used to separate label names, label values, and other strings from each other
// when calculating their combined hash value (aka signature aka fingerprint).
const SeparatorByte byte = 255

var (
	// cache the signature of an empty label set.
	emptyLabelSignature = hashNew()
)

// FormatAIOPID converts labels to id
func FormatAIOPID(ls template.KV) uint32 {
	labelNames := make(LabelNames, 0, len(ls))
	for labelName := range ls {
		labelNames = append(labelNames, LabelName(labelName))
	}
	sort.Sort(labelNames)

	sum := hashNew()
	for _, labelName := range labelNames {
		sum = hashAdd(sum, string(labelName))
		sum = hashAddByte(sum, SeparatorByte)
		sum = hashAdd(sum, string(ls[string(labelName)]))
		sum = hashAddByte(sum, SeparatorByte)
	}

	return sum
}

// FormatAIOPTime format time to AIOP time
func FormatAIOPTime(t time.Time) string {
	shanghai, _ := TimeIn(t, ShanghaiTZ)
	return fmt.Sprintf("%d.%02d.%02d.%02d:%02d:%02d", shanghai.Year(), shanghai.Month(), shanghai.Day(), shanghai.Hour(), shanghai.Minute(), shanghai.Second())
}

// FormatAIOPLevel converts prometheus severity to aiop alerting level code
func FormatAIOPLevel(severity string) int {
	switch severity {
	case "emergency":
		return 3
	case "critical":
		return 2
	default: // warning info or other using lowest level
		return 1
	}
}

// FormatAIOPStatus converts status to aiop status
func FormatAIOPStatus(status string) string {
	if status == "firing" {
		return "PROBLEM"
	}
	return "RESOLVED"
}

func outputJSON(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}
