// Copyright 2010  The "Go-Wizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const CHAR_HEADER = '=' // Header under the project name

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
func copyFile(destination, source string, perm uint32) os.Error {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destination, src, perm)
	if err != nil {
		return err
	}

	return nil
}

// Creates a string of characters with length of `name` to use it under that name.
func createHeader(name string) string {
	header := make([]byte, len(name))

	for i, _ := range header {
		header[i] = CHAR_HEADER
	}

	return string(header)
}

// Shows data on 'tag', execpt started with "_".
func debug(tag map[string]interface{}) {
	fmt.Println("  = Debug\n")

	for k, v := range tag {
		// Skip "_"
		if k[0] == '_' {
			continue
		}
		fmt.Printf("  %s: %v\n", k, v)
	}
	os.Exit(0)
}

