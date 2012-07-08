// Copyright 2010  The "Gowizard" Authors
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// http://mozilla.org/MPL/2.0/.

package wizard

import (
	"fmt"
	"io/ioutil"
	"os"
)

// copyFile copies a file from source to destination.
func copyFile(destination, source string) error {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("copy error reading: %s", err)
	}

	err = ioutil.WriteFile(destination, src, _PERM_FILE)
	if err != nil {
		return fmt.Errorf("copy error writing: %s", err)
	}

	return nil
}

// createFile creates a file.
func createFile(dst string) (*os.File, error) {
	file, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("file error: %s", err)
	}
	if err = file.Chmod(_PERM_FILE); err != nil {
		return nil, fmt.Errorf("file error: %s", err)
	}

	return file, nil
}
