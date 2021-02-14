package main

import (
	"log"
	"os"

	"github.com/nsqio/go-nsq"
)

// NSQConsumer nsq consumer structure
type NSQConsumer struct {
	publisher *DingDingPublisher
	opts      *Options
	topic     string
	consumer  *nsq.Consumer

	msgChan chan *nsq.Message

	termChan chan bool
	hupChan  chan bool
}

// NewNSQConsumer create NSQConsumer
func NewNSQConsumer(opts *Options, topic string, cfg *nsq.Config, config *NsqToDingDingConfig) (*NSQConsumer, error) {
	log.Println("NewNSQConsumer topic", topic)
	publisher, err := NewDingDingPublisher(config.Filter)
	if err != nil {
		return nil, err
	}

	consumer, err := nsq.NewConsumer(topic, opts.Channel, cfg)
	if err != nil {
		return nil, err
	}

	nsqConsumer := &NSQConsumer{
		publisher: publisher,
		opts:      opts,
		topic:     topic,
		consumer:  consumer,
		msgChan:   make(chan *nsq.Message, 1),
		termChan:  make(chan bool),
		hupChan:   make(chan bool),
	}
	consumer.AddHandler(nsqConsumer)

	err = consumer.ConnectToNSQDs(config.NsqdTCPAddresses)
	if err != nil {
		return nil, err
	}

	err = consumer.ConnectToNSQLookupds(config.LookupdHTTPAddresses)
	if err != nil {
		return nil, err
	}

	return nsqConsumer, nil
}

func (nsqConsumer *NSQConsumer) updateConfig(filter *MsgFilterConfig) {
	nsqConsumer.publisher.updateConfig(filter)
}

// HandleMessage implement of NSQ HandleMessage interface
func (nsqConsumer *NSQConsumer) HandleMessage(m *nsq.Message) error {
	m.DisableAutoResponse()
	nsqConsumer.msgChan <- m
	return nil
}

func (nsqConsumer *NSQConsumer) router() {
	closeDingDing, exit := false, false
	for {
		select {
		case <-nsqConsumer.consumer.StopChan:
			closeDingDing, exit = true, true
		case <-nsqConsumer.termChan:
			nsqConsumer.consumer.Stop()
		case <-nsqConsumer.hupChan:
			closeDingDing = true
		case m := <-nsqConsumer.msgChan:
			err := nsqConsumer.publisher.handleMessage(m)
			if err != nil {
				// 重试
				m.Requeue(-1)
				log.Println("NSQConsumer router msg deal fail", err)
				os.Exit(1)
			}
			m.Finish()
		}

		if closeDingDing {
			nsqConsumer.Close()
			closeDingDing = false
		}

		if exit {
			break
		}
	}
}

// Close close this NSQConsumer
func (nsqConsumer *NSQConsumer) Close() {
	log.Println("NSQConsumer Close")
}
