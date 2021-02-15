package cmd

import (
	"github.com/JieTrancender/nsq_to_consumer/consumerentity"
	cmd "github.com/JieTrancender/nsq_to_consumer/libconsumer/cmd"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/cmd/instance"
)

// Name of this consumer
const Name = "nsq_to_consumer"

// RootCmd to handle consumer cli
var RootCmd *cmd.ConsumerRootCmd

// NsqToConsumerSettings constains the default settings for consumer
func NsqToConsumerSettings() instance.Settings {
	return instance.Settings{
		Name: Name,
	}
}

// NsqToConsumer build the consumer root command for executing nsq_to_consumer and it's subcommands.
func NsqToConsumer(inputs consumerentity.PluginFactory, settings instance.Settings) *cmd.ConsumerRootCmd {
	command := cmd.GenRootCmdWithSettings(consumerentity.New(inputs), settings)
	return command
}

// var rootCmd = &cobra.Command{
// 	Use:   "nsq_to_consumer",
// 	Short: "nsq_to_consumer is a nsq consumer",
// 	Long:  "nsq_to_consumer is a tool, which consumes nsq messages. consume messages to elasticsearch or other targets.",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("Hello, World!")
// 	},
// }

// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(1)
// 	}
// }
