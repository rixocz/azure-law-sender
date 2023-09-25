package cmd

import (
	"github.com/rixocz/azure-law-sender/azure"
	"github.com/rixocz/azure-law-sender/version"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

var workspaceId = ""
var table = ""
var subscriptionId = ""

func NewSendCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "send [<data>]",
		Short: "Send data to Azure Log Analytics workspace.",
		Long: `If data are provided it should be JSON format, 
instead it will be wrapped into JSON. {"data" : "<data>"}`,
		Version: version.AppVersion,
		RunE: func(cmd *cobra.Command, args []string) error {
			var dataReader io.Reader
			if len(args) > 0 {
				dataReader = strings.NewReader(args[0])
			} else {
				dataReader = strings.NewReader(`{"name":"tester"}`)
			}
			collector, err := azure.NewCollector(subscriptionId, workspaceId, table)
			if err != nil {
				return err
			}
			return collector.SendData(dataReader)
		},
	}

	prepareFlags(command)

	return command
}

func prepareFlags(command *cobra.Command) {
	command.Flags().StringVarP(&workspaceId, "workspace-id", "w", "", "Azure Log Analytics Workspace ID as UUID")
	_ = command.MarkFlagRequired("workspace-id")

	command.Flags().StringVarP(&table, "table", "t", "", "Azure Log Analytics Workspace table name")
	_ = command.MarkFlagRequired("table")

	command.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Azure Subscription ID as UUID")
	_ = command.MarkFlagRequired("subscription-id")
}
