package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nsq_to_consumer",
	Short: "nsq_to_consumer is a nsq consumer",
	Long:  "nsq_to_consumer is a tool, which consumes nsq messages. consume messages to elasticsearch or other targets.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, World!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
