package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JieTrancender/nsq_to_consumer/libconsumer/cmd/instance"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/consumer"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/logger"
)

// ConsumerRootCmd handles all application command line interface, parses user flags and runs subcommands
type ConsumerRootCmd struct {
	cobra.Command
	RunCmd     *cobra.Command
	VersionCmd *cobra.Command
}

// GenRootCmdWithSettings returns the root command to use for your consumer. It takes the run command,
// which will be called if no args are given
func GenRootCmdWithSettings(consumerCreator consumer.Creator, settings instance.Settings) *ConsumerRootCmd {
	name := settings.Name
	logger.NewLogger(name).Info("GenRootCmdWithSettings")
	fmt.Println("name GenRootCmdWithSettings")
	if settings.IndexPrefix == "" {
		settings.IndexPrefix = settings.Name
	}

	rootCmd := &ConsumerRootCmd{}
	rootCmd.Use = settings.Name

	// rootCmd.RunCmd = genRunCmd(settings, consumerCreator)
	rootCmd.VersionCmd = genVersionCmd(settings)

	// // Root command is an alias for run
	// rootCmd.Run = rootCmd.RunCmd.Run

	// // Persistent flags, common across all subcommands
	// rootCmd.PersistentFlags().AddGoFlag(flag.CommandLine.Lookup("path.home"))

	// // Register subcommands common to all consumers
	// rootCmd.AddCommand(rootCmd.RunCmd)
	rootCmd.AddCommand(rootCmd.VersionCmd)

	return rootCmd
}
