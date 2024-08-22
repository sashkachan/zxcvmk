package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ResticProvider struct {
	BackupRepositoryPasswordLocation string
	BackupRepository                 string
}

// NewResticProvider creates a new instance of ResticProvider.
func NewResticProvider(passwordLocation, repository string) *ResticProvider {
	return &ResticProvider{
		BackupRepositoryPasswordLocation: passwordLocation,
		BackupRepository:                 repository,
	}
}

// ListSnapshots returns a list of available snapshots from the restic repository.
func (r ResticProvider) ListSnapshots(filterPaths []string) ([]*Snapshot, error) {
	command := []string{"restic", "snapshots", "--json"}
	command = append(command, "-r", r.BackupRepository)
	if len(filterPaths) > 0 {
		for _, path := range filterPaths {
			command = append(command, "--path", path)
		}
	}
	cmd := exec.Command(command[0], command[1:]...)
	if r.BackupRepositoryPasswordLocation != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("RESTIC_PASSWORD_FILE=%s", r.BackupRepositoryPasswordLocation))
	}

	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("error executing command: %w", err)
	}

	var snapshots []*Snapshot
	if err = json.Unmarshal(output, &snapshots); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return snapshots, nil
}

// RestoreSnapshot restores a specific snapshot to the given location.
func (r ResticProvider) RestoreSnapshot(snapshotID string, target string, paths []string) error {
	if snapshotID == "" {
		return errors.New("snapshotID cannot be empty")
	}
	args := []string{"restore", snapshotID, "-r", r.BackupRepository}
	if len(paths) > 0 {
		for _, path := range paths {
			args = append(args, "--path", path)
			args = append(args, "--include", path)
		}
	}
	args = append(args, "--target", target)
	finfo, err := os.Stat(target)
	if os.IsNotExist(err) {
		return fmt.Errorf("restic restore failed as the target %s does not exist", target)
	}
	if !finfo.IsDir() {
		return fmt.Errorf("restic restore failed as the target %s is not a directory", target)
	}
	cmd := exec.Command("restic", args...)
	if r.BackupRepositoryPasswordLocation != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("RESTIC_PASSWORD_FILE=%s", r.BackupRepositoryPasswordLocation))
	}

	combined_output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore restore failed: %s", string(combined_output))
	}
	return err
}

// MountSnapshot
func (r ResticProvider) MountSnapshot(snapshotID string, mountPath string) error {
	// TODO: implement

	return nil
}
