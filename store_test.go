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

	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)
	key := "hello-man"
	reader := bytes.NewReader([]byte("the data inside the file / reader"))
	if err := store.writeStream(key, reader); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	r, err := store.Read(key)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("Expected no error reading data, got %v", err)
	}
	fmt.Println("Data read from store:", string(b))
	if string(b) != "the data inside the file / reader" {
		t.Errorf("Expected data to be 'the data inside the file / reader', got '%s'", string(b))
	}

	if store == nil {
		t.Errorf("Expected store to be created, got nil")
	}
	//if store.StoreOpts.PathTransformFunc == nil {
}
