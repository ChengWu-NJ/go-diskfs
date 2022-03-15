package cmd

import (
	"fmt"
	"log"

	godisk "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/sampleapp/os_img_modifier/pkg"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "modify os an image file",
	Long: `modify files that locate in a filesystem which locate ` +
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

		taskFile := cmd.Flag("taskFile").Value.String()
		if taskFile == "" {
			return fmt.Errorf(`no taskFile flag provided`)
		}

		disk, err := godisk.Open(`/tmp/dietpi.img`)
		if err != nil {
			return err
		}
		defer disk.File.Close()

		fs, err := disk.GetFilesystem(pNum)
		if err != nil {
			return err
		}

		task, err := pkg.CreateModiTaskFromYaml(taskFile)
		if err != nil {
			return err
		}

		if err := task.Modify(fs); err != nil {
			return err
		}

		log.Println(`all finished`)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modifyCmd)

	modifyCmd.PersistentFlags().String("image", "", "path name of the os image file")
	modifyCmd.PersistentFlags().Int("partitionNum", 1, "number of the partition")
	modifyCmd.PersistentFlags().String("taskFile", "", "path name of the task file")
}
