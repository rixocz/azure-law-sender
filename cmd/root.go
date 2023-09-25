package cmd

import (
	"github.com/rixocz/azure-law-sender/version"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "azure-law-sender",
		Short: "Azure Log Analytics workspace data sender",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		Version: version.AppVersion,
	}

	command.AddCommand(NewSendCommand())

	return command
}
