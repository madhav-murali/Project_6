package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "storage"

type PathTransformFunc func(string) PathKey

var CASPathTransformFunc PathTransformFunc = func(key string) PathKey {
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
	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: key,
	} // Replace with actual transformation logic if needed
}

type PathKey struct {
	Pathname string
	Filename string
}

// FullPath returns the full path including the key.
// This is useful for constructing the complete file path in the storage.
func (pk PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", pk.Pathname, pk.Filename)
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		Pathname: key, // Default implementation returns the key as the pathname
		Filename: key, // Default implementation returns the key unchanged
	} // Default implementation returns the path unchanged
}

type StoreOpts struct {
	// Root directory for the storage of all files in the system
	Root string
	// Function to transform the path before saving
	PathTransformFunc PathTransformFunc
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
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc // Use default if not provided
	}
	if opts.Root == "" {
		opts.Root = defaultRootFolderName // Set default root folder if not provided
	}
	return &Store{
		StoreOpts: opts,
	}
}

// Checks if the key exists in the storage.
// It uses the PathTransformFunc to determine the path for the key.
func (s *Store) HasKey(key string) (bool, error) {
	pathKey := s.PathTransformFunc(key)
	if pathKey.Pathname == "" {
		return false, fmt.Errorf("invalid path for key: %s", key)
	}

	pathAndFileName := pathKey.FullPath() // Get the full path and filename
	if _, err := os.Stat(pathAndFileName); os.IsNotExist(err) {
		return false, nil // Key does not exist
	} else if err != nil {
		return false, fmt.Errorf("error checking file %s: %w", pathAndFileName, err)
	}

	log.Printf("Key [%s] exists in storage.\n", key)
	return true, nil // Key exists
}

// FirstFolder returns the first folder in the path.
// This is useful for determining the first segment of the path.
func FirstFolder(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Delete removes the data associated with the given key from the storage.
// It returns an error if the key does not exist or if there is an issue deleting the
// data.
// It uses the PathTransformFunc to determine the path for the key.
func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	if pathKey.Pathname == "" {
		return fmt.Errorf("invalid path for key: %s", key)
	}

	pathAndFileName := pathKey.FullPath() // Get the full path and filename

	// Remove the file from the disk
	if err := os.RemoveAll(FirstFolder(pathAndFileName)); err != nil {
		return fmt.Errorf("error deleting file %s: %w", pathAndFileName, err)
	}

	log.Printf("Deleted file [%s] from disk.\n", pathAndFileName)
	return nil
}

// Read reads the data from the storage for the given key.
// It returns an io.Reader that can be used to read the data.
// If the key does not exist, it returns an error.
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.ReadStream(key)
	if err != nil {
		return nil, fmt.Errorf("error reading stream for key %s: %w", key, err)
	}
	defer f.Close() // Ensure the file is closed after reading
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return nil, fmt.Errorf("error copying data from file %s: %w", key, err)
	}
	return buf, nil
	// Implement the logic to read data from the storage.
}

// Used to read the stream from the storage.
// This function opens the file corresponding to the key and returns an io.ReadCloser.
func (s *Store) ReadStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	if pathKey.Pathname == "" {
		return nil, fmt.Errorf("invalid path for key: %s", key)
	}

	pathAndFileName := pathKey.FullPath() // Get the full path and filename
	file, err := os.Open(pathAndFileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", pathAndFileName, err)
	}

	log.Printf("Opened file %s for reading\n", pathAndFileName)
	return file, nil
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		return err
	}

	// buf := new(bytes.Buffer)
	// io.Copy(buf, r) // Read the data from the reader into a buffer

	// filenameBytes := md5.Sum(buf.Bytes())            // Create an MD5 hash of the data
	// filename := hex.EncodeToString(filenameBytes[:]) // Convert the hash to a string
	pathAndFileNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	file, err := os.Create(pathAndFileNameWithRoot)
	if err != nil {
		return err
	}

	defer file.Close() // Ensure the file is closed after writing

	n, err := io.Copy(file, r)
	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s\n", n, pathAndFileNameWithRoot)

	if pathKey.Pathname == "" {
		return nil // No transformation, return early
	}
	return nil
}
