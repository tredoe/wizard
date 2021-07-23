// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package wizard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// copyFile copies a file from source to destination.
func copyFile(destination, source string) error {
	src, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("copy error reading: %s", err)
	}

	err = os.WriteFile(destination, src, _FILE_PERM)
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
	if err = file.Chmod(_FILE_PERM); err != nil {
		return nil, fmt.Errorf("file error: %s", err)
	}

	return file, nil
}

// getProjectName returns the project name from Readme file.
// It should be in the first line.
func getProjectName() (string, error) {
	info, err := os.Stat(_README)
	if os.IsNotExist(err) {
		info, err = os.Stat(filepath.Join("..", _README))
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file %s not found", _README)
		}
	}

	file, err := os.Open(info.Name())
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	line, _, err := buf.ReadLine()

	if err != nil {
		return "", err
	}
	return string(line), nil
}
