package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "os_img_modifier",
		Short: "a tool to modify os image file",
		Long:  `a tool to modify os image file directly without mounting filesystem`,
	}

	ctx = context.Background()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
