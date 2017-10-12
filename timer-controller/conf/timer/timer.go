package timer

import (
	"time"

	"subassembly/timer-controller/conf/commons"
)

type TimerConf struct {
	MQ *commons.RabbitmqConf
	// Redis

	// 定时器触发周期
	PollCycle time.Duration
}
