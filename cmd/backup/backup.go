package backup

import (
	"fmt"
	"log"
	"zxcvmk/pkg/config"
	"zxcvmk/pkg/providers"
)

// identify backup targets
// set backup actions:
// - mount
// - restore
// - test (?)
//
//

type BackupArguments struct {
	SnapshotID string
	Path       string
	Output     string
}

func init() {

}

func setupBackupProvider(cfg *config.Config) providers.BackupProvider {
	var backupProviderImpl providers.BackupProvider
	var activeProvider config.BackupProvider
	for _, provider := range cfg.BackupProviders {
		if provider.Name == cfg.BackupProvider {
			activeProvider = provider
			switch provider.Name {
			case "restic":
				backupProviderImpl = providers.NewResticProvider(activeProvider.SnapshotListCommand, activeProvider.BackupRepositoryPasswordLocation, activeProvider.BackupRepository)
			}
			break
		}
	}
	return backupProviderImpl
}

func Restore(cfg *config.Config, backupArguments BackupArguments) {
	backupProviderImpl := setupBackupProvider(cfg)
	_, err := backupProviderImpl.ListSnapshots()
	if err != nil {
		fmt.Printf("Error listing snapshots: %s", err)
	}
}

func Mount(cfg *config.Config) {
	log.Println("Mount")
}

func List(cfg *config.Config, backupArguments BackupArguments) {
	backupProviderImpl := setupBackupProvider(cfg)
	snapshots, err := backupProviderImpl.ListSnapshots()
	if err != nil {
		fmt.Printf("Error listing snapshots: %s", err)
		return
	}
	var output string
	if backupArguments.Output != "" {
		output = backupArguments.Output
	} else {
		output = "json"
	}
	out, _ := config.Output(snapshots, output)
	fmt.Println(out)

}
