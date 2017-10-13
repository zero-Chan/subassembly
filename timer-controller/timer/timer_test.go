package timer

import (
	"testing"

	"code-lib/notify/rabbitmq"

	conf "subassembly/timer-controller/conf/timer"
)

func Test_Timer(t *testing.T) {
	cfg := &conf.TimerConf{
		PollCycle: 1e9,
		MQ: &rabbitmq.RabbitNotifyConf{
			RabbitClientConf: &rabbitmq.RabbitClientConf{
				Host:     "localhost",
				Port:     5672,
				UserName: "guest",
				Password: "guest",
				VHost:    "/",
			},
			ConsumerInuse: true,
			Exchange:      "cza.test.timer",
			Kind:          "direct",
		},
	}

	_, err := NewTimer(cfg)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
}
