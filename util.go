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
	"path/filepath"
)

// addLicense creates a license file.
func (p *project) addLicense(dir string) error {
	addPatent := false

	dataDir := filepath.Join(p.dataDir, "license")
	license := ListLowerLicense[p.cfg.License]

	licenseDst := func(name string) string {
		if p.cfg.IsUnlicense {
			name = "UNLICENSE.txt"
		} else if p.cfg.IsNewProject {
			name = "LICENSE.txt"
		} else {
			name = "LICENSE_" + name + ".txt"
		}

		return filepath.Join(dir, "doc", name)
	}

	switch lic := p.cfg.License; lic {
	case "none":
		break
	default:
		copyFile(licenseDst(license), filepath.Join(dataDir, license+".txt"))

		if lic == "unlicense" {
			addPatent = true
		}
	}

	if addPatent {
		// The owner is the organization, else the Authors of the project
		p.cfg.Owner = p.cfg.Org
		if p.cfg.Owner == "" {
			p.cfg.Owner = fmt.Sprintf("%q Authors", p.cfg.Project)
		}

		p.parseFromFile(filepath.Join(dir, "doc", "PATENTS.txt"),
			filepath.Join(dataDir, "PATENTS.txt"))
	}

	return nil
}

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
