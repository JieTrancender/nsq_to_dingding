package main

import (
	"os"

	"github.com/JieTrancender/nsq_to_consumer/cmd"
	inputs "github.com/JieTrancender/nsq_to_consumer/input/default-inputs"
)

func main() {
	if err := cmd.NsqToConsumer(inputs.Init, cmd.NsqToConsumerSettings()).Execute(); err != nil {
		os.Exit(1)
	}
}
