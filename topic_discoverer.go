package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/nsqio/go-nsq"
)

// FilterConfigStruct filter config structure
// type FilterConfigStruct struct {
// 	FilterKeys []string `json:"filterKeys"`
// }

// TopicDiscoverer struct of topic discoverer
type TopicDiscoverer struct {
	opts          *Options
	topics        map[string]*NSQConsumer
	termChan      chan os.Signal
	hupChan       chan os.Signal
	logger        *log.Logger
	wg            sync.WaitGroup
	cfg           *nsq.Config
	protocol      string
	url           string
	accessToken   []string
	etcdEndpoints []string
	etcdUsername  string
	etcdPassword  string
	etcdPath      string // etcd config path
	etcdCli       *clientv3.Client
}

func newTopicDiscoverer(opts *Options, cfg *nsq.Config, hupChan chan os.Signal, termChan chan os.Signal,
	protocol, url string, accessToken []string,
	etcdEndpoints []string, etcdUsername, etcdPassword string) (*TopicDiscoverer, error) {
	discoverer := &TopicDiscoverer{
		opts:          opts,
		topics:        make(map[string]*NSQConsumer),
		termChan:      termChan,
		hupChan:       hupChan,
		logger:        log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags),
		cfg:           cfg,
		protocol:      protocol,
		url:           url,
		accessToken:   accessToken,
		etcdEndpoints: etcdEndpoints,
		etcdUsername:  etcdUsername,
		etcdPassword:  etcdPassword,
		etcdPath:      "/config/nsq_to_dingding/",
	}

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	discoverer.etcdCli = etcdCli

	return discoverer, nil
}

func (discoverer *TopicDiscoverer) isTopicAllowed(topic string) bool {
	if len(discoverer.opts.TopicPatterns) == 0 {
		return true
	}

	var match bool
	var err error
	for _, pattern := range discoverer.opts.TopicPatterns {
		match, err = regexp.MatchString(pattern, topic)
		if err == nil {
			break
		}
	}

	return match
}

func (discoverer *TopicDiscoverer) updateTopics(topics []string) {
	for _, topic := range topics {
		if _, ok := discoverer.topics[topic]; ok {
			continue
		}

		if !discoverer.isTopicAllowed(topic) {
			discoverer.logger.Printf("skipping topic %s (doesn't match any pattern)\n", topic)
			continue
		}

		nsqConsumer, err := NewNSQConsumer(discoverer.opts, topic, discoverer.cfg,
			discoverer.protocol, discoverer.url, discoverer.accessToken)
		if err != nil {
			discoverer.logger.Printf("error: could not register topic %s: %s", topic, err)
			continue
		}
		discoverer.topics[topic] = nsqConsumer

		discoverer.wg.Add(1)
		go func(nsqConsumer *NSQConsumer) {
			nsqConsumer.router()
			discoverer.wg.Done()
		}(nsqConsumer)
	}
}

// initAndWatchConfig get and watch etcd config
func (discoverer *TopicDiscoverer) initAndWatchConfig() error {
	resp, err := discoverer.etcdCli.Get(context.Background(), discoverer.etcdPath, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// err := json.Unmarshal(resp.Node.Value)

	fmt.Println(resp)

	return errors.New("fsdf")

	// cli, err := clientv3.New(clientv3.Config{
	// 	Endpoints: discoverer.Endpoints
	// })
}

func (discoverer *TopicDiscoverer) run() error {
	err := discoverer.initAndWatchConfig()
	if err != nil {
		return err
	}

	var ticker <-chan time.Time
	if len(discoverer.opts.Topics) == 0 {
		ticker = time.Tick(discoverer.opts.TopicRefreshInterval)
	}
	discoverer.updateTopics(discoverer.opts.Topics)

forloop:
	for {
		select {
		case <-ticker:
			discoverer.updateTopics(discoverer.opts.Topics)
		case <-discoverer.termChan:
			discoverer.etcdCli.Close()

			for _, nsqConsumer := range discoverer.topics {
				// nsqConsumer.consumer.Stop()
				close(nsqConsumer.termChan)
			}
			break forloop
		case <-discoverer.hupChan:
			for _, nsqConsumer := range discoverer.topics {
				nsqConsumer.hupChan <- true
			}
			break forloop
		}
	}

	discoverer.wg.Wait()

	return nil
}
