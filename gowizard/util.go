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
	"io/ioutil"
	"os"
	"path"
)


/* Gets an array from map keys. */
func arrayKeys(m map[string]string) []string {
	a := make([]string, len(m))

	i := 0
	for k, _ := range m {
		a[i] = k
		i++
	}

	return a
}

/* Copies a file from source to destination. */
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

/* Create a string of characters with length of `name` to use it under that name.
*/
func header(name string) string {
	const char = '='

	header := make([]byte, len(name))

	for i, _ := range header {
		header[i] = char
	}

	return string(header)
}


// === Implementation of interface 'Visitor' for 'path.Walk'
// ===
type finderGo struct {
	files vector.StringVector
}

func newFinderGo() *finderGo {
	return &finderGo{}
}

/* Skips directories created on compilation. */
func (self *finderGo) VisitDir(path string, f *os.FileInfo) bool {
	dirName := f.Name

	if dirName == "_test" || dirName == "_obj" {
		return false
	}
	return true
}

/* Adds all Go files to the list. */
func (self *finderGo) VisitFile(filePath string, f *os.FileInfo) {
	name := f.Name

	if ext := path.Ext(name); ext == "go" {
		self.files.Push(filePath)
	}
}

