package main

import (
	"bytes"
	"log"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "testKey"
	//expected := "testKey" // Assuming the default transform function returns the key unchanged

	pathname := CASPathTransformFunc(key)
	//result := transformFunc(key)
	log.Println(pathname)

}
func TestStore(t *testing.T) {

	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)

	reader := bytes.NewReader([]byte("test data"))
	if err := store.writeStream("testKey", reader); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if store == nil {
		t.Errorf("Expected store to be created, got nil")
	}
	//if store.StoreOpts.PathTransformFunc == nil {
}
