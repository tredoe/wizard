// Copyright 2010  The "GoWizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package wizard

import (
	"io/ioutil"
	"log"
	"os"
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

// Copies a file from source to destination.
func copyFile(destination, source string, perm uint32) {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatal("copy error reading:", err)
	}

	err = ioutil.WriteFile(destination, src, perm)
	if err != nil {
		log.Fatal("copy error writing:", err)
	}
}

// Creates a file.
func createFile(dst string) *os.File {
	file, err := os.Create(dst)
	if err != nil {
		log.Fatal("file error:", err)
	}
	if err = file.Chmod(_PERM_FILE); err != nil {
		log.Fatal("file error:", err)
	}

	return file
}

// Creates a string of characters with length of `name` to use it under that name.
func createHeader(name string) string {
	header := make([]byte, len(name))

	for i, _ := range header {
		header[i] = _CHAR_HEADER
	}

	return string(header)
}
