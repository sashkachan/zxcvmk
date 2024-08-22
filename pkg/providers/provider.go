package providers

// type BackupProvider defines the methods that a backup provider must implement.
type BackupProvider interface {
	ListSnapshots(filterPaths []string) ([]*Snapshot, error)
	MountSnapshot(snapshotID string, mountPath string) error
	RestoreSnapshot(snapshotID string, target string, paths []string) error
}

type Snapshot struct {
	Time     string   `json:"time"`
	Tree     string   `json:"tree"`
	Paths    []string `json:"paths"`
	Hostname string   `json:"hostname"`
	Username string   `json:"username"`
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
	ID       string   `json:"id"`
	ShortID  string   `json:"short_id"`
}
