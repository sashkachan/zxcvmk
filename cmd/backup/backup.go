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

func init() {

}

func Restore(cfg *config.Config) {
	log.Println("Restore")
	log.Printf("Current provider: %s", cfg.BackupProvider)
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
	snapshots, err := backupProviderImpl.ListSnapshots()
	if err != nil {
		fmt.Printf("Error listing snapshots: %s", err)
	}
	fmt.Printf("Snapshots: %d", len(snapshots))
}

func Mount(cfg *config.Config) {
	log.Println("Mount")
}

func Print(cfg *config.Config) {
	fmt.Println("Current backup targets:")
	for _, p := range cfg.BackupTargets {
		fmt.Printf("%+v\n", p.Location)
	}
}
