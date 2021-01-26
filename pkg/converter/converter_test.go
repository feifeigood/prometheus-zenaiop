package converter

import (
	"testing"
	"time"
)

func TestTimeIn(t *testing.T) {
	for _, name := range []string{
		"",
		"Local",
		"Asia/Shanghai",
		"America/Metropolis",
	} {
		tm, err := TimeIn(time.Now(), name)
		if err == nil {
			t.Logf("TimeIn => %s %s", tm.Location(), tm.Format("15:04"))
		} else {
			t.Logf("TimeIn => %s <time unknow>", name)
		}
	}
}
