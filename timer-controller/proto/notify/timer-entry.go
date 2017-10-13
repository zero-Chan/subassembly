package notify

import (
	"encoding/json"
	"time"
)

type TimerNotice struct {
	// 目的地
	Destination RabbitmqDestination `json:"Destinetion"`

	// 消息发出时间
	SendTime time.Time `json:"SendUnixTime"`

	// 超时时长
	// 超时时间　＝　${SendUnixTime} + Expire
	Expire time.Duration `json:"Expire"`

	// 真实消息体
	Target json.RawMessage `json:"Target"`
}
