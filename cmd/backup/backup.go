package backup

import (
	"fmt"
	"log"
	"os"
	"zxcvmk/pkg/config"
	"zxcvmk/pkg/providers"
)

type BackupArguments struct {
	SnapshotID string
	Paths      []string
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
				backupProviderImpl = providers.NewResticProvider(activeProvider.BackupRepositoryPasswordLocation, activeProvider.BackupRepository)
			}
			break
		}
	}
	return backupProviderImpl
}

func Restore(cfg *config.Config, backupArguments BackupArguments) {
	backupProviderImpl := setupBackupProvider(cfg)
	snapshots, err := backupProviderImpl.ListSnapshots(backupArguments.Paths)
	if err != nil {
		fmt.Printf("Error listing snapshots: %s", err)
	}
	if snapshot, found := findSnapshotByID(snapshots, backupArguments.SnapshotID); found {
		target, err := createSnapshotMountTarget()
		if err != nil {
			log.Fatal("Snapshot target directory could not be created", err)
		}
		err = backupProviderImpl.RestoreSnapshot(snapshot.ID, target, backupArguments.Paths)
		if err != nil {
			log.Fatalf("restore failed: %s", err.Error())
		}
	} else {
		log.Fatal("Snapshot not found")
	}
}

func List(cfg *config.Config, backupArguments BackupArguments) {
	backupProviderImpl := setupBackupProvider(cfg)
	snapshots, err := backupProviderImpl.ListSnapshots(backupArguments.Paths)
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

func findSnapshotByID(snapshots []*providers.Snapshot, id string) (*providers.Snapshot, bool) {
	for i := range snapshots {
		if snapshots[i].ID == id {
			return snapshots[i], true
		}
	}
	return nil, false
}

func createSnapshotMountTarget() (string, error) {
	tmpdir := os.TempDir()
	target_tmpdir, err := os.MkdirTemp(tmpdir, "snapshot-")
	if err != nil {
		return "", err
	}
	return target_tmpdir, nil
}
