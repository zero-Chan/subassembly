package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	logi "code-lib/log"
	"code-lib/log/golog"

	conf "subassembly/timer-controller/conf/logic"
	"subassembly/timer-controller/logic"
	proto "subassembly/timer-controller/proto/notify"
)

var (
	cfgfile *string
	log     logi.Logger
)

func init() {
	cfgfile = flag.String("cfg", "conf/mylogic.json", "Config file path.")
	flag.Parse()

	log = golog.NewDefault().Virtualize()
}

func main() {
	cfg := &LogicaConf{}
	err := ParseJSONConfigFile(*cfgfile, cfg)
	if err != nil {
		log.Errorf("parse json config file fail: %s", err)
		return
	}

	logica, err := CreateLogica(cfg)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	logica.Set("txt", "Hello")
	data, err := logica.Marshal()
	if err != nil {
		log.Errorf("logica marshal fail: %s", err)
		return
	}

	err = logica.Push2Timer(data, time.Second*5, proto.RabbitmqDestination{
		Exchange:   "cza.test.handler",
		RoutingKey: "firsthandler",
	})
	if err != nil {
		log.Errorf("%s", err)
		return
	}
}

type LogicaConf struct {
	*conf.BaseLogicConf
}

type Logica struct {
	*logic.BaseLogic `json:"-"`

	cfg *LogicaConf `json:"-"`
	Key string      `json:"Key"`
	Val string      `json:"Val"`
}

func CreateLogica(cfg *LogicaConf) (logica Logica, err error) {
	if cfg == nil {
		err = fmt.Errorf("[InitError] `Func: CreateLogica` `Reason: cfg is nil.`")
	}

	logica = Logica{
		cfg: cfg,
	}

	logica.BaseLogic, err = logic.NewBaseLogic(cfg.BaseLogicConf)
	if err != nil {
		return
	}

	return
}

func (this *Logica) Set(key, val string) {
	this.Key = key
	this.Val = val
}

func (this *Logica) Marshal() (data []byte, err error) {
	return json.Marshal(this)
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
