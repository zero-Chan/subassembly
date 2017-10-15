package logic

import (
	"encoding/json"
	"fmt"
	"time"

	"code-lib/notify"
	rabbitnotify "code-lib/notify/rabbitmq"

	conf "subassembly/timer-controller/conf/logic"
	proto "subassembly/timer-controller/proto/notify"
)

type BaseLogic struct {
	cfg       *conf.BaseLogicConf `json:"-"`
	publishMQ notify.Notify       `json:"-"`
}

func CreateBaseLogic(cfg *conf.BaseLogicConf) (logic BaseLogic, err error) {
	logic = BaseLogic{
		cfg: cfg,
	}

	logic.publishMQ, err = rabbitnotify.NewRabbitNotify(cfg.TimerMQ)
	if err != nil {
		return
	}

	// tips: exchange and queue should create by timer, so do MQ.Init without here.

	return
}

func NewBaseLogic(cfg *conf.BaseLogicConf) (logic *BaseLogic, err error) {
	l, err := CreateBaseLogic(cfg)

	return &l, err
}

func (this *BaseLogic) Push2Timer(target []byte, expire time.Duration, dest proto.RabbitmqDestination) (err error) {
	ErrorPrefix := "[RabbitPushError] `Func: BaseLogic.Push2Timer` "

	notifyMsg := proto.TimerNotice{
		Destination: dest,
		SendTime:    time.Now(),
		Expire:      expire,
		Target:      json.RawMessage(target),
	}

	sendbuf, err := json.Marshal(notifyMsg)
	if err != nil {
		err = fmt.Errorf(ErrorPrefix+"`Reason: %s`", err)
		return
	}

	err = this.publishMQ.Push(sendbuf)
	if err != nil {
		return
	}

	return
}
