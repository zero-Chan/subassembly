package timer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"code-lib/notify"
	rabbitnotify "code-lib/notify/rabbitmq"

	conf "subassembly/timer-controller/conf/timer"
	proto "subassembly/timer-controller/proto/notify"
	"subassembly/timer-controller/timer/persistence"
)

const (
	defaultQueueName = "controller_${PollCycle}"
)

type Timer struct {
	cfg *conf.TimerConf

	// 定时器触发周期
	// 订阅队列为: ${cfg.MQ.QueueName}_${PollCycle}
	PollCycle time.Duration

	ConsumeMQ notify.Notify

	PublishMQs map[proto.RabbitmqDestination]notify.Notify

	Persistence persistence.Persistence
}

func CreateTimer(cfg *conf.TimerConf) (timer Timer, err error) {
	ErrorPrefix := "[InitError] `Func: CreateTimer` "

	if cfg == nil {
		err = fmt.Errorf(ErrorPrefix + "`Reason: cfg is nil.`")
		return
	}

	timer = Timer{
		cfg:       cfg,
		PollCycle: cfg.PollCycle,
	}

	err = timer.configCheck()
	if err != nil {
		err = fmt.Errorf(ErrorPrefix + "`Reason: cfg is nil.`")
	}

	// MQ
	timer.ConsumeMQ, err = rabbitnotify.NewRabbitNotify(cfg.MQ)
	if err != nil {
		return
	}

	// Persistence
	timer.Persistence = persistence.NewHashMap()

	return
}

func NewTimer(cfg *conf.TimerConf) (timer *Timer, err error) {
	t, err := CreateTimer(cfg)
	return &t, err
}

func (this *Timer) configCheck() (err error) {
	if this.cfg == nil {
		err = fmt.Errorf("Timer.cfg is nil.")
		return
	}

	if this.cfg.MQ == nil {
		err = fmt.Errorf("Timer.cfg.MQ is nil.")
		return
	}

	cfg := this.cfg
	mqcfg := this.cfg.MQ

	if cfg.PollCycle <= 0 {
		err = fmt.Errorf("Timer.cfg.PollCycle[%s] invalid.", cfg.PollCycle.String())
	}

	timerDefQueue := strings.Replace(defaultQueueName, "${PollCycle}", cfg.PollCycle.String(), -1)

	if mqcfg.RoutingKey == "" {
		mqcfg.RoutingKey = timerDefQueue
	}

	if mqcfg.QueueName == "" {
		mqcfg.QueueName = timerDefQueue
	}

	mqcfg.ConsumerInuse = true

	return
}

func (this *Timer) Run() (err error) {
	err = this.ConsumeMQ.Receive()
	if err != nil {
		return
	}

	return
}

func (this *Timer) polling() {
	PollCycle := time.NewTicker(this.PollCycle)
	for {
		select {
		// 定时器到期
		case now := <-PollCycle.C:
			datas := this.Persistence.Get(now)
			errlist := this.sendDatas(datas)

			// make delete list
			var deletelist persistence.DeleteTimeList = append(errlist, time.Unix(0, 0), now)
			sort.Sort(deletelist)

			this.Persistence.Delete(deletelist[0], deletelist[1], deletelist[2:]...)
		}
	}
}

func (this *Timer) sendDatas(datas [][]byte) (errlist []time.Time) {
	var err error

	errlist = make([]time.Time, 0)
	for _, data := range datas {
		val := proto.TimerNotice{}
		err = json.Unmarshal(data, &val)
		if err != nil {
			// TODO
			// log.Notice
			continue
		}

		err = this.sendData(data, val.Destination)
		if err != nil {
			// TODO
			// log.Notice
			errlist = append(errlist, val.SendUnixTime.Add(val.Expire))
			continue
		}
		// 失败的不要delete, 下一次重试
		// TODO
		// 如果使用confirm模式则delete需要改进
		// 而且要考虑是　每次单条delete(确保性高) 还是　批量delete(效率高). 现在先采用批量delete
	}

	return
}

func (this *Timer) sendData(data []byte, dest proto.RabbitmqDestination) (err error) {
	// Get publish notify
	// TODO:
	// 如果rabbit使用confirm模式，则是收到应答的时候才delete消息
	// 第一版默认推送到Channel就算成功

	pnotify, ok := this.PublishMQs[dest]
	if !ok {
		pnotify, err = rabbitnotify.NewRabbitNotify(&rabbitnotify.RabbitNotifyConf{
			RabbitClientConf: &rabbitnotify.RabbitClientConf{
				Host:     this.cfg.MQ.Host,
				Port:     this.cfg.MQ.Port,
				UserName: this.cfg.MQ.UserName,
				Password: this.cfg.MQ.Password,
				VHost:    this.cfg.MQ.VHost,
			},
			PublisherInuse: true,
			Exchange:       dest.Exchange,
			RoutingKey:     dest.RoutingKey,
		})
		if err != nil {
			return
		}
		this.PublishMQs[dest] = pnotify
	}

	err = pnotify.Push(data)

	return
}
