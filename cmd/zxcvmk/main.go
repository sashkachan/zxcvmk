package main

import (
	"fmt"
	"log"
	"os"
	"zxcvmk/cmd/backup"
	k8svolumes "zxcvmk/cmd/k8s-volumes"
	"zxcvmk/pkg/config"

	"github.com/spf13/cobra"
)

func Execute() {
	backupArguments := backup.BackupArguments{}
	replantArguments := k8svolumes.K8sArguments{}
	// get config location from env
	config_location := os.Getenv("ZXCVMK_CONFIG")
	var defaultConfig = "config.yaml"
	if config_location == "" {
		config_location = defaultConfig
	}
	var cfg, err = config.LoadConfig(config_location)

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

	var k8sCmd = &cobra.Command{
		Use: "k8s",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	var k8sVolumeReplantCmd = &cobra.Command{
		Use: "k8s-volume-replant",
		Run: func(cmd *cobra.Command, args []string) {
			k8svolumes.Replant(cfg, replantArguments)
		},
	}

	if err != nil {
		log.Fatalf("Can't load config: %s", err)
	}
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(k8sCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupListCmd)
	k8sCmd.AddCommand(k8sVolumeReplantCmd)

	backupRestoreCmd.Flags().StringVar(&backupArguments.SnapshotID, "snapshot-id", "", "Specify the snapshot ID")
	backupRestoreCmd.Flags().StringArrayVar(&backupArguments.Paths, "filter-path", []string{}, "Specify the path filter (can be used multiple times)")
	backupRestoreCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")
	err = backupRestoreCmd.MarkFlagRequired("snapshot-ids")
	if err == nil {
		log.Fatal(err.Error())
	}
	backupListCmd.Flags().StringArrayVar(&backupArguments.Paths, "filter-path", []string{}, "Specify the path filter (can be used multiple times)")
	backupListCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")


	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.Pvc, "pvc", "", "Specify the pvc to replant")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.Namespace, "namespace", "", "Specify the namespace of the pvc to replant")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.Deployment, "deployment", "", "Specify the deployment of the pvc to replant")
	k8sVolumeReplantCmd.Flags().BoolVar(&replantArguments.DryRun, "dry-run", false, "dry-run")

	_ = rootCmd.Execute()
}

func main() {
	Execute()
}
