package scratchdb

import (
	"testing"
	"os"
)

func TestOpenDB(t *testing.T) {
	defer os.RemoveAll("test")

	testDB, err := Open("test", Options{true, false})
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	if _, err := os.Stat("test"); os.IsNotExist(err) {
		t.Fatalf("expected directory 'test' to exist")
	}

	if _, err := os.Stat(testDB.activeFile); os.IsNotExist(err) {
		t.Fatalf("expected active file %s to exist", testDB.activeFile)
	}

	if testDB.options.ReadWrite == false {
		t.Fatalf("option.ReadWrite == false; should be true")
	}

	if testDB.options.SyncOnPut == true {
		t.Fatalf("options.SyncOnPut == true; should be false")
	}
}

func TestOpenCloseOpen(t *testing.T) {
	defer os.RemoveAll("test")

	testDB, err := Open("test", Options{true, true})
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	err = testDB.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	testDB, err = Open("test", Options{true, true})
	if err != nil {
		t.Fatalf("failed to open existing: %v", err)
	}
}

func TestOpenPut(t *testing.T) {}

func TestNotOpenPut(t *testing.T) {}

// TODO 3: Only one process can have ReadWrite: true
func TestReadWriterConflict(t *testing.T) {}