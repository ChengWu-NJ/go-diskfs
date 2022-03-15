package cmd

import (
	"fmt"

	"github.com/diskfs/go-diskfs/sampleapp/os_img_modifier/pkg"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list files or directories in a path",
	Long: `list files or directories in a path which locate in a filesystem which locate ` +
		`in a partition of an os image file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageName := cmd.Flag("image").Value.String()
		if imageName == "" {
			return fmt.Errorf(`no image flag provided`)
		}

		pNum, err := cmd.Flags().GetInt("partitionNum")
		if err != nil {
			return fmt.Errorf(`partitionNum flag err:[%v]`, err)
		}
		if pNum < 1 {
			return fmt.Errorf(`partitionNum should be greater than 0`)
		}

		pathName := cmd.Flag("path").Value.String()
		if pathName == "" {
			return fmt.Errorf(`no path flag provided`)
		}

		entries, err := pkg.LsDir(imageName, pNum, pathName)
		if err != nil {
			return err
		}

		for _, e := range entries {
			fmt.Printf("name: %s\tsize: %d\tisDir: %v\n", e.Name(), e.Size(), e.IsDir())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.PersistentFlags().String("image", "", "path name of the os image file")
	lsCmd.PersistentFlags().Int("partitionNum", 1, "number of the partition")
	lsCmd.PersistentFlags().String("path", "/", "full path name")
}
