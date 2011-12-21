// Copyright 2010  The "GoWizard" Authors
//
// Use of this source code is governed by the BSD 2-Clause License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package wizard

import (
	"bufio"
	"errors"
	"fmt"
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

	dirProject := p.cfg.Project
	if !p.cfg.IsNewProject {
		dirProject = "." // actual directory
	}

	dirData := filepath.Join(p.dirData, "license")
	license := ListLowerLicense[p.cfg.License]

	licenseDst := func(name string) string {
		if p.cfg.IsUnlicense {
			name = "UNLICENSE.txt"
		} else if p.cfg.IsNewProject {
			name = "LICENSE.txt"
		} else {
			name = "LICENSE_" + name + ".txt"
		}

		return filepath.Join(dirProject, name)
	}

	switch lic := p.cfg.License; lic {
	case "none":
		break
	case "bsd-2", "bsd-3":
		p.parseFromFile(licenseDst(license), filepath.Join(dirData, license+".txt"))
		addPatent = true
	default:
		copyFile(licenseDst(license), filepath.Join(dirData, license+".txt"))

		// The license LGPL must also add the GPL license text.
		if lic == "lgpl" {
			tmp := p.cfg.IsNewProject

			p.cfg.IsNewProject = false
			copyFile(licenseDst("GPL"), filepath.Join(dirData, "GPL.txt"))
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

		p.parseFromFile(filepath.Join(dirProject, "PATENTS.txt"),
			filepath.Join(dirData, "PATENTS.txt"))
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
func dirData() (string, error) {
	goEnv := os.Getenv("GOPATH")

	if goEnv != "" {
		goto _Found
	}
	if goEnv = os.Getenv("GOROOT"); goEnv != "" {
		goto _Found
	}
	if goEnv = os.Getenv("GOROOT_FINAL"); goEnv != "" {
		goto _Found
	}

_Found:
	if goEnv == "" {
		return "", errors.New("environment variable GOROOT neither" +
			" GOROOT_FINAL has been set")
	}

	return filepath.Join(goEnv, _DIR_DATA), nil
}
