package cmd

import (
	"fmt"

	"github.com/diskfs/go-diskfs/sampleapp/os_img_modifier/pkg"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read a file from os an image file",
	Long: `read the content of a file which locate in a filesystem which locate ` +
		`in a partition of an os image file, and the max read size is 64k`,
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

		fileName := cmd.Flag("file").Value.String()
		if fileName == "" {
			return fmt.Errorf(`no file flag provided`)
		}

		f1Content, err := pkg.ReadFile(imageName, pNum, fileName)
		if err != nil {
			return err
		}

		fmt.Println(string(f1Content))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.PersistentFlags().String("image", "", "path name of the os image file")
	readCmd.PersistentFlags().Int("partitionNum", 1, "number of the partition")
	readCmd.PersistentFlags().String("file", "", "path name of the read file")
}
