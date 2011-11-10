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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	// Permissions
	_PERM_DIRECTORY = 0755
	_PERM_FILE      = 0644

	_CHAR_CODE_COMMENT = "//" // For comments in source code files
	_CHAR_HEADER       = "="  // Header under the project name

	// Configuration file per user
	_USER_CONFIG = ".gowizard"

	// Subdirectory where is installed through "goinstall"
	_SUBDIR_GOINSTALLED = "src/pkg/github.com/kless/GoWizard/data"
)
/*
// VCS configuration files to push to a server.
var configVCS = map[string]string{
	"bzr": ".bzr/branch/branch.conf",
	"git": ".git/config",
	"hg":  ".hg/hgrc",
}
*/
// Project types
var ListProject = map[string]string{
	"cmd": "Command line program",
	"pkg": "Package",
	"cgo": "Package that calls C code",
}

// Available licenses
var ListLicense = map[string]string{
	"apache-2": "Apache License, version 2.0",
	"bsd-2":    "BSD-2 Clause license",
	"bsd-3":    "BSD-3 Clause license",
	"cc0-1":    "Creative Commons CC0, version 1.0 Universal",
	"gpl-3":    "GNU General Public License, version 3 or later",
	"lgpl-3":   "GNU Lesser General Public License, version 3 or later",
	"agpl-3":   "GNU Affero General Public License, version 3 or later",
	"none":     "Proprietary license",
}

// Version control systems (VCS)
var ListVCS = map[string]string{
	"bzr":   "Bazaar",
	"git":   "Git",
	"hg":    "Mercurial",
	"other": "other VCS",
	"none":  "none",
}

// * * *

// Represents all information to create a project
type project struct {
	dirData    string // directory with templates
	dirProject string // directory of project created

	cfg *Conf
	set *template.Set // set of templates
}

// Creates information for the project.
// "isFirstRun" indicates if it is the first time in be called.
func NewProject(cfg *Conf, isFirstRun bool) (*project, error) {
	var err error

	p := new(project)
	if isFirstRun {
		if p.dirData, err = dirData(); err != nil {
			return nil, err
		}
	}
	p.dirProject = filepath.Join(cfg.ProjectName, cfg.PackageName)
	p.set = new(template.Set)
	p.cfg = cfg

	return p, nil
}

// Adds license file in directory `dir`.
func (p *project) addLicense(dir string) {
	dirTmpl := filepath.Join(p.dirData, "license")
	lic := p.cfg.License

	switch lic {
	case "none":
		break
	case "bsd-2", "bsd-3":
		p.parseFromFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, lic+".txt"), true)
	default:
		copyFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, lic+".txt"), _PERM_FILE)

		// License LGPL must also add the GPL license text.
		if lic == "lgpl-3" {
			copyFile(filepath.Join(dir, "LICENSE-GPL"),
				filepath.Join(dirTmpl, "gpl-3.txt"), _PERM_FILE)
		}
	}
}

// Creates a new project.
func (p *project) Create() error {
	if err := os.MkdirAll(p.dirProject, _PERM_DIRECTORY); err != nil {
		return fmt.Errorf("directory error: %s", err)
	}

	p.parseTemplates(_CHAR_CODE_COMMENT, 0)

	// === Render project files
	if p.cfg.ProjecType != "cmd" {
		p.parseFromVar(filepath.Join(p.dirProject, p.cfg.PackageName)+".go",
			"Pkg")
		p.parseFromVar(filepath.Join(p.dirProject, p.cfg.PackageName)+"_test.go",
			"Test")
	} else {
		p.parseFromVar(filepath.Join(p.dirProject, p.cfg.PackageName)+".go",
			"Cmd")
	}
	p.parseFromVar(filepath.Join(p.dirProject, "Makefile"), "Makefile")

	// === Render common files
	dirTmpl := filepath.Join(p.dirData, "templ") // Base directory of templates

	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "CONTRIBUTORS.md"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.md"), false)
	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "NEWS.md"),
		filepath.Join(dirTmpl, "NEWS.md"), false)
	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "README.md"),
		filepath.Join(dirTmpl, "README.md"), true)

	// The file AUTHORS is for copyright holders.
	if !strings.HasPrefix(p.cfg.License, "cc0") {
		p.parseFromFile(filepath.Join(p.cfg.ProjectName, "AUTHORS.md"),
			filepath.Join(dirTmpl, "AUTHORS.md"), false)
	}

	// === Add file related to VCS
	switch p.cfg.VCS {
	case "other", "none":
		break
	default:
		ignoreFile := "." + p.cfg.VCS + "ignore"

		if p.cfg.VCS == "hg" {
			tmplIgnore = hgIgnoreTop + tmplIgnore
		}

		if err := ioutil.WriteFile(filepath.Join(p.cfg.ProjectName, ignoreFile),
			[]byte(tmplIgnore), _PERM_FILE); err != nil {
			return fmt.Errorf("write error: %s", err)
		}
	}

	// === License file
	p.addLicense(p.cfg.ProjectName)

	return nil
}
