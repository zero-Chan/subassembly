package notify

import (
	"github.com/streadway/amqp"
)

type RabbitmqPublishProto struct {
	exchange   string
	routingKey string

	// 如果 mandatory　为 true，
	// 当 消息推送给一个没有绑定队列的routerKey　时
	// 消息会被丢弃
	// 默认值为false
	mandatory bool

	// 如果 immediate 为 true.
	// 当　消息推送给一个没有消费者的队列　时
	// 消息会被丢弃
	// 默认值为false
	immediate bool

	// 消息体
	msg amqp.Publishing
}

func CreateRabbitmqPublishProto(exchange string, routingKey string, body []byte) RabbitmqPublishProto {
	proto := RabbitmqPublishProto{
		exchange:   exchange,
		routingKey: routingKey,
		mandatory:  false,
		immediate:  false,
		msg: amqp.Publishing{
			Body: body,
		},
	}

	return proto
}

func NewRabbitmqPublishProto(exchange string, routingKey string, body []byte) *RabbitmqPublishProto {
	proto := CreateRabbitmqPublishProto(exchange, routingKey, body)
	return &proto
}

func (this *RabbitmqPublishProto) SetMandatory(b bool) {
	this.mandatory = b
}

func (this *RabbitmqPublishProto) SetImmediate(b bool) {
	this.immediate = b
}

func (this *RabbitmqPublishProto) SetMsg(msg amqp.Publishing) {
	this.msg = msg
}
