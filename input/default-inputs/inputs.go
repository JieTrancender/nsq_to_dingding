package inputs

import (
	v2 "github.com/JieTrancender/nsq_to_consumer/input/v2"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/consumer"
)

// Init init plugins
func Init(info consumer.Info) []v2.Plugin {
	return make([]v2.Plugin, 0)
}
