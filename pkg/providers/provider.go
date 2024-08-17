package providers

// BackupProvider defines the methods that a backup provider must implement.
type BackupProvider interface {
	ListSnapshots() ([]*Snapshot, error)
	RestoreSnapshot(snapshotID string, targetLocation string) error
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
