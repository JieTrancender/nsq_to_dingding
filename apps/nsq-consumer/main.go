package main

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func processMessage(body []byte) error {
	fmt.Printf("consumer nsq msg:%s\n", string(body))

	return nil
}

type myMessageHandler struct{}

// HandleMessage implements the Handler interface
func (h *myMessageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}

	err := processMessage(m.Body)

	// returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return err
}

func main() {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("topic", "channel", config)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(&myMessageHandler{})

	err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		log.Fatal(err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	consumer.Stop()
}
