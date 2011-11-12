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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Creates the user configuration file.
func AddConfig(cfg *Conf) error {
	tmpl := template.Must(template.New("Config").Parse(tmplUserConfig))

	envHome := os.Getenv("HOME")
	if envHome == "" {
		return errors.New("could not add user configuration file because $HOME is not set")
	}

	file, err := createFile(filepath.Join(envHome, _USER_CONFIG))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("execution failed: %s", err)
	}
	return nil
}

// Creates a license file.
func AddLicense(p *project, isNewProject bool) error {
	licenseLower := p.cfg.License
	dirProject := p.cfg.ProjectName

	if !isNewProject {
		licenseLower = strings.ToLower(licenseLower)
		if ok := checkLicense(licenseLower); !ok {
			return errors.New("error")
		}

		dirProject = "." // actual directory
	}

	dirData := filepath.Join(p.dirData, "license")
	license := ListLicense[licenseLower][0]

	filename := func(name string) string {
		if strings.HasPrefix(name, "BSD") {
			name = strings.TrimRight(name, "-23")
		}
		return "LICENSE-"+name+".txt"
	}

	switch licenseLower {
	case "none":
		break
	case "bsd-2", "bsd-3":
		p.parseFromFile(filepath.Join(dirProject, filename(license)),
			filepath.Join(dirData, license+".txt"), true)
	default:
		copyFile(filepath.Join(dirProject, filename(license)),
			filepath.Join(dirData, license+".txt"), _PERM_FILE)

		// License LGPL must also add the GPL license text.
		if licenseLower == "lgpl" {
			copyFile(filepath.Join(dirProject, filename("GPL")),
				filepath.Join(dirData, "GPL.txt"), _PERM_FILE)
		}
	}

	return nil
}

// * * *

// Copies a file from source to destination.
func copyFile(destination, source string, perm uint32) error {
	src, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("copy error reading: %s", err)
	}

	err = ioutil.WriteFile(destination, src, perm)
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
		return "", errors.New("Environment variable GOROOT neither" +
			" GOROOT_FINAL has been set")
	}

	return filepath.Join(goEnv, _DIR_DATA), nil
}
