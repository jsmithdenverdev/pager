package pubsub

type Handler interface {
	Handle(message Message) error
}

type HandlerFunc func(message Message) error

func (handler HandlerFunc) Handle(message Message) error {
	return handler(message)
}
