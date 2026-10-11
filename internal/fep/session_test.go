package fep

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistryRoundTripIsAtomic(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	s := &Session{Env: "主机109", Zone: "36", Prog: "axmt500_wf", GDCPID: 4242, GDCPort: 7400, ProxyPort: 7600, InjectPort: 7899, Tag: "fep", Started: time.Now()}
	if err := r.Save(s); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(r.Path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file left behind: the write is not atomic")
	}
	got, err := r.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Prog != s.Prog || got.GDCPID != s.GDCPID || got.InjectPort != s.InjectPort {
		t.Fatalf("round trip mismatch: %+v", got)
	}
	if err := r.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := r.Load(); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after clear: %v", err)
	}
}

func TestRegistryReportsCorruption(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	if err := os.WriteFile(r.Path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := r.Load()
	if err == nil || errors.Is(err, ErrNoSession) {
		t.Fatalf("corrupt registry must be reported, got %v", err)
	}
}

func TestSessionPIDsOnlyListsRegistered(t *testing.T) {
	s := &Session{GDCPID: 7}
	pids := s.PIDs()
	if len(pids) != 1 || pids[0] != 7 {
		t.Fatalf("PIDs = %v", pids)
	}
	// Cleanup must never touch anything but the registered process.
	if (&Session{}).PIDs() != nil {
		t.Fatal("empty session must own nothing")
	}
}

func TestRegistryStopClearsEvenWhenProcessIsGone(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	if err := r.Save(&Session{GDCPID: 0x7fffffff}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, SessionFileName)); !os.IsNotExist(err) {
		t.Fatal("registry not cleared")
	}
}
