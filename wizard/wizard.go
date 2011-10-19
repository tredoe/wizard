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
	"log"
	"os"
	"path/filepath"
	"strings"
	"template"
)

//
// === Variables

const (
	ERROR = 1 // Exit status code if there is any error.

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

// VCS configuration files to push to a server.
var configVCS = map[string]string{
	"bzr": ".bzr/branch/branch.conf",
	"git": ".git/config",
	"hg":  ".hg/hgrc",
}

// Project types
var listProject = map[string]string{
	"cmd": "Command line program",
	"pkg": "Package",
	"cgo": "Package that calls C code",
}

// Available licenses
var listLicense = map[string]string{
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
var listVCS = map[string]string{
	"bzr":   "Bazaar",
	"git":   "Git",
	"hg":    "Mercurial",
	"other": "other VCS",
	"none":  "none",
}

//
// === Type

// Represents all information to create a project
type project struct {
	dirData    string // directory with templates
	dirProject string // directory of project created

	cfg *conf
	set *template.Set // set of templates
}

// Creates information for the project.
// "isFirstRun" indicates if it is the first time in be called.
func NewProject(isFirstRun bool) *project {
	cfg := initConfig()

	p := new(project)
	if isFirstRun {
		p.dirData = dirData()
	}
	p.dirProject = filepath.Join(cfg.ProjectName, cfg.PackageName)
	p.set = new(template.Set)
	p.cfg = cfg

	return p
}

// * * *

// Adds license file in directory `dir`.
func (p *project) addLicense(dir string) {
	dirTmpl := filepath.Join(p.dirData, "license")

	switch p.cfg.license {
	case "none":
		break
	case "bsd-3":
		p.parseFromFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, "bsd-3.txt"), false)
	default:
		copyFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, p.cfg.license+".txt"), _PERM_FILE)

		// License LGPL must also add the GPL license text.
		if p.cfg.license == "lgpl-3" {
			copyFile(filepath.Join(dir, "LICENSE-GPL"),
				filepath.Join(dirTmpl, "gpl-3.txt"), _PERM_FILE)
		}
	}
}

// Creates a new project.
func (p *project) Create() {
	if err := os.MkdirAll(p.dirProject, _PERM_DIRECTORY); err != nil {
		log.Fatal("directory error:", err)
	}

	p.parseTemplates(_CHAR_CODE_COMMENT, 0)

	// === Render project files
	if p.cfg.projecType != "cmd" {
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
	dirTmpl := filepath.Join(p.dirData, "tmpl") // Base directory of templates

	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "CONTRIBUTORS.mkd"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.mkd"), false)
	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "NEWS.mkd"),
		filepath.Join(dirTmpl, "NEWS.mkd"), false)
	p.parseFromFile(filepath.Join(p.cfg.ProjectName, "README.mkd"),
		filepath.Join(dirTmpl, "README.mkd"), true)

	// The file AUTHORS is for copyright holders.
	if !strings.HasPrefix(p.cfg.license, "cc0") {
		p.parseFromFile(filepath.Join(p.cfg.ProjectName, "AUTHORS.mkd"),
			filepath.Join(dirTmpl, "AUTHORS.mkd"), false)
	}

	// === Add file related to VCS
	switch p.cfg.vcs {
	case "other", "none":
		break
	default:
		ignoreFile := "." + p.cfg.vcs + "ignore"

		if p.cfg.vcs == "hg" {
			tmplIgnore = hgIgnoreTop + tmplIgnore
		}

		if err := ioutil.WriteFile(filepath.Join(p.cfg.ProjectName, ignoreFile),
			[]byte(tmplIgnore), _PERM_FILE); err != nil {
			log.Fatal("write error:", err)
		}
	}

	// === License file
	p.addLicense(p.cfg.ProjectName)

	// === User configuration file
	if p.cfg.addUserConf {
		envHome := os.Getenv("HOME")

		if envHome != "" {
			p.parseFromVar(filepath.Join(envHome, _USER_CONFIG), "Config")
		} else {
			log.Print("could not add user configuration file because $HOME is not set")
		}
	}

	// === Print messages
	if p.cfg.AuthorIsOrg {
		fmt.Print(`
  * The organization has been added as author.
    Update the CONTRIBUTORS file to add people.
`)
	}
}

//
// === Utility

// Gets the path of the templates directory.
func dirData() string {
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
		log.Fatal("Environment variable GOROOT neither" +
			" GOROOT_FINAL has been set")
	}

	return filepath.Join(goEnv, _SUBDIR_GOINSTALLED)
}
