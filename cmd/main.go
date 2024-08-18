package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"zxcvmk/cmd/backup"
	"zxcvmk/pkg/config"
)

func Execute() {
	backupArguments := backup.BackupArguments{}

	var defaultConfig = "config.yaml"
	var cfg, err = config.LoadConfig(defaultConfig)

	var rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("This is a root command. It does nothing.")
		},
	}

	var backupCmd = &cobra.Command{
		Use: "backup",
		Run: func(cmd *cobra.Command, args []string) {
			backup.List(cfg, backupArguments)
		},
	}

	var backupRestore = &cobra.Command{
		Use: "restore",
		Run: func(cmd *cobra.Command, args []string) {
			backup.Restore(cfg, backupArguments)
		},
	}

	var backupList = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			backup.List(cfg, backupArguments)
		},
	}

	if err != nil {
		log.Fatalf("Can't load config: %s", err)
	}
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupRestore)
	backupCmd.AddCommand(backupList)
	backupRestore.Flags().StringVar(&backupArguments.SnapshotID, "snapshot-id", "", "Specify the snapshot ID")
	backupRestore.Flags().StringVar(&backupArguments.Path, "filter-path", "", "Specify the path filter")
	backupRestore.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")
	backupList.Flags().StringVar(&backupArguments.SnapshotID, "snapshot-id", "", "Specify the snapshot ID")
	backupList.Flags().StringVar(&backupArguments.Path, "filter-path", "", "Specify the path filter")
	backupList.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")

	rootCmd.Execute()
}

func main() {
	Execute()
}
