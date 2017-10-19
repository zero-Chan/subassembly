package timer

import (
	"testing"

	cconf "code-lib/conf"
	"code-lib/notify/rabbitmq"
	"code-lib/redis"

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

		Persistence: &conf.RedisPersistenceConf{
			Single: &redis.RedisClientConf{
				Addr: cconf.AddrConf{
					Host: "localhost",
					Port: 10379,
				},
			},
		},
	}

	timer, err := NewTimer(cfg)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	err = timer.Run()
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

}
