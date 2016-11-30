// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package wizard enables to create the base of new Go projects.
package wizard

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	// Permissions
	_DIR_PERM  = 0755
	_FILE_PERM = 0644

	_COMMENT_CHAR = "//" // For comments in source code files
	_HEADER_CHAR  = "="  // Header under the project name

	// Subdirectory where is installed through "go get"
	_DATA_PATH = "github.com/tredoe/wizard/data"

	_README      = "README.md"
	_USER_CONFIG = ".gowizard" // Configuration file per user
)

// Version control systems (VCS)
var (
	ListVCSsorted = []string{"bzr", "git", "hg", "none"}

	ListVCS = map[string]string{
		"bzr":  "Bazaar",
		"git":  "Git",
		"hg":   "Mercurial",
		"none": "none",
	}

	/*// VCS configuration files
	listConfigVCS = map[string]string{
		"bzr": ".bzr/branch/branch.conf",
		"git": ".git/config",
		"hg":  ".hg/hgrc",
	}*/
)

// Available licenses
var (
	ListLicenseSorted = []string{"AGPL", "Apache", "CC0", "GPL", "MPL", "none"}

	ListLicense = map[string]string{
		"AGPL":   "GNU Affero General Public License, version 3 or later",
		"Apache": "Apache License, version 2.0",
		"CC0":    "Creative Commons CC0, version 1.0 Universal",
		"GPL":    "GNU General Public License, version 3 or later",
		"MPL":    "Mozilla Public License, version 2.0",
		"none":   "proprietary license",
	}
	ListLowerLicense = map[string]string{
		"agpl":   "AGPL",
		"apache": "Apache",
		"cc0":    "CC0",
		"gpl":    "GPL",
		"mpl":    "MPL",
		"none":   "none",
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
func (p *project) Create() (err error) {
	dirs := []string{
		p.cfg.Program,
		filepath.Join(p.cfg.Program, "doc"),
		filepath.Join(p.cfg.Program, "testdata"),
	}
	for _, v := range dirs {
		if err = os.Mkdir(v, _DIR_PERM); err != nil {
			return fmt.Errorf("directory error: %s", err)
		}
	}

	p.parseLicense(_COMMENT_CHAR)
	p.parseProject()

	if len(p.cfg.ImportPaths) != 0 {
		p.cfg.ImportPath = path.Join(p.cfg.ImportPaths[0], p.cfg.Program)
	}

	// Render project files

	err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Go")
	if err != nil {
		return err
	}
	err = p.parseFromVar(filepath.Join(p.cfg.Program, "_"+p.cfg.Program)+"_test.go", "Test")
	if err != nil {
		return err
	}
	err = p.parseFromVar(filepath.Join(p.cfg.Program, "_example_test.go"), "Example")
	if err != nil {
		return err
	}

	// Add license file

	if p.cfg.License != "none" {
		license := ListLowerLicense[p.cfg.License]
		licenseDst := filepath.Join(p.cfg.Program, "LICENSE-"+license+".txt")

		err = copyFile(licenseDst, filepath.Join(p.dataDir, license+".txt"))
		if err != nil {
			return err
		}
	}

	// Render common files

	err = p.parseFromVar(filepath.Join(p.cfg.Program, _README), "Readme")
	if err != nil {
		return err
	}
	err = p.parseFromVar(filepath.Join(p.cfg.Program, "CONTRIBUTORS.txt.md"), "Contributors")
	if err != nil {
		return err
	}
	err = p.parseFromVar(filepath.Join(dirs[1], "_changelog.txt.md"), "Changelog")
	if err != nil {
		return err
	}

	// The file AUTHORS is for copyright holders.
	if p.cfg.License != "cc0" {
		err = p.parseFromVar(filepath.Join(p.cfg.Program, "AUTHORS.txt.md"), "Authors")
		if err != nil {
			return err
		}
	}

	// == VCS

	if p.cfg.VCS != "none" {
		ignoreFile := "." + p.cfg.VCS + "ignore"
		if err = p.parseFromVar(filepath.Join(p.cfg.Program, ignoreFile),
			"Ignore"); err != nil {
			return err
		}

		// Initialize VCS
		out, err := exec.Command(p.cfg.VCS, "init", p.cfg.Program).CombinedOutput()
		if err != nil {
			return err
		}
		if out != nil {
			out_ := string(out)
			if wd, err := os.Getwd(); err == nil {
				out_ = strings.Replace(out_, wd+string(os.PathSeparator), "", 1)
			}

			fmt.Print(out_)
		}
	}

	return nil
}
