package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"

	"github.com/JieTrancender/nsq_to_consumer/libconsumer/cmd/instance"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/consumer"
)

func genRunCmd(settings instance.Settings, consumerCreator consumer.Creator) *cobra.Command {
	name := settings.Name
	runCmd := cobra.Command{
		Use:   "run",
		Short: "Run " + name,
		Run: func(cmd *cobra.Command, args []string) {
			err := instance.Run(settings, consumerCreator)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	// Run subcommand flags, only available to *consumer run
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("httpprof"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("cpuprofile"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("memprofile"))

	// if settings.RunFlags != nil {
	// 	runCmd.Flags().AddFlagSet(settings.RunFlags)
	// }

	return &runCmd
}
