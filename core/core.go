package core

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
)

var (
	Inbound = "."
	Dest    = "~/Documents"

	previewCache map[string]string
	previewsMu   sync.Mutex

	directoryFileCache map[string][]fs.DirEntry
	directoryFilesMu   sync.Mutex

	TimestampPrefixMatch, _ = regexp.Compile(`^\d\d\d\d\d\d\d\d-\d\d\d\d\d\d\.\d\d\d_(.*)$`)
)

func init() {
	previewCache = make(map[string]string)
	directoryFileCache = make(map[string][]fs.DirEntry)

	Dest, _ = homedir.Expand(Dest)
}

// GetInboundFiles returns a slice of all inbound files
func GetInboundFiles() ([]fs.DirEntry, error) {
	inboundFiles, err := os.ReadDir(Inbound)
	if err != nil {
		return nil, fmt.Errorf("could not read inbound directory: %s", err)
	}
	WarmInboundFilePreviewCache(inboundFiles)
	return inboundFiles, nil
}

// WarmInboundFilePreviewCache loads text previews for all inbound files into the cache
// the first preview will be loaded sync to have it available as soon as the it is displayed by some ui
// the rest will be loaded async in goroutines
func WarmInboundFilePreviewCache(files []fs.DirEntry) {
	for i, f := range files {
		if i == 0 {
			UpdateFilePreviewCache(f.Name())
			continue
		}
		go UpdateFilePreviewCache(f.Name())
	}
}

// UpdateFilePreviewCache upates the text preview for the given file name in the cache
func UpdateFilePreviewCache(filename string) {
	previewsMu.Lock()
	defer previewsMu.Unlock()
	previewCache[filename] = GetDocPreview(filename)
}

// GetCachedDocPreview returns the text preview for the given file name from the cache
func GetCachedDocPreview(name string) string {
	if val, ok := previewCache[name]; ok {
		return val
	}
	return ""
}

// GetDirectories returns a slice of the existing directories
func GetDirectories() ([]fs.DirEntry, error) {
	dirs := make([]fs.DirEntry, 0, 10)

	tmp, err := os.ReadDir(Dest)
	if err != nil {
		return nil, fmt.Errorf("could not read destination directory: %s", err)
	}

	for _, potDir := range tmp {
		if potDir.IsDir() {
			dirs = append(dirs, potDir)
		}
	}

	WarmDirectoryFilesCache(dirs)
	return dirs, nil
}

// WarmDirectoryFilesCache loads the list of files per directory and puts it into the cache
// the first element will be loaded sync to have it available as soon as the it is displayed by some ui
// the rest will be loaded async in goroutines
func WarmDirectoryFilesCache(directories []fs.DirEntry) {
	for i, b := range directories {
		if i == 0 {
			UpdateDirectoryFilesCache(b.Name())
		}
		go UpdateDirectoryFilesCache(b.Name())
	}
}

// UpdateDirectoryFilesCache upates the list of files for the given directory name in the cache
func UpdateDirectoryFilesCache(directoryname string) {
	directoryFilesMu.Lock()
	defer directoryFilesMu.Unlock()
	if bfs, _ := GetDirectoryFiles(directoryname); bfs != nil {
		directoryFileCache[directoryname] = bfs
	}
}

// GetCachedDirectoryFiles returns the list of files for the given directory name from the cache
func GetCachedDirectoryFiles(directoryname string) ([]fs.DirEntry, error) {
	if files, ok := directoryFileCache[directoryname]; ok {
		return files, nil
	}
	return make([]fs.DirEntry, 0), fmt.Errorf("could not find file list for %s", directoryname)
}

// GetDirectoryFiles returns a slice of files in the directory
func GetDirectoryFiles(directory string) ([]fs.DirEntry, error) {
	directoryFiles, err := os.ReadDir(filepath.Join(Dest, directory))
	if err != nil {
		return nil, fmt.Errorf("could not read directory directory: %s", err)
	}
	sortFilesByModTime(directoryFiles)
	return directoryFiles, nil
}

// CountDirectory returns the number of files in a directory
func CountDirectory(directory string) int {
	directoryFiles, err := os.ReadDir(filepath.Join(Dest, directory))
	if err != nil {
		return 0
	}

	cnt := 0
	for _, f := range directoryFiles {
		if f.IsDir() {
			continue
		}

		cnt++
	}

	return cnt
}

// MoveFileToDirectory moves the given file with the given new name to the given directory.
// This also triggers a cache update for this directory.
func MoveFileToDirectory(name, newName, directoryName string) (string, error) {
	// test if both source and destination can be written
	// destination
	destTest := filepath.Join(Dest, directoryName, "docin.test")
	err := ioutil.WriteFile(destTest, []byte("write test"), 0755)
	if err != nil {
		return "", fmt.Errorf("test to write destination failed: %s", err)
	}
	os.Remove(destTest)
	// inbound
	sourceTest := filepath.Join(Inbound, "docin.test")
	err = ioutil.WriteFile(sourceTest, []byte("write test"), 0755)
	if err != nil {
		return "", fmt.Errorf("test to write inbound failed: %s", err)
	}
	os.Remove(sourceTest)

	// read inbound file
	bytesRead, err := ioutil.ReadFile(filepath.Join(Inbound, name))
	if err != nil {
		return "", fmt.Errorf("could not read original file; did not move it: %s", err)
	}

	// write file to destination directory
	newName = checkFixExtension(name, newName)
	err = ioutil.WriteFile(filepath.Join(Dest, directoryName, newName), bytesRead, 0755)
	if err != nil {
		return "", fmt.Errorf("could not write file into directory; did not move it: %s", err)
	}

	// remove original
	err = os.Remove(filepath.Join(Inbound, name))
	if err != nil {
		return "", fmt.Errorf("could not remove original, please do manually!: %s", err)
	}

	go UpdateDirectoryFilesCache(directoryName)

	return newName, nil
}

// GetTimestampFilePrefix returns as timestamp prefix
// Quite long, not sure about that yet, but it avoids duplicates / overwrites
func GetTimestampFilePrefix() string {
	return time.Now().Format("20060102-150405.000_")
}

// RemoveTimeStampFilePrefix would remove the timestamp from the files name
// if the file is prefixed with a timestamp that matches this programs default
func RemoveTimeStampFilePrefix(name string) string {
	return TimestampPrefixMatch.ReplaceAllString(name, `$1`)
}
