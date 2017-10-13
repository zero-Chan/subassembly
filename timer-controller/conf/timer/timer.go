package timer

import (
	"time"

	"code-lib/notify/rabbitmq"
)

type TimerConf struct {
	// 定时器触发周期
	// Nanosecond  Duration = 1
	// Microsecond          = 1e3
	// Millisecond          = 1e6
	// Second               = 1e9
	// Minute               = 6e10
	// Hour                 = 3.6e12
	PollCycle time.Duration `json:"PollCycle"`

	MQ *rabbitmq.RabbitNotifyConf `json:"MQ"`
	// Redis

}
