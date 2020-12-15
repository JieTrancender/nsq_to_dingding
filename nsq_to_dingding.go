package main

import (
	"flag"
	"fmt"
	"github.com/mreiferson/go-options"
	"github.com/nsqio/go-nsq"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ArrayFlags for multi flag value
type ArrayFlags []string

// Set implement for Flag.Set
func (arrayFlags *ArrayFlags) Set(value string) error {
	*arrayFlags = append(*arrayFlags, value)

	return nil
}

// String for Flag.String
func (arrayFlags *ArrayFlags) String() string {
	return fmt.Sprint(*arrayFlags)
}

// Get for Flag.Get
func (arrayFlags *ArrayFlags) Get() interface{} {
	return []string(*arrayFlags)
}

// VERSION verions of nsqToDingDing
const VERSION = "0.0.1"

func flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("nsqToDingDing", flag.ExitOnError)

	fs.Bool("version", false, "show version")
	fs.String("log-level", "info", "set log verbosity: debug, info, warn, error, or fatal")
	fs.String("log-prefix", "[nsqToDingDing]", "log message prefix")

	fs.String("channel", "nsqToDingDing", "nsq channel")
	fs.Int("max-in-flight", 200, "max number of messages to allow in flight")

	fs.String("output-dir", "/tmp", "directory to write output files to")
	fs.String("work-dir", "", "directory for in-progress files before moving to output-dir")
	fs.Duration("topic-refresh", time.Minute, "how frequently the topic list should be refreshed")

	fs.Duration("sync-interval", 30*time.Second, "sync file to dingding duration")
	fs.Int("publisher-num", 10, "number of concurrent publishers")

	fs.Duration("http-client-connect-timeout", 2*time.Second, "timeout for HTTP connect")
	fs.Duration("http-client-request-timeout", 5*time.Second, "timeout for HTTP request")

	fs.String("http-protocol", "https", "http protocol(default https)")
	fs.String("http-url", "oapi.dingtalk.com/robot/send", "http url(default oapi.dingtalk.com/robot/send)")
	fs.String("http-access-token", "", "ding ding access token")

	nsqdTCPAddrs := ArrayFlags{}
	lookupdHTTPAddrs := ArrayFlags{}
	topics := ArrayFlags{}
	topicPatterns := ArrayFlags{}
	consumerOpts := ArrayFlags{}
	fs.Var(&nsqdTCPAddrs, "nsqd-tcp-address", "nsqd TCP address (may be given multiple times)")
	fs.Var(&lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
	fs.Var(&topics, "topic", "nsq topic (may be given multiple times)")
	fs.Var(&topicPatterns, "topic-pattern", "nsq topic pattern (may be given multiple times)")
	fs.Var(&consumerOpts, "consumer-opt", "option to passthrough to nsq.Config (may be given multiple times, http://godoc.org/github.com/nsqio/go-nsq#Config)")

	return fs
}

func main() {
	fs := flagSet()
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("parse fail:%v", err)
	}

	if args := fs.Args(); len(args) > 0 {
		log.Fatalf("unknown arguments: %s", args)
	}

	opts := NewOptions()
	options.Resolve(opts, fs, nil)

	// logger := log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags)

	if fs.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Printf("nsq_to_dingding@v%s go-nsq@v%s\n", VERSION, nsq.VERSION)
	}

	if opts.Channel == "" {
		log.Fatal("--channel is required")
	}

	if opts.HTTPClientConnectTimeout <= 0 {
		log.Fatal("--http-client-connect-timeout should be positive")
	}

	if opts.HTTPClientRequestTimeout <= 0 {
		log.Fatal("--http-client-request-timeout should be positive")
	}

	if len(opts.NSQDTCPAddrs) == 0 && len(opts.NSQLookupdHTTPAddrs) == 0 {
		log.Fatal("--nsqd-tcp-address or --lookupd-http-address required")
	}

	if len(opts.NSQDTCPAddrs) != 0 && len(opts.NSQLookupdHTTPAddrs) != 0 {
		log.Fatal("use --nsqd-tcp-address or --lookupd-http-address not both")
	}

	if len(opts.Topics) == 0 && len(opts.TopicPatterns) == 0 {
		log.Fatal("--topic or --topic-pattern required")
	}

	if len(opts.Topics) != 0 && len(opts.NSQLookupdHTTPAddrs) == 0 {
		log.Fatal("--lookupd-http-address must be specified when no --topic specified")
	}

	if opts.WorkDir == "" {
		opts.WorkDir = opts.OutputDir
	}

	cfg := nsq.NewConfig()
	cfgFlag := nsq.ConfigFlag{Config: cfg}
	for _, opt := range opts.ConsumerOpts {
		_ = cfgFlag.Set(opt)
	}
	cfg.UserAgent = fmt.Sprintf("nsq_to_dingding/%s go-nsq/%s", VERSION, nsq.VERSION)
	cfg.MaxInFlight = opts.MaxInFlight

	hupChan := make(chan os.Signal, 1)
	termChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	httpProtocol := fs.Lookup("http-protocol").Value.(flag.Getter).Get().(string)
	httpURL := fs.Lookup("http-url").Value.(flag.Getter).Get().(string)
	httpAccessToken := fs.Lookup("http-access-token").Value.(flag.Getter).Get().(string)

	if httpAccessToken == "" {
		fmt.Printf("warn: http access token is empty")
	}

	fmt.Printf("full url: %s://%s?accessToken=%s\n", httpProtocol, httpURL, httpAccessToken)
	discoverer, _ := newTopicDiscoverer(opts, cfg, hupChan, termChan,
		httpProtocol, httpURL, httpAccessToken)
	discoverer.run()
}
