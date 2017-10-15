package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"

	logi "code-lib/log"
	"code-lib/log/golog"
	"code-lib/notify"
	"code-lib/notify/rabbitmq"
)

var (
	cfgfile *string
	log     logi.Logger
)

func init() {
	cfgfile = flag.String("cfg", "conf/handler.json", "Config file path.")
	flag.Parse()

	log = golog.NewDefault().Virtualize()
}

func main() {
	cfg := &HandlerConf{}
	err := ParseJSONConfigFile(*cfgfile, cfg)
	if err != nil {
		log.Errorf("parse json config file fail: %s", err)
		return
	}

	handler, err := CreateHandler(cfg)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	err = handler.Handle()
	if err != nil {
		log.Errorf("%s", err)
		return
	}
}

type HandlerConf struct {
	*rabbitmq.RabbitNotifyConf
}

type Handler struct {
	cfg *HandlerConf

	consumerMQ notify.Notify
}

func CreateHandler(cfg *HandlerConf) (handler Handler, err error) {
	handler = Handler{
		cfg: cfg,
	}

	handler.consumerMQ, err = rabbitmq.NewRabbitNotify(cfg.RabbitNotifyConf)
	if err != nil {
		return
	}

	err = handler.consumerMQ.Init()
	if err != nil {
		return
	}

	return
}

func (this *Handler) Handle() (err error) {
	err = this.consumerMQ.Receive()
	if err != nil {
		return
	}

	for data := range this.consumerMQ.Pop() {
		log.Printf("Handler get data: %s", string(data))
		err = this.consumerMQ.Ack()
		if err != nil {
			log.Errorf("%s", err)
			continue
		}
	}

	return
}

func ParseJSONConfigFile(file string, cfg interface{}) (err error) {
	fp, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(buffer, fp)
	if err != nil {
		return
	}

	return json.Unmarshal(buffer.Bytes(), cfg)
}
