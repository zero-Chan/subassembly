package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"

	logi "code-lib/log"
	"code-lib/log/golog"

	conf "subassembly/timer-controller/conf/timer"
	"subassembly/timer-controller/timer"
)

var (
	cfgfile *string
	log     logi.Logger
)

func init() {
	cfgfile = flag.String("cfg", "conf/timer.json", "Config file path.")
	flag.Parse()

	log = golog.NewDefault().Virtualize()
}

func main() {
	cfg := &conf.TimerConf{}
	err := ParseJSONConfigFile(*cfgfile, cfg)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	mytimer, err := timer.NewTimer(cfg)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	err = mytimer.Run()
	if err != nil {
		log.Errorf("%s", err)
		return
	}
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
