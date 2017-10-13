package notify

type RabbitmqDestination struct {
	Exchange   string `json:"Exchange"`
	RoutingKey string `json:"RoutingKey"`
}
