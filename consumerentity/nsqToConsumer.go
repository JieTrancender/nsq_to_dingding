package consumerentity

import (
	"fmt"

	v2 "github.com/JieTrancender/nsq_to_consumer/input/v2"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/consumer"
)

// NsqConsumer is a consumer object. Contains all objects needed to run the consumer
type NsqConsumer struct {
	pluginFactory PluginFactory
}

// PluginFactory ...
type PluginFactory func(consumer.Info) []v2.Plugin

// New creates a new nsqToConsumer pointer instance.
func New(plugins PluginFactory) consumer.Creator {
	return func(c *consumer.Consumer) (consumer.ConEntity, error) {
		return newConsumerEntity(c, plugins)
	}
}

func newConsumerEntity(c *consumer.Consumer, plugins PluginFactory) (consumer.ConEntity, error) {
	nsqConsumer := &NsqConsumer{}

	return nsqConsumer, nil
}

// Run allows the entity to be run as a consumer.
func (c *NsqConsumer) Run(consumer *consumer.Consumer) error {
	fmt.Println("nsqToConsumer run...")

	return nil
}

// Stop is called on exit.
func (c *NsqConsumer) Stop() {
	fmt.Println("nsqToConsumer stop...")
}
