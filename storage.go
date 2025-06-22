package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) string

func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5 // Define the block size for the path;
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	// This function transforms the key into a path suitable for CAS storage.
	// For example, it could hash the key or format it in a specific way.
	// Here, we simply return the key as a placeholder.
	return strings.Join(paths, "/") // Replace with actual transformation logic if needed
}

var DefaultPathTransformFunc = func(key string) string {
	return key // Default implementation returns the path unchanged
}

type StoreOpts struct {
	// Save stores the data in the storage.
	PathTransformFunc PathTransformFunc // Function to transform the path before saving
	// SaveFunc          func(data []byte) error // Function to save data
}

type Store struct {
	//StoreOpts
	StoreOpts
}

// func (s *Store) Save(data []byte) error {
// 	// Implement the logic to save data to the storage.
// 	// This is a placeholder implementation.
// 	return nil
// }

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r) // Read the data from the reader into a buffer

	filenameBytes := md5.Sum(buf.Bytes())            // Create an MD5 hash of the data
	filename := hex.EncodeToString(filenameBytes[:]) // Convert the hash to a string
	pathAndFileName := pathName + "/" + filename

	file, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(file, buf)
	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s\n", n, pathAndFileName)

	if pathName == "" {
		return nil // No transformation, return early
	}
	return nil
}
