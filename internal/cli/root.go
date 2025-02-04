// Package cli controls the main user entry point into both the API and interacting with it.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var appVersion = "0.0.dev_000000"

// The base entry point into the CLI
var RootCmd = &cobra.Command{
	Use:   "rc3",
	Short: "Easily spin up recurse internal Linux VMs/containers",
	Long: `Easily spin up recurse internal Linux VMs/containers.

Need a quick Linux VM or container? Our recurse internal cloud lets you spin up and manage compute effortlessly.

Great for quick presentations or internal recurse only applications.
`,
}

func init() {
	RootCmd.SetVersionTemplate(humanizeVersion(appVersion))
	RootCmd.AddCommand(cmdUp)
}

func Execute() error {
	return RootCmd.Execute()
}

func humanizeVersion(version string) string {
	semver, hash, err := strings.Cut(version, "_")
	if !err {
		return ""
	}
	return fmt.Sprintf("rc3 %s [%s]\n", semver, hash)
}
