package main

import (
	"fmt"
	"os"
	"zxcvmk/cmd/backup"
	k8svolumes "zxcvmk/cmd/k8s-volumes"
	"zxcvmk/pkg/config"

	"log/slog"

	"github.com/spf13/cobra"
)

func Execute() {
	backupArguments := backup.BackupArguments{}
	replantArguments := k8svolumes.K8sArguments{}
	var debugLevel bool
	// get config location from env
	config_location := os.Getenv("ZXCVMK_CONFIG")
	var defaultConfig = "config.yaml"
	if config_location == "" {
		config_location = defaultConfig
	}
	var cfg, err = config.LoadConfig(config_location)

	var rootCmd = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
			fmt.Println("This is a root command. It does nothing.")
		},
	}

	var backupCmd = &cobra.Command{
		Use: "backup",
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
			backup.List(cfg, backupArguments)
		},
	}

	var backupRestoreCmd = &cobra.Command{
		Use: "restore",
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
			backup.Restore(cfg, backupArguments)
		},
	}

	var backupListCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
			backup.List(cfg, backupArguments)
		},
	}

	var k8sCmd = &cobra.Command{
		Use: "k8s",
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
		},
	}

	var k8sVolumeReplantCmd = &cobra.Command{
		Use: "k8s-volume-replant",
		Run: func(cmd *cobra.Command, args []string) {
			SetupLogger(debugLevel)
			k8svolumes.Replant(cfg, replantArguments)
		},
	}

	if err != nil {
		slog.Error("Can't load config", "error", err)
		return
	}
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(k8sCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupListCmd)
	k8sCmd.AddCommand(k8sVolumeReplantCmd)

	rootCmd.PersistentFlags().BoolVar(&debugLevel, "debug", false, "Debug level")

	backupRestoreCmd.Flags().StringVar(&backupArguments.SnapshotID, "snapshot-id", "", "Specify the snapshot ID")
	backupRestoreCmd.Flags().StringArrayVar(&backupArguments.Paths, "filter-path", []string{}, "Specify the path filter (can be used multiple times)")
	backupRestoreCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")
	err = backupRestoreCmd.MarkFlagRequired("snapshot-ids")
	if err == nil {
		slog.Error("error setting up", "error", err)
		return
	}
	backupListCmd.Flags().StringArrayVar(&backupArguments.Paths, "filter-path", []string{}, "Specify the path filter (can be used multiple times)")
	backupListCmd.Flags().StringVar(&backupArguments.Output, "output", "", "Output type")

	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.PvcSrc, "pvc-src", "", "Specify the pvc source")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.PvcDst, "pvc-dst", "", "Specify the pvc target")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.Namespace, "namespace", "", "Specify the namespace of the pvc to replant")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.Deployment, "deployment", "", "Specify the deployment of the pvc to replant")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.DestVolumeSize, "dst-size", "", "Specify the destination pvc size")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.DestStorageClassName, "dst-storage-classname", "", "Specify the destination pvc classname")
	k8sVolumeReplantCmd.Flags().StringVar(&replantArguments.DeploymentVolumeName, "deployment-volume-name", "", "Specify the deployment volume name to replace")
	k8sVolumeReplantCmd.Flags().BoolVar(&replantArguments.DryRun, "dry-run", false, "dry-run")

	err = k8sVolumeReplantCmd.MarkFlagRequired("dst-size")
	if err != nil {
		slog.Error("dst-size is not provided")
		return
	}
	err = k8sVolumeReplantCmd.MarkFlagRequired("dst-storage-classname")
	if err != nil {
		slog.Error("dst-storage-classname is not provided")
		return
	}
	// err = k8sVolumeReplantCmd.MarkFlagRequired("deployment-volume-name")
	// if err != nil {
	// 	slog.Error("deployment-volume-name is not provided")
	// 	return
	// }
	_ = rootCmd.Execute()
}

func SetupLogger(debugLevel bool) {
	var handlerOptions slog.HandlerOptions
	var slogLevel slog.Level
	if debugLevel == true {
		slogLevel = slog.LevelDebug
	} else {
		slogLevel = slog.LevelInfo
	}
	handlerOptions = slog.HandlerOptions{
		Level: slogLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &handlerOptions))
	slog.SetDefault(logger)

	slog.Info("Debug level", "debugLevel", debugLevel)
}

func main() {
	Execute()
}
