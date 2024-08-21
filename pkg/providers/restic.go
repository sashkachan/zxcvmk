package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ResticProvider struct {
	SnapshotListCommand              []string
	BackupRepositoryPasswordLocation string
	BackupRepository                 string
}

// NewResticProvider creates a new instance of ResticProvider.
func NewResticProvider(snapshotListCommand []string, passwordLocation, repository string) *ResticProvider {
	return &ResticProvider{
		SnapshotListCommand:              snapshotListCommand,
		BackupRepositoryPasswordLocation: passwordLocation,
		BackupRepository:                 repository,
	}
}

// ListSnapshots returns a list of available snapshots from the restic repository.
func (r ResticProvider) ListSnapshots() ([]*Snapshot, error) {
	command := r.SnapshotListCommand[1:]
	command = append(command, "-r", r.BackupRepository)

	cmd := exec.Command(r.SnapshotListCommand[0], command...)
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
func (r ResticProvider) RestoreSnapshot(snapshotID string, targetLocation string) error {
	if snapshotID == "" {
		return errors.New("snapshotID cannot be empty")
	}

	cmd := exec.Command("restic", "restore", snapshotID, "--target", targetLocation)
	_, err := cmd.Output()
	return err
}

// MountSnapshot
func (r ResticProvider) MountSnapshot(snapshotID string, mountPath string) error {
	// TODO: implement

	return nil
}
