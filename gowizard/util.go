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
	"os"
)


/* Copies a file from source one to destination one. */
func CopyFile(destination, source string) os.Error {
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

