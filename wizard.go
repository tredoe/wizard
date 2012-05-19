// Copyright 2010  The "gowizard" Authors
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

package gowizard

import (
	"fmt"
	"go/build"
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
var ListLicense = map[string]string{
	"Apache":    "Apache License, version 2.0",
	"BSD-2":     "BSD 2-Clause License",
	"BSD-3":     "BSD 3-Clause License",
	"CC0":       "Creative Commons CC0, version 1.0 Universal (Not intended for software)",
	"GPL":       "GNU General Public License, version 3 or later",
	"LGPL":      "GNU Lesser General Public License, version 3 or later",
	"AGPL":      "GNU Affero General Public License, version 3 or later",
	"Unlicense": "Public domain",
	"none":      "Proprietary license",
}

var ListLowerLicense = map[string]string{
	"apache":    "Apache",
	"bsd-2":     "BSD-2",
	"bsd-3":     "BSD-3",
	"cc0":       "CC0",
	"gpl":       "GPL",
	"lgpl":      "LGPL",
	"agpl":      "AGPL",
	"unlicense": "Unlicense",
	"none":      "none",
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
	dataDir    string // directory with templates
	projectDir string // directory of project created

	cfg  *Conf
	tmpl *template.Template // set of templates
}

// Creates information for the project.
func NewProject(cfg *Conf) (*project, error) {
	var err error
	p := new(project)

	// == Gets the path of the templates directory
	pkg, err := build.Import(_DATA_PATH, build.Default.GOPATH, build.FindOnly)
	if err != nil {
		return nil, fmt.Errorf("NewProject: data directory not found: %s", err)
	}

	p.dataDir = pkg.Dir
	// ==

	if cfg.IsNewProject {
		p.projectDir = cfg.Project
	} else {
		p.projectDir = cfg.Program
	}

	p.cfg = cfg
	p.tmpl = new(template.Template)

	return p, nil
}

// Creates a new project.
func (p *project) Create() error {
	dirTmpl := filepath.Join(p.dataDir, "templ") // Base directory of templates

	if err := os.MkdirAll(p.projectDir, _PERM_DIRECTORY); err != nil {
		return fmt.Errorf("directory error: %s", err)
	}
	if p.cfg.IsNewProject {
		if err := os.Mkdir(filepath.Join(p.cfg.Project, "doc"), _PERM_DIRECTORY); err != nil {
			return fmt.Errorf("directory error: %s", err)
		}
	}

	p.parseLicense(CHAR_COMMENT)
	p.parseProject()

	// === License file
	p.addLicense()

	// === Render project files
	if p.cfg.Type != "cmd" {
		p.parseFromVar(filepath.Join(p.projectDir, p.cfg.Program)+".go",
			"Pkg")
		p.parseFromVar(filepath.Join(p.projectDir, p.cfg.Program)+"_test.go",
			"Test")
	} else {
		p.parseFromVar(filepath.Join(p.projectDir, p.cfg.Program)+".go",
			"Cmd")
	}

	// === Option "add"
	// ===
	if !p.cfg.IsNewProject {
		return nil
	}

	// === Render common files
	p.parseFromFile(filepath.Join(p.cfg.Project, "doc", "CONTRIBUTORS.md"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.md"))
	p.parseFromFile(filepath.Join(p.cfg.Project, "doc", "NEWS.md"),
		filepath.Join(dirTmpl, "NEWS.md"))
	p.parseFromFile(filepath.Join(p.cfg.Project, "README.md"),
		filepath.Join(dirTmpl, "README.md"))

	// The file AUTHORS is for copyright holders.
	if p.cfg.License != "unlicense" && p.cfg.License != "cc0" {
		p.parseFromFile(filepath.Join(p.cfg.Project, "doc", "AUTHORS.md"),
			filepath.Join(dirTmpl, "AUTHORS.md"))
	}

	// === Add file related to VCS
	switch p.cfg.VCS {
	case "other", "none":
		break
	default:
		ignoreFile := "." + p.cfg.VCS + "ignore"
		p.parseFromVar(filepath.Join(p.cfg.Project, ignoreFile), "Ignore")
	}

	return nil
}
