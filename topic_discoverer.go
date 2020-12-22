package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/nsqio/go-nsq"
)

// MsgFilterConfig msg fileter config structure
type MsgFilterConfig struct {
	URL              string   `json:"url"`
	Protocol         string   `json:"protocol"`
	HTTPAccessTokens []string `json:"http-access-tokens"`
	FilterKeys       []string `json:"filterKeys"`
	IgnoreKeys       []string `json:"ignoreKeys"`
	NotAtKeys        []string `json:"notAtKeys"`
}

// NsqToDingDingConfig config structure
type NsqToDingDingConfig struct {
	LookupdHTTPAddresses []string         `json:"lookupd-http-addresses"`
	NsqdTCPAddresses     []string         `json:"nsqd-tcp-addresses"`
	Topics               []string         `json:"topics"`
	TopicRefreshInterval time.Duration    `json:"topic-refresh-interval"`
	Filter               *MsgFilterConfig `json:"filter"`
}

// TopicDiscoverer struct of topic discoverer
type TopicDiscoverer struct {
	opts          *Options
	topics        map[string]*NSQConsumer
	termChan      chan os.Signal
	hupChan       chan os.Signal
	logger        *log.Logger
	wg            sync.WaitGroup
	cfg           *nsq.Config
	etcdEndpoints []string
	etcdUsername  string
	etcdPassword  string
	etcdPath      string // etcd config path
	etcdCli       *clientv3.Client
	config        *NsqToDingDingConfig
	watcher       clientv3.Watcher
}

func newTopicDiscoverer(opts *Options, cfg *nsq.Config, hupChan chan os.Signal, termChan chan os.Signal,
	etcdEndpoints []string, etcdUsername, etcdPassword string) (*TopicDiscoverer, error) {
	discoverer := &TopicDiscoverer{
		opts:          opts,
		topics:        make(map[string]*NSQConsumer),
		termChan:      termChan,
		hupChan:       hupChan,
		logger:        log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags),
		cfg:           cfg,
		etcdEndpoints: etcdEndpoints,
		etcdUsername:  etcdUsername,
		etcdPassword:  etcdPassword,
		etcdPath:      "/config/nsq_to_dingding/default",
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

func (discoverer *TopicDiscoverer) updateTopics(topics []string) {
	for _, topic := range topics {
		if _, ok := discoverer.topics[topic]; ok {
			continue
		}

		nsqConsumer, err := NewNSQConsumer(discoverer.opts, topic, discoverer.cfg, discoverer.config)
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

func (discoverer *TopicDiscoverer) updateConifg() {
	for _, consumer := range discoverer.topics {
		consumer.updateConfig(discoverer.config.Filter)
	}
}

func newNsqToDingDingConfig() *NsqToDingDingConfig {
	config := &NsqToDingDingConfig{
		TopicRefreshInterval: 30,
		Filter: &MsgFilterConfig{
			Protocol: "https",
			URL:      "oapi.dingtalk.com/robot/send",
		},
	}

	return config
}

func (discoverer *TopicDiscoverer) watchConfig(watchStartVer int64) {
	discoverer.watcher = clientv3.NewWatcher(discoverer.etcdCli)
	watchChan := discoverer.watcher.Watch(context.Background(), discoverer.etcdPath, clientv3.WithRev(watchStartVer))
	for resp := range watchChan {
		for _, ev := range resp.Events {
			if ev.Type == clientv3.EventTypePut {
				fmt.Printf("watchConfig %s, %s %s", ev.Type, string(ev.Kv.Key), string(ev.Kv.Value))
				config := newNsqToDingDingConfig()
				err := json.Unmarshal(ev.Kv.Value, config)
				if err != nil {
					fmt.Println("配置格式错误", string(ev.Kv.Value))
					break
				}

				// todo: 检查配置格式
				discoverer.config = config

				// 所有消费者更新
				discoverer.updateConifg()

				// 更新配置信息
				fmt.Println("更新配置信息", discoverer.config)

				break
			}
		}
	}
}

// initAndWatchConfig get and watch etcd config
func (discoverer *TopicDiscoverer) initAndWatchConfig() error {
	kv := clientv3.NewKV(discoverer.etcdCli)
	resp, err := kv.Get(context.Background(), discoverer.etcdPath)
	if err != nil {
		return err
	}

	isConfigFound := false
	var config = newNsqToDingDingConfig()
	for _, ev := range resp.Kvs {
		fmt.Printf("range %s %s\n", string(ev.Key), string(discoverer.etcdPath))
		if string(ev.Key) == discoverer.etcdPath {
			// todo: schema check
			err := json.Unmarshal(ev.Value, config)
			if err != nil {
				return fmt.Errorf("配置格式错误:%s", string(ev.Value))
			}

			isConfigFound = true
			break
		}
	}

	if !isConfigFound {
		return fmt.Errorf("Config is not exist in %s", discoverer.etcdPath)
	}

	if len(config.LookupdHTTPAddresses) == 0 && len(config.NsqdTCPAddresses) == 0 {
		return fmt.Errorf("Config is invalid, lookupd-http-address or nsqd-tcp-address is required")
	}

	if len(config.LookupdHTTPAddresses) != 0 && len(config.NsqdTCPAddresses) != 0 {
		return fmt.Errorf("Config is invalid, use lookupd-http-address or nsqd-tcp-address, not both")
	}

	if len(config.Topics) == 0 {
		return fmt.Errorf("Config is invalid, topic is required")
	}

	discoverer.config = config

	fmt.Println("init config", config)
	discoverer.wg.Add(1)
	watchStartVer := resp.Header.Revision + 1
	go discoverer.watchConfig(watchStartVer)

	return nil
}

func (discoverer *TopicDiscoverer) run() error {
	err := discoverer.initAndWatchConfig()
	if err != nil {
		return err
	}

	ticker := time.Tick(discoverer.config.TopicRefreshInterval * time.Second)
	discoverer.updateTopics(discoverer.config.Topics)

forloop:
	for {
		select {
		case <-ticker:
			discoverer.updateTopics(discoverer.config.Topics)
		case <-discoverer.termChan:
			discoverer.watcher.Close()
			discoverer.wg.Done()

			discoverer.etcdCli.Close()

			for _, nsqConsumer := range discoverer.topics {
				close(nsqConsumer.termChan)
			}
			break forloop
		case <-discoverer.hupChan:
			discoverer.watcher.Close()
			discoverer.wg.Done()

			discoverer.etcdCli.Close()

			for _, nsqConsumer := range discoverer.topics {
				nsqConsumer.hupChan <- true
			}
			break forloop
		}
	}

	discoverer.wg.Wait()

	return nil
}
