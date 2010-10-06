// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"container/vector"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)


// Gets an array from map keys.
func arrayKeys(m map[string]string) []string {
	a := make([]string, len(m))

	i := 0
	for k, _ := range m {
		a[i] = k
		i++
	}

	return a
}

// Creates a backup of a file.
func backup(fname string) (ok bool) {
	if err := copyFile(fname+"~", fname); err != nil {
		return false
	}
	return true
}

// Copies a file from source to destination.
func copyFile(destination, source string) os.Error {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destination, src, PERM_FILE)
	if err != nil {
		return err
	}

	return nil
}

// Creates a string of characters with length of `name` to use it under that name.
func header(name string) string {
	header := make([]byte, len(name))

	for i, _ := range header {
		header[i] = CHAR_HEADER
	}

	return string(header)
}


// === Implementation of interface 'Visitor' for 'path.Walk'
// ===
type finder struct {
	ext   string
	files vector.StringVector
}

func newFinder(ext string) *finder {
	_finder := new(finder)

	if ext != ".go" && ext != ".mkd" {
		panic("File extension not supported")
	}
	_finder.ext = ext

	return _finder
}

// Skips directories created on compilation.
func (self *finder) VisitDir(path string, f *os.FileInfo) bool {
	dirName := f.Name

	if dirName == "_test" || dirName == "_obj" {
		return false
	}

	return true
}

// Adds all files to the list, according to the extension.
func (self *finder) VisitFile(filePath string, f *os.FileInfo) {
	name := f.Name

	if self.ext == ".go" && path.Ext(name) == ".go" {
		self.files.Push(filePath)
	} else if self.ext == ".mkd" && path.Ext(name) == ".mkd" {

		self.files.Push(filePath)
	}
}

// ===

// Base to find all files with extension `ext` on path `pathName`.
func _finder(ext string, pathName string) []string {
	finder := newFinder(ext)
	path.Walk(pathName, finder, nil)

	if len(finder.files) == 0 {
		fmt.Fprintf(os.Stderr,
			"%s: no files with extension %q in directory %q\n",
			argv0, ext, pathName)
		os.Exit(ERROR)
	}

	return finder.files
}

// Finds all Go source files.
func finderGo(pathName string) []string {
	return _finder(".go", pathName)
}

// Finds all markup text files, except README.
func finderMkd(pathName string) []string {
	return _finder(".mkd", pathName)
}

