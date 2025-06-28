package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "testKey"
	//expected := "testKey" // Assuming the default transform function returns the key unchanged

	pathname := CASPathTransformFunc(key)
	//result := transformFunc(key)
	log.Println(pathname.Pathname)

}

func TestStore(t *testing.T) {

	store := newstore()
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("testKey-%d", i)

		reader := bytes.NewReader([]byte("the data inside the file / reader"))

		if n, err := store.writeStream(key, reader); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		fmt.Println("Wrote", n, "bytes to store with key:", key)
		r, err := store.Read(key)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if has, err := store.HasKey(key); err != nil {
			t.Errorf("Expected no error checking key, got %v", err)
		} else if !has {
			t.Errorf("Expected key %s to exist, but it does not", key)
		}

		b, err := io.ReadAll(r)

		if err != nil {
			t.Errorf("Expected no error reading data, got %v", err)
		}
		fmt.Println("Data read from store:", string(b))

		if string(b) != "the data inside the file / reader" {
			t.Errorf("Expected data to be 'the data inside the file / reader', got '%s'", string(b))
		}

		if err = store.Delete(key); err != nil {
			t.Errorf("Expected no error deleting key, got %v", err)
		}

		if ok, err := store.HasKey(key); (ok) || err != nil {
			t.Errorf("Expected key %s to be deleted, but it still exists", key)
		}

	}
	defer teardown(t, store)

	//if store.StoreOpts.PathTransformFunc == nil {
}

// helper for creating a new store with default options
func newstore() *Store {
	opts := StoreOpts{
		Root:              defaultRootFolderName,
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)
	return store
}

// teardown function to clear the store after tests
// This function is called after each test to ensure the store is cleared.
func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Errorf("Expected no error clearing store, got %v", err)
	}
}
