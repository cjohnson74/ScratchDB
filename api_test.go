package scratchdb

import (
	"log"
	"os"
	"testing"
)

func TestOpenDB(t *testing.T) {
	defer os.RemoveAll("test")

	testDB, err := Open("test", Options{true, false})
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer testDB.Close()
	

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

func TestCloseDB(t *testing.T) {
	defer os.RemoveAll("test")

	testDB, err := Open("test", Options{false, false})
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	testDB.Close()
}

func TestOpenCloseOpen(t *testing.T) {
	defer os.RemoveAll("test")

	testDB, err := Open("test", Options{true, true})

	err = testDB.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	testDB, err = Open("test", Options{true, true})
	if err != nil {
		t.Fatalf("failed to open existing: %v", err)
	}
}

func TestOpenPut(t *testing.T) {
	defer os.RemoveAll("test")
	testDB, err := Open("test", Options{true, true})
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer testDB.Close()

	key := []byte("Carson Johnson")
	value := []byte("Pizza")

	putErr := testDB.Put(key, value)
	if putErr != nil {
		t.Fatalf("Failed to Put into DB: %v", putErr)
	}

	entry, err := os.ReadFile(testDB.activeFile)
	if err != nil {
		t.Fatalf("failed to read from active file after put. %v", err)
	}
	log.Printf("%s", string(entry))

	if len(entry) == 0 {
		t.Fatalf("no content written to the DB")
	}

	log.Print(testDB.keyDir)

	if len(testDB.keyDir) == 0 {
		t.Fatalf("key dir empty after put.")
	}
}

func TestNotOpenPut(t *testing.T) {
	defer os.RemoveAll("test")
	testDB, err := Open("test", Options{true, true})
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	testDB.Close()

	key := []byte("Carson Johnson")
	value := []byte("Pizza")

	putErr := testDB.Put(key, value)

	if putErr == nil {
		t.Fatalf("expected put to fail on closed DB: %v", putErr)
	}
}

func TestNoWritePut(t *testing.T) {
	defer os.RemoveAll("test")
	testDB, err := Open("test", Options{false, true})
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	testDB.Close()

	key := []byte("Carson Johnson")
	value := []byte("Pizza")

	putErr := testDB.Put(key, value)
	if putErr == nil {
		t.Fatalf("expected put to fail with DB ReadWrite false: %v", putErr)
	}
}

func TestReadWriterConflict(t *testing.T) {
	defer os.RemoveAll("test")

	testDB1, err := Open("test", Options{true, true})
	log.Println(testDB1.options.ReadWrite)
	if err != nil {
		t.Fatalf("failed to open DB1: %v", err)
	}
	defer testDB1.Close()

	testDB2, err2 := Open("test", Options{true, true})
	if testDB2 != nil {
		testDB2.Close()
		t.Fatalf("db was created even though single writer should be enforced")
	}
	if err2 == nil {
		t.Fatalf("expected error opening second writer, got nil")
	}
}