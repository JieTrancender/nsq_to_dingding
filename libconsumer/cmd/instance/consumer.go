package instance

import (
	"fmt"
	"os"

	errw "github.com/pkg/errors"

	"github.com/JieTrancender/nsq_to_consumer/libconsumer/consumer"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/version"
)

// Consumer provides the runnable and configurable instance of a consumer.
type Consumer struct {
	consumer.Consumer
}

// NewConsumer creates a new consumer instance
func NewConsumer(name, v string) (*Consumer, error) {
	if v == "" {
		v = version.GetDefaultVersion()
	}

	c := consumer.Consumer{
		Info: consumer.Info{
			Consumer: name,
			Version:  v,
		},
	}

	return &Consumer{Consumer: c}, nil
}

// Run initializes and runs a Consumer imlementation. name is the name of Consumer (eg nsq_to_consumer).
// version is version number of the Consumer implementation.
// bt is the `Creator` callback for creating a new consumer instance.
func Run(settings Settings, creator consumer.Creator) error {
	err := setUmaskWithSettings(settings)
	if err != nil && err != errNotImplemented {
		return errw.Wrap(err, "could not set umask")
	}

	name := settings.Name
	version := settings.Version

	return handleError(func() error {
		defer func() {
			if r := recover(); r != nil {
				// logp.NewLogger(name).Fatalw("Failed due to panic.", "panic", r, zap.Stack("stack"))
				fmt.Printf("Failed due to panic.")
			}
		}()
		c, err := NewConsumer(name, version)
		if err != nil {
			return err
		}

		// Add basic info
		// todo

		return c.launch(settings, creator)
	}())
}

// createConsumer creates and returns the consumer, this method also initializes all needed items,
// including publisher
func (c *Consumer) createConEntity(creator consumer.Creator) (consumer.ConEntity, error) {
	fmt.Println("createConsumer...")

	// consumer, err := consumerCreator(&c.Consumer)
	// if err != nil {
	// 	return nil, err
	// }

	// conEntity := consumer.ConEntity{}
	conEntity, err := creator(&c.Consumer)
	if err != nil {
		return nil, err
	}

	return conEntity, nil
}

func (c *Consumer) launch(settings Settings, creator consumer.Creator) error {
	consumer, err := c.createConEntity(creator)
	if err != nil {
		return err
	}

	return consumer.Run(&c.Consumer)
}

// handleError handles the given error by logging it and then returning the error.
// If the err is nil or is a ErrGracefulExit error then the method will return nil without logging anything.
func handleError(err error) error {
	if err == nil || err == consumer.ErrGracefulExit {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Exiting: %v\n", err)
	return err
}

func setUmaskWithSettings(settings Settings) error {
	if settings.Umask != nil {
		return setUmask(*settings.Umask)
	}
	return setUmask(0027) // 0640 for files | 0750 for dirs
}
