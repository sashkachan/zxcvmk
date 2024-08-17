package main

import (
	"fmt"
	"log"
	"zxcvmk/cmd/backup"
	"zxcvmk/pkg/config"

	"github.com/spf13/cobra"
)

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
		backup.Print(cfg)
		fmt.Println(cfg.BackupProvider)
	},
}

var backupRestore = &cobra.Command{
	Use: "restore",
	Run: func(cmd *cobra.Command, args []string) {
		backup.Restore(cfg)
	},
}

func Execute() {
	if err != nil {
		log.Fatalf("Can't load config: %s", err)
	}
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupRestore)
	rootCmd.Execute()
}

func main() {
	Execute()
}
