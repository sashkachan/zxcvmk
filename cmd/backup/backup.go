package backup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func runPostRestoreHook(cfg *config.Config, paths []string) {
	for _, path := range paths {
		for _, cfgPath := range cfg.BackupTargets {
			if path == cfgPath.Location && cfgPath.PostRestoreHook != nil {
				cmd := exec.Command(cfgPath.PostRestoreHook[0], cfgPath.PostRestoreHook[1:]...)
				result, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("error executing post-restore-hook for %s: %s", path, result)
				}
			}
		}
	}
}

func runPreRestoreHook(cfg *config.Config, paths []string) {
	for _, path := range paths {
		for _, cfgPath := range cfg.BackupTargets {
			if path == cfgPath.Location && cfgPath.PreRestoreHook != nil {
				cmd := exec.Command(cfgPath.PreRestoreHook[0], cfgPath.PreRestoreHook[1:]...)
				result, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("error executing pre-restore-hook for %s: %s", path, result)
				}
			}
		}
	}
}

func rsyncPaths(from string, paths []string) error {
	for _, path := range paths {
		source := filepath.Join(from, path)
		rsyncArgs := []string{"-a", "-v", "--relative", source, path}
		cmd := exec.Command("rsync", rsyncArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error rsync paths %s to %s: %s", source, path, output)
			return err
		}
	}
	return nil
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
			return
		}

		runPreRestoreHook(cfg, backupArguments.Paths)
		err = rsyncPaths(target, backupArguments.Paths)
		if err != nil {
			log.Printf("failed to rsync contents: %s", err)
		}
		runPostRestoreHook(cfg, backupArguments.Paths)
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
