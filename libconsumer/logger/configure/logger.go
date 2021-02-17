package configure

import (
	"flag"
	"strings"

	"github.com/JieTrancender/nsq_to_consumer/libconsumer/common"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/logger"
)

var (
	verbose        bool
	toStderr       bool
	debugSelectors []string
	environment    logger.Environment
)

type enviromentVar logger.Environment

func init() {
	flag.BoolVar(&verbose, "v", false, "Log at INFO level")
	flag.BoolVar(&toStderr, "e", false, "Log to stderr and disable syslog/file output")
	common.StringArrVarFlag(nil, &debugSelectors, "d", "Enable certain debug selectors")
	// flag.Var((*enviromentVar)(&environment), "enviroment", "set environment being ran in")
}

// Logging builds a logger.Config based on the given common.Config and the specified CLI flags.
func Logging(consumerName string) error {
	config := logger.DefaultConfig(logger.DefaultEnvironment)
	config.Consumer = consumerName
	// logger.Info("Logging", consumerName)

	applyFlags(&config)
	return logger.Configure(config)
}

func applyFlags(cfg *logger.Config) {
	if toStderr {
		cfg.IsStderr = true
	}

	if cfg.Level > logger.InfoLevel && verbose {
		cfg.Level = logger.InfoLevel
	}

	for _, selectors := range debugSelectors {
		cfg.Selectors = append(cfg.Selectors, strings.Split(selectors, ", ")...)
	}

	// Elevate level if selectors are specified on the CLI.
	if len(debugSelectors) > 0 {
		cfg.Level = logger.DebugLevel
	}
}
