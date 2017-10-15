package logic

import (
	"code-lib/notify/rabbitmq"
)

type BaseLogicConf struct {
	TimerMQ *rabbitmq.RabbitNotifyConf `json:"TimerMQ"`
}
