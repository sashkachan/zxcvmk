package backup

import (
	"fmt"
	"log"
	"os"
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
	snapshots, err := backupProviderImpl.ListSnapshots()
	if err != nil {
		fmt.Printf("Error listing snapshots: %s", err)
	}

	if snapshot, found := findSnapshotByID(snapshots, backupArguments.SnapshotID); found {
		if backupArguments.Path != "" {
			for _, path := range snapshot.Paths {
				if backupArguments.Path == path {
					// restore this path
				}
			}
		}
	} else {
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

func mountSnapshotID(id string, paths []string) error {
	target_tmpdir, err := createSnapshotMountTarget(id)
	if err != nil {
		return err
	}

	return nil
}
func restoreSnapshotID(id string, paths []string) error {

	return nil
}

func findSnapshotByID(snapshots []*providers.Snapshot, id string) (*providers.Snapshot, bool) {
	for i := range snapshots {
		if snapshots[i].ID == id {
			return snapshots[i], true
		}
	}
	return nil, false
}

func createSnapshotMountTarget(snapshotID string) (string, error) {
	tmpdir := os.TempDir()
	target_tmpdir, err := os.MkdirTemp(tmpdir, "snapshot-")
	if err != nil {
		return "", err
	}
	return target_tmpdir, nil
}
