// Copyright 2010, The "gowizard" Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

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

