// Copyright 2010  The "Gowizard" Authors
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package wizard allows to create the base of new Go projects and to add new
// packages or commands to the project.
package wizard

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"text/template"
)

const (
	// Permissions
	_DIRECTORY_PERM = 0755
	_FILE_PERM      = 0644

	_COMMENT_CHAR = "//" // For comments in source code files
	_HEADER_CHAR  = "="  // Header under the project name

	// Subdirectory where is installed through "go get"
	_DATA_PATH = "github.com/kless/wizard/data"

	_README      = "README.md"
	_USER_CONFIG = ".gowizard" // Configuration file per user
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

// Version control systems (VCS)
var ListVCS = map[string]string{
	"bzr":   "Bazaar",
	"git":   "Git",
	"hg":    "Mercurial",
	"other": "other VCS",
	"none":  "none",
}

// Available licenses
var (
	ListLicense = map[string]string{
		"AGPL":   "GNU Affero General Public License, version 3 or later",
		"Apache": "Apache License, version 2.0",
		"CC0":    "Creative Commons CC0, version 1.0 Universal",
		"GPL":    "GNU General Public License, version 3 or later",
		"MPL":    "Mozilla Public License, version 2.0",
		"none":   "Proprietary license",
	}
	ListLowerLicense = map[string]string{
		"agpl":   "AGPL",
		"apache": "Apache",
		"cc0":    "CC0",
		"gpl":    "GPL",
		"mpl":    "MPL",
		"none":   "none",
	}
	listLicenseURL = map[string]string{
		"agpl":   "http://www.gnu.org/licenses/agpl.html",
		"apache": "http://www.apache.org/licenses/LICENSE-2.0",
		"cc0":    "http://creativecommons.org/publicdomain/zero/1.0/",
		"gpl":    "http://www.gnu.org/licenses/gpl.html",
		"mpl":    "http://mozilla.org/MPL/2.0/",
	}
	listLicenseFaq = map[string]string{
		"agpl":   "http://www.gnu.org/licenses/gpl-faq.html",
		"apache": "http://www.apache.org/foundation/license-faq.html",
		"cc0":    "http://creativecommons.org/about/cc0",
		"gpl":    "http://www.gnu.org/licenses/gpl-faq.html",
		"mpl":    "http://www.mozilla.org/MPL/2.0/FAQ.html",
	}
)

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
	err := os.Mkdir(p.cfg.Program, _DIRECTORY_PERM)
	if err != nil {
		return fmt.Errorf("directory error: %s", err)
	}
	if p.cfg.IsNewProject {
		if err := os.Mkdir(filepath.Join(p.cfg.Program, "doc"),
			_DIRECTORY_PERM); err != nil {
			return fmt.Errorf("directory error: %s", err)
		}
	}

	p.parseLicense(_COMMENT_CHAR)
	p.parseProject()

	// == Render project files
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+"_test.go",
		"Test"); err != nil {
		return err
	}

	if p.cfg.Type != "cmd" {
		err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Pkg")
	} else {
		err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Cmd")
	}
	if err != nil {
		return err
	}

	// Option "add"
	if !p.cfg.IsNewProject {
		if err = p.addLicense("."); err != nil { // actual directory
			return err
		}
		return nil
	}

	// License file
	if err = p.addLicense(p.cfg.Program); err != nil {
		return err
	}

	// Render common files
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, "doc", "CONTRIBUTORS.md"),
		"Contributors"); err != nil {
		return err
	}
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, "doc", "NEWS.md"),
		"News"); err != nil {
		return err
	}
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, "doc", _README),
		"Readme"); err != nil {
		return err
	}

	// The file AUTHORS is for copyright holders.
	if p.cfg.License != "cc0" {
		if err = p.parseFromVar(filepath.Join(p.cfg.Program, "doc", "AUTHORS.md"),
			"Authors"); err != nil {
			return err
		}
	}

	// Add file related to VCS
	switch p.cfg.VCS {
	case "other", "none":
		break
	default:
		ignoreFile := "." + p.cfg.VCS + "ignore"
		if err = p.parseFromVar(filepath.Join(p.cfg.Program, ignoreFile),
			"Ignore"); err != nil {
			return err
		}
	}

	return nil
}

// * * *

// addLicense creates a license file.
func (p *project) addLicense(dir string) error {
	if p.cfg.License == "none" {
		return nil
	}

	license := ListLowerLicense[p.cfg.License]
	licenseDst := filepath.Join(dir, "doc", "LICENSE_"+license+".txt")

	// Check if it exist.
	if !p.cfg.IsNewProject {
		if _, err := os.Stat(licenseDst); !os.IsNotExist(err) {
			return nil
		}
	}

	if err := copyFile(licenseDst, filepath.Join(p.dataDir, license+".txt")); err != nil {
		return err
	}
	return nil
}
