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

	var backupRestoreCmd = &cobra.Command{
		Use: "restore",
		Run: func(cmd *cobra.Command, args []string) {
			backup.Restore(cfg, backupArguments)
		},
	}

	var backupListCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			backup.List(cfg, backupArguments)
		},
	}

	if err != nil {
		log.Fatalf("Can't load config: %s", err)
	}
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupRestoreCmd)

	backupCmd.AddCommand(backupListCmd)
	backupRestoreCmd.Flags().StringVar(&backupArguments.SnapshotID, "snapshot-id", "", "Specify the snapshot ID")
	backupRestoreCmd.Flags().StringVar(&backupArguments.Path, "filter-path", "", "Specify the path filter")
	backupRestoreCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")
	err = backupRestoreCmd.MarkFlagRequired("snapshot-ids")
	if err == nil {
		log.Fatal(err.Error())
	}
	backupListCmd.Flags().StringVar(&backupArguments.Path, "filter-path", "", "Specify the path filter")
	backupListCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")

	_ = rootCmd.Execute()
}

func main() {
	Execute()
}
