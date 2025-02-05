// Package cli controls the main user entry point into both the API and interacting with it.
package cli

import (
	"fmt"
	"strings"

	"github.com/clintjedwards/rc3/internal/cli/global"
	"github.com/clintjedwards/rc3/internal/cli/service"
	"github.com/clintjedwards/rc3/internal/conf"
	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev"

// The base entry point into the CLI
var RootCmd = &cobra.Command{
	Use:   "rc3",
	Short: "Easily spin up recurse internal Linux VMs/containers",
	Long: `Easily spin up recurse internal Linux VMs/containers.

Need a quick Linux VM or container? Our recurse internal cloud lets you spin up and manage compute effortlessly.

Great for quick presentations or internal recurse only applications.

### Environment Variables supported:
` + strings.Join(conf.GetCLIEnvVars(), "\n"),
	Version: " ", // We leave this added but empty so that the rootcmd will supply the -v flag
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		global.InitState(cmd)
	},
}

func init() {
	RootCmd.SetVersionTemplate(humanizeVersion(appVersion))
	RootCmd.AddCommand(cmdUp)
	RootCmd.AddCommand(service.CmdService)
}

func Execute() error {
	return RootCmd.Execute()
}

func humanizeVersion(version string) string {
	return fmt.Sprintf("rc3 v%s\n", version)
}
