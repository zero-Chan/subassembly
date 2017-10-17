package timer

import (
	"time"

	"code-lib/notify/rabbitmq"
	"code-lib/redis"
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

	Persistence *RedisPersistenceConf `json:"Persistence"`
}

// 支持单节点模式和cluster模式
type RedisPersistenceConf struct {
	// redis-cli
	Cluster *redis.RedisClusterClientConf `json:"Cluster"`
	Single  *redis.RedisClientConf        `json:"Single"`

	// sessionID
	SessionIDFile string `json:"SessionIDFile"`
}
