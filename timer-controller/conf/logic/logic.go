package logic

import (
	"code-lib/notify/rabbitmq"
)

type LogicConf struct {
	TimerMQ *rabbitmq.RabbitNotifyConf `json:"TimerMQ"`
}
