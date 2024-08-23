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
					log.Printf("pre-restore-hook failed with %s: %s", err, result)
				}
			}
		}
	}
}

func rsyncPaths(from string, paths []string) error {
	for _, path := range paths {
		full_path := filepath.Join(from, path)
		if full_path[len(full_path)-1] != filepath.Separator {
			full_path = full_path + string(filepath.Separator)
		}
		rsyncArgs := []string{"-a", "-v", full_path, path}
		cmd := exec.Command("rsync", rsyncArgs...)
		cmd.Dir = from
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error rsync paths %s to %s: %s", from, path, output)
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
		defer func() {
			_ = deleteSnapshotMountTarget(target)
		}()
		if err != nil {
			log.Fatal("Snapshot target directory could not be created", err)
		}
		err = backupProviderImpl.RestoreSnapshot(snapshot.ID, target, backupArguments.Paths)
		if err != nil {
			log.Fatalf("restore failed: %s", err.Error())
			return
		}

		runPreRestoreHook(cfg, backupArguments.Paths)
		defer func() {
			runPostRestoreHook(cfg, backupArguments.Paths)
		}()
		err = rsyncPaths(target, backupArguments.Paths)
		if err != nil {
			log.Printf("failed to rsync contents: %s", err)
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

func deleteSnapshotMountTarget(target string) error {
	err := os.RemoveAll(target)
	if err != nil {
		log.Printf("error when removeall %s: %s", target, err)
		return err
	}
	return nil
}
