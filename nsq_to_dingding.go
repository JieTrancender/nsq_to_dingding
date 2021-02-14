package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mreiferson/go-options"
	"github.com/nsqio/go-nsq"
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
const VERSION = "1.0.3"

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
	fs.Duration("lookupd-poll-interval", 6*time.Second, "lookupd poll interval")

	fs.Duration("dial-timeout", 6*time.Second, "dial nsqd timeout")
	fs.Duration("sync-interval", 30*time.Second, "sync file to dingding duration")
	fs.Int("publisher-num", 10, "number of concurrent publishers")

	fs.Duration("http-client-connect-timeout", 2*time.Second, "timeout for HTTP connect")
	fs.Duration("http-client-request-timeout", 5*time.Second, "timeout for HTTP request")

	fs.String("http-protocol", "https", "http protocol(default https)")
	fs.String("http-url", "oapi.dingtalk.com/robot/send", "http url(default oapi.dingtalk.com/robot/send)")

	fs.String("etcd-username", "", "etcd basic auth username")
	fs.String("etcd-password", "", "etcd basic auth password")
	fs.String("etcd-path", "/config/nsq_to_dingding/default", "etcd config path")

	consumerOpts := ArrayFlags{}
	etcdEndpoints := ArrayFlags{}

	fs.Var(&consumerOpts, "consumer-opt", "option to passthrough to nsq.Config (may be given multiple times, http://godoc.org/github.com/nsqio/go-nsq#Config)")
	fs.Var(&etcdEndpoints, "etcd-endpoint", "etcd endpoint(may be given multiple times)")

	return fs
}

func mainTmp() {
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
	cfg.DialTimeout = fs.Lookup("dial-timeout").Value.(flag.Getter).Get().(time.Duration)

	hupChan := make(chan os.Signal, 1)
	termChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	etcdEndpoints := fs.Lookup("etcd-endpoint").Value.(flag.Getter).Get().([]string)
	etcdUsername := fs.Lookup("etcd-username").Value.(flag.Getter).Get().(string)
	etcdPassword := fs.Lookup("etcd-password").Value.(flag.Getter).Get().(string)
	etcdPath := fs.Lookup("etcd-path").Value.(flag.Getter).Get().(string)

	if len(etcdEndpoints) == 0 {
		log.Fatal("error: not any etcd endpoint")
	}

	// fmt.Printf("full url: %s://%s?accessToken=%s\n", httpProtocol, httpURL, httpAccessToken)
	discoverer, err := newTopicDiscoverer(opts, cfg, hupChan, termChan,
		etcdEndpoints, etcdUsername, etcdPassword, etcdPath)
	if err != nil {
		log.Fatal("newTopicDiscoverer fail ", err)
	}
	err = discoverer.run()
	if err != nil {
		log.Fatal("run failed, err: ", err)
	}
}
