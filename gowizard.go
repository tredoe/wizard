// Copyright 2010  The "gowizard" Authors
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// http://mozilla.org/MPL/2.0/.

// Package gowizard allows to create the base of new Go projects and to add new
// packages or commands to the project.
package wizard

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const (
	// Permissions
	_PERM_DIRECTORY = 0755
	_PERM_FILE      = 0644

	CHAR_COMMENT = "//" // For comments in source code files
	_CHAR_HEADER = "="  // Header under the project name

	// Configuration file per user
	_USER_CONFIG = ".gowizard"

	README = "README.md"

	// Subdirectory where is installed through "go install"
	_DATA_PATH = "github.com/kless/gowizard/data"
)

/*// VCS configuration files to push to a server.
var configVCS = map[string]string{
	"bzr": ".bzr/branch/branch.conf",
	"git": ".git/config",
	"hg":  ".hg/hgrc",
}*/

// Project types
var ListType = map[string]string{
	"cmd": "Command line program",
	"pkg": "Package",
	"cgo": "Package that calls C code",
}

// Available licenses
var (
	ListLicense = map[string]string{
		"AGPL":      "GNU Affero General Public License, version 3 or later",
		"Apache":    "Apache License, version 2.0",
		"CC0":       "Creative Commons CC0, version 1.0 Universal",
		"GPL":       "GNU General Public License, version 3 or later",
		"MPL":       "Mozilla Public License, version 2.0",
		"Unlicense": "Public domain",
		"none":      "Proprietary license",
	}

	ListLowerLicense = map[string]string{
		"agpl":      "AGPL",
		"apache":    "Apache",
		"cc0":       "CC0",
		"gpl":       "GPL",
		"mpl":       "MPL",
		"unlicense": "Unlicense",
		"none":      "none",
	}

	ListLicenseURL = map[string]string{
		"agpl":      "http://www.gnu.org/licenses/agpl.html",
		"apache":    "http://www.apache.org/licenses/LICENSE-2.0",
		"cc0":       "http://creativecommons.org/publicdomain/zero/1.0/",
		"gpl":       "http://www.gnu.org/licenses/gpl.html",
		"mpl":       "http://mozilla.org/MPL/2.0/",
		"unlicense": "http://unlicense.org/",
	}
)

// Version control systems (VCS)
var ListVCS = map[string]string{
	"bzr":   "Bazaar",
	"git":   "Git",
	"hg":    "Mercurial",
	"other": "other VCS",
	"none":  "none",
}

// * * *

// project represents all information to create a project.
type project struct {
	dataDir string             // directory with templates
	tmpl    *template.Template // set of templates
	cfg     *Conf
}

// NewProject initializes information for a new project.
func NewProject(cfg *Conf) (*project, error) {
	// To get the path of the templates directory.
	pkg, err := build.Import(_DATA_PATH, build.Default.GOPATH, build.FindOnly)
	if err != nil {
		return nil, fmt.Errorf("NewProject: data directory not found: %s", err)
	}

	return &project{pkg.Dir, new(template.Template), cfg}, nil
}

// Create creates a new project.
func (p *project) Create() error {
	dirTmpl := filepath.Join(p.dataDir, "templ") // Base directory of templates

	if err := os.Mkdir(p.cfg.Program, _PERM_DIRECTORY); err != nil {
		return fmt.Errorf("directory error: %s", err)
	}
	if p.cfg.IsNewProject {
		if err := os.Mkdir(filepath.Join(p.cfg.Program, "doc"), _PERM_DIRECTORY); err != nil {
			return fmt.Errorf("directory error: %s", err)
		}
	}

	p.parseLicense(CHAR_COMMENT)
	p.parseProject()

	// Render project files
	p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+"_test.go", "Test")

	if p.cfg.Type != "cmd" {
		p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Pkg")
	} else {
		p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Cmd")
	}

	// == Option "add"
	if !p.cfg.IsNewProject {
		p.addLicense(".") // actual directory
		return nil
	}

	// License file
	p.addLicense(p.cfg.Program)

	// Render common files
	p.parseFromFile(filepath.Join(p.cfg.Program, "doc", "CONTRIBUTORS.md"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.md"))
	p.parseFromFile(filepath.Join(p.cfg.Program, "doc", "NEWS.md"),
		filepath.Join(dirTmpl, "NEWS.md"))
	p.parseFromFile(filepath.Join(p.cfg.Program, README),
		filepath.Join(dirTmpl, README))

	// The file AUTHORS is for copyright holders.
	if p.cfg.License != "unlicense" && p.cfg.License != "cc0" {
		p.parseFromFile(filepath.Join(p.cfg.Program, "doc", "AUTHORS.md"),
			filepath.Join(dirTmpl, "AUTHORS.md"))
	}

	// Add file related to VCS
	switch p.cfg.VCS {
	case "other", "none":
		break
	default:
		ignoreFile := "." + p.cfg.VCS + "ignore"
		//p.parseFromVar(filepath.Join(p.cfg.Program, ignoreFile), "Ignore")
		if err := ioutil.WriteFile(filepath.Join(p.cfg.Program, ignoreFile),
			[]byte(tmplIgnore), _PERM_FILE); err != nil {
			return fmt.Errorf("write error: %s", err)
		}
	}

	return nil
}
