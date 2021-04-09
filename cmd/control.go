package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/control"
)

// ControlCmd :: Steampipe control management
func ControlCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "control [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe control management",
		// TODO(nw) expand long description
		Long: `Steampipe control management.

Run Steampipe control packs and output in a variety of formats.`,
	}

	cmd.AddCommand(ControlRunCmd())

	return cmd
}

// ControlRunCmd :: handler for control run
func ControlRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Args:  cobra.ArbitraryArgs,
		Run:   runControlRunCmd,
		Short: "Steampipe control runner",
		Long: `Steampipe control runner.

Run a  as a local service, exposing it as a database endpoint for
connection from any Postgres compatible database client.`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddStringFlag(constants.ArgOutput, "", "", "The format(s) to output the report to. Must be one or more of csv, html or json in a comma-separated list", cmdconfig.FlagOptions.Required()).
		AddStringFlag(constants.ArgOutputDir, "", "", "The directory to output the control results to. Defaults to ./control-runs.").
		AddBoolFlag(constants.ArgNoColor, "", false, "Remove color from the terminal output.")

	return cmd
}

func runControlRunCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Running control...")

	output := cmdconfig.Viper().GetString("output")
	outputDir := cmdconfig.Viper().GetString("output-dir")
	noColor := cmdconfig.Viper().GetBool("no-color")

	reportingOptions := control.ControlReportingOptions{
		OutputFormats:   strings.Split(output, ","),
		OutputDirectory: outputDir,
		WithColor:       !noColor,
	}

	control.RunControl(reportingOptions)
}