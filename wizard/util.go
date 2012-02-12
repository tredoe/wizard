// Copyright 2010  The "GoWizard" Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wizard

import (
	"bufio"
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Creates a license file.
func (p *project) addLicense() error {
	var addPatent bool

	projectDir := p.cfg.Project
	if !p.cfg.IsNewProject {
		projectDir = "." // actual directory
	}

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

		return filepath.Join(projectDir, name)
	}

	switch lic := p.cfg.License; lic {
	case "none":
		break
	case "bsd-2", "bsd-3":
		p.parseFromFile(licenseDst(license), filepath.Join(dataDir, license+".txt"))
		addPatent = true
	default:
		copyFile(licenseDst(license), filepath.Join(dataDir, license+".txt"))

		// The license LGPL must also add the GPL license text.
		if lic == "lgpl" {
			tmp := p.cfg.IsNewProject

			p.cfg.IsNewProject = false
			copyFile(licenseDst("GPL"), filepath.Join(dataDir, "GPL.txt"))
			p.cfg.IsNewProject = tmp
		}

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

		p.parseFromFile(filepath.Join(projectDir, "PATENTS.txt"),
			filepath.Join(dataDir, "PATENTS.txt"))
	}

	return nil
}

// Finds the first line that matches the copyright header to return the year.
func ProjectYear(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("no project directory: %s", err)
	}
	defer file.Close()

	fileBuf := bufio.NewReader(file)

	for {
		line, err := fileBuf.ReadString('\n')
		if err == io.EOF {
			break
		}

		if reCopyright.MatchString(line) || reCopyleft.MatchString(line) {
			for _, v := range strings.Split(line, " ") {
				if reYear.MatchString(v) {
					return strconv.Atoi(v)
				}
			}
		}
	}
	panic("unreached")
}

// * * *

// Copies a file from source to destination.
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

// Creates a file.
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

// Gets the path of the templates directory.
func dataDir() (string, error) {
	tree, pkg, err := build.FindTree(_DATA_PATH)
	if err != nil {
		return "", errors.New("dataDir: data directory not found")
	}

	return filepath.FromSlash(filepath.Join(tree.SrcDir(), pkg)), nil
}
