package cli

import "github.com/spf13/cobra"

var cmdUp = &cobra.Command{
	Use:     "up <path>",
	Short:   "Create or update new VM or container",
	Example: `$ rc3 up ./my_container.toml`,
	RunE:    upsertInstance,
}

func upsertInstance(cmd *cobra.Command, args []string) error {
	return nil
}
