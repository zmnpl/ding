package core

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func sortFilesByModTime(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		infi, _ := files[i].Info()
		infj, _ := files[j].Info()
		return infi.ModTime().After(infj.ModTime())
	})
}

// checkFixExtension fixes the new files extension based on the old one
func checkFixExtension(name, newName string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if !strings.HasSuffix(strings.ToLower(newName), ext) {
		newName += ext
	}
	return newName
}
