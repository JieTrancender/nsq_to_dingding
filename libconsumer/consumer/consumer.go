package consumer

// type Creator func(*Consumer, *common.Config) (Consumer, error)

// Creator todo
type Creator func(*Consumer) (ConEntity, error)

// ConEntity is the interface that must be implemented by every Consumer
type ConEntity interface {
	Run(c *Consumer) error

	Stop()
}

// Consumer contains the basic consumer data and the publisher client used to publish events.
type Consumer struct {
	Info Info // consumer metadata.
	// Publisher Pipeline
}
