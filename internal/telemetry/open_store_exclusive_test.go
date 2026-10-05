package telemetry

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression: OpenStore must not unlink -shm (or rename the DB) while another
// Store already holds the writer. Doing so splits the WAL index so a second
// opener's commits can be silently discarded.
func TestOpenStore_PreservesSHMWhileHeld(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "telemetry.db")

	holder, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("holder OpenStore: %v", err)
	}
	defer holder.Close()

	shmPath := dbPath + "-shm"
	before, err := os.Stat(shmPath)
	if err != nil {
		t.Fatalf("expected -shm after holder open: %v", err)
	}

	if _, err := holder.DB().Exec(`CREATE TABLE IF NOT EXISTS repro(x INTEGER); INSERT INTO repro(x) VALUES(1)`); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	second, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("second OpenStore: %v", err)
	}
	defer second.Close()

	after, err := os.Stat(shmPath)
	if err != nil {
		t.Fatalf("expected -shm to remain after second OpenStore: %v", err)
	}
	if !os.SameFile(before, after) {
		t.Fatalf("second OpenStore replaced -shm inode while holder was open")
	}

	if _, err := second.DB().Exec(`INSERT INTO repro(x) VALUES(2)`); err != nil {
		t.Fatalf("second write: %v", err)
	}
	if _, err := holder.DB().Exec(`INSERT INTO repro(x) VALUES(3)`); err != nil {
		t.Fatalf("holder write after second open: %v", err)
	}

	var holderCount, secondCount int
	if err := holder.DB().QueryRow(`SELECT COUNT(*) FROM repro`).Scan(&holderCount); err != nil {
		t.Fatalf("holder count: %v", err)
	}
	if err := second.DB().QueryRow(`SELECT COUNT(*) FROM repro`).Scan(&secondCount); err != nil {
		t.Fatalf("second count: %v", err)
	}
	if holderCount != 3 || secondCount != 3 {
		t.Fatalf("diverged counts holder=%d second=%d, want both 3", holderCount, secondCount)
	}

	backups, _ := filepath.Glob(dbPath + ".corrupt.*")
	if len(backups) != 0 {
		t.Fatalf("unexpected corrupt backups while DB was held open: %v", backups)
	}
}

func TestOpenStore_ExclusiveLockBlocksWhileHeld(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "telemetry.db")

	holder, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("holder OpenStore: %v", err)
	}
	defer holder.Close()

	probe, err := lockStoreShared(dbPath)
	if err != nil {
		t.Fatalf("shared lock: %v", err)
	}
	defer unlockStore(probe)

	if tryLockStoreExclusive(probe) {
		t.Fatal("exclusive store lock succeeded while holder Store is open")
	}

	if err := holder.Close(); err != nil {
		t.Fatalf("close holder: %v", err)
	}
	if !tryLockStoreExclusive(probe) {
		t.Fatal("exclusive store lock failed after holder closed")
	}
	downgradeStoreLockShared(probe)
}
