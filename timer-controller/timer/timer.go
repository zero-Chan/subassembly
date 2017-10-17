package timer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code-lib/log"
	"code-lib/log/golog"

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

	log log.Logger

	// 定时器触发周期
	// 订阅队列为: ${cfg.MQ.QueueName}_${PollCycle}
	pollCycle time.Duration

	// 消费队列
	consumeMQ notify.Notify

	// 目的地队列
	publishMQs map[proto.RabbitmqDestination]notify.Notify

	// 持久化
	persistence persistence.Persistence

	// 停止信号
	stopStart chan bool
	stopEnd   chan bool
	isstop    bool
}

func CreateTimer(cfg *conf.TimerConf) (timer Timer, err error) {
	ErrorPrefix := "[InitError] `Func: CreateTimer` "

	if cfg == nil {
		err = fmt.Errorf(ErrorPrefix + "`Reason: cfg is nil.`")
		return
	}

	timer = Timer{
		cfg:        cfg,
		log:        golog.NewDefault().Virtualize(),
		pollCycle:  cfg.PollCycle,
		publishMQs: make(map[proto.RabbitmqDestination]notify.Notify),
		stopStart:  make(chan bool),
		stopEnd:    make(chan bool),
		isstop:     false,
	}

	err = timer.configCheck()
	if err != nil {
		err = fmt.Errorf(ErrorPrefix+"`Reason: %s`", err)
		return
	}

	// MQ
	timer.consumeMQ, err = rabbitnotify.NewRabbitNotify(cfg.MQ)
	if err != nil {
		return
	}

	err = timer.consumeMQ.Init()
	if err != nil {
		return
	}

	// Persistence - redis
	//	timer.persistence = persistence.NewHashMap()
	timer.persistence, err = persistence.NewRedis(timer.cfg.Persistence)
	if err != nil {
		err = fmt.Errorf(ErrorPrefix+"`Reason: %s`", err)
		return
	}

	// Notice
	timer.log.Noticef("Timer[%s] Start!!!", timer.Name())

	return
}

func (this *Timer) Close() (err error) {
	ErrorPrefix := "[CloseError] `Func: Timer.Close` "

	// TODO
	// close consumerMQ. 先关闭入口
	//	err = this.consumeMQ.Close()

	// 关闭polling
	this.stopStart <- true
	<-this.stopEnd

	// close persistence
	err = this.persistence.Close()
	if err != nil {
		// only log.Error
		this.log.Errorf(ErrorPrefix+"`Reason: %s`", err)
	}

	// close publishMQs
	// err = for range this.publishMQs.Close()

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

func (this *Timer) SetLog(log log.Logger) {
	this.log = log
}

// Name
// ${Exchange}/${queue}
func (this *Timer) Name() string {
	return this.cfg.MQ.Exchange + "/" + this.cfg.MQ.QueueName
}

func (this *Timer) Run() (err error) {
	err = this.consumeMQ.Receive()
	if err != nil {
		return
	}

	this.polling()

	return
}

func (this *Timer) polling() {
	ErrorPrefix := "[TimerPollError] `Func: Time.polling` "

	var (
		err       error
		PollCycle = time.NewTicker(this.pollCycle)
	)

	for {
		select {
		case data, ok := <-this.consumeMQ.Pop():
			if !ok {
				this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] pop from consumeMQ fail.`", this.Name())
				continue
			}

			val := proto.TimerNotice{}
			err = json.Unmarshal(data, &val)
			if err != nil {
				this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] get invalid protocol.` `params: proto=%s`", this.Name(), string(data))
				continue
			}

			// Notice
			this.log.Noticef("Timer[%s] Get data: %s", this.Name(), string(data))

			expiretime := val.SendTime.Add(val.Expire)
			err = this.persistence.Set(expiretime, data)
			if err != nil {
				// only log.Errorf
				this.log.Errorf(ErrorPrefix+"Reason: Timer[%s] persistence.Set data fail: %s", this.Name(), err)
			}

			err = this.consumeMQ.Ack()
			if err != nil {
				this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] ack fail: %s`", this.Name(), err)
				continue
			}

		// 定时器到期
		case now := <-PollCycle.C:
			datas, err := this.persistence.Get(now)
			if err != nil {
				this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] persistence.Get datas fail: %s`", this.Name(), err)
				continue
			}

			errlist := this.sendDatas(datas)

			// make delete list
			deletelist := persistence.NewDeleteTimeList(time.Unix(0, 0), now, errlist...)
			err = this.persistence.Delete(deletelist[0], deletelist[1], deletelist[2:]...)
			if err != nil {
				this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] persistence.Delete datas fail: %s`", this.Name(), err)
				continue
			}

			if this.isstop {
				this.stopEnd <- true
				return
			}

		// 退出信号
		case <-this.stopStart:
			this.isstop = true
		}
	}
}

func (this *Timer) sendDatas(datas [][]byte) (errlist []time.Time) {
	ErrorPrefix := "[TimerSendError] `Func: Time.sendDatas` "

	var err error

	errlist = make([]time.Time, 0)
	for _, data := range datas {
		val := proto.TimerNotice{}
		err = json.Unmarshal(data, &val)
		if err != nil {
			this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] get invalid proto from persistence.` `params: proto=%s`", this.Name(), string(data))
			continue
		}

		err = this.sendData(val.Target, val.Destination)
		if err != nil {
			// log.Notice
			errlist = append(errlist, val.SendTime.Add(val.Expire))
			this.log.Errorf(ErrorPrefix+"`Reason: Timer[%s] push data to rabbitmq fail.` `params: data=%s dest=%+v`", this.Name(), string(data), val.Destination)
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

	pnotify, ok := this.publishMQs[dest]
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

		// no init rabbitmq exchange / queue. publishMQs will publish whether exchange/queue exist or not.
		this.publishMQs[dest] = pnotify
	}

	err = pnotify.Push(data)
	if err != nil {
		return
	}

	// Notice
	this.log.Noticef("Timer[%s] send data: %s", this.Name(), string(data))

	return
}
