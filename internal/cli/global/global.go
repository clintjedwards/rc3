package global

import (
	"log"

	"github.com/clintjedwards/polyfmt"
	"github.com/clintjedwards/rc3/internal/conf"
	"github.com/spf13/cobra"
)

// Structure for values that all commands need.
type Context struct {
	Fmt    polyfmt.Formatter
	Config *conf.CLI
}

// Static global for the lifetime of the command
var CLIContext *Context

func InitState(cmd *cobra.Command) {
	// Including these in the pre run hook instead of in the enclosing/parent command definition
	// allows cobra to still print errors and usage for its own cli verifications, but
	// ignore our errors.
	cmd.SilenceUsage = true  // Don't print the usage if we get an upstream error
	cmd.SilenceErrors = true // Let us handle error printing ourselves

	// Now we need to provide the command line with some state which we use to display the spinner
	// and also make sure the command line inherits the proper variable chain(config file -> envvar -> flags)
	CLIContext = &Context{}

	// Initiate the CLI config but we don't use the config path feature so just leave it empty.
	CLIContext.NewConfig("")

	// Initiate the formatter(this controls the command line output)
	format, _ := cmd.Flags().GetString("format")
	if format != "" {
		CLIContext.Config.Format = format
	}

	CLIContext.NewFormatter()
}

func (c *Context) NewFormatter() {
	clifmt, err := polyfmt.NewFormatter(polyfmt.Mode(c.Config.Format), polyfmt.DefaultOptions())
	if err != nil {
		log.Fatal(err)
	}

	c.Fmt = clifmt
}

func (c *Context) NewConfig(configPath string) {
	config, err := conf.InitCLIConfig(configPath, true)
	if err != nil {
		log.Fatal(err)
	}

	c.Config = config
}
