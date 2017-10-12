package notify

import (
	"encoding/json"
	"time"
)

type TimerEntryProto struct {
	// 目的地
	Destination RabbitmqPublishProto `json:"destinetion"`

	// 消息发出时间
	SendUnixTime time.Time `json:"SendUnixTime"`

	// 超时时长
	// 超时时间　＝　${SendUnixTime} + Expire
	Expire time.Duration `json:"expire"`

	// 真实消息体
	Target *json.RawMessage `json:"Target"`
}
