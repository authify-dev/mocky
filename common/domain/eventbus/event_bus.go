package eventbus

import (
	"common/utils"
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type EventBus interface {
	Publish(ctx context.Context, events []DomainEvent) error
	Consume(queue, key string) utils.Result[<-chan *message.Message]
}

type SettingsEventBus struct {
	Username string
	Password string
	Protocol string
	Host     string
	Port     string
	Exchange string
}
