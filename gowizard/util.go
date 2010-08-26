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
	"io/ioutil"
	"log"
	"os"

	"github.com/kless/goconfig/config"
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

/* Returns the INI configuration file. */
func configFile() (file *config.File) {
	var err os.Error

	if *fUpdate {
		if file, err = config.ReadFile(_FILE_NAME); err != nil {
			log.Exit(err)
		}
	} else {
		file = config.NewFile()
	}

	return
}

/* Create a string of characters with length of ProjectName to use under that name.
*/
func projectHeader() string {
	const char = '='

	header := make([]byte, len(*fProjectName))

	for i, _ := range header {
		header[i] = char
	}

	return string(header)
}

