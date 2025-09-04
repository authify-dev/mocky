package eventbus

import "common/utils"

type DomainEvent interface {
	Marshal() utils.Result[[]byte]
	EventName() string
	AggregateID() string
	EventID() string
}
