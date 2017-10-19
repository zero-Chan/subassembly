package persistence

import (
	"testing"
	"time"
)

func Test_Deletelist(t *testing.T) {
	others := make([]time.Time, 0)
	others = append(others, time.Unix(0, 100), time.Unix(1, 0))

	list := NewDeleteTimeList(time.Unix(0, 0), time.Now(), others...)

	for idx, clock := range list {
		t.Logf("list[%d] = %d", idx, clock.UnixNano())
	}
}
