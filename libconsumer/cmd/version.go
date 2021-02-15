package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/JieTrancender/nsq_to_consumer/libconsumer/cmd/instance"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/common/cli"
	"github.com/JieTrancender/nsq_to_consumer/libconsumer/version"
)

// genVersionCmd generates the command version for a consumer.
func genVersionCmd(settings instance.Settings) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current version info",
		Run: cli.RunWith(
			func(_ *cobra.Command, args []string) error {
				buildTime := "unknown"
				if bt := version.BuildTime(); !bt.IsZero() {
					buildTime = bt.String()
				}

				fmt.Printf("%s version %s (%s), consumer %s [%s built %s]\n",
					"", "", runtime.GOARCH, version.GetDefaultVersion(), version.Commit(), buildTime)
				return nil
			}),
	}
}
