// Copyright 2010  The "Go-Wizard" Authors
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
)

// === Variables
// ===

const (
	ERROR = 1 // Exit status code if there is any error.

	// Permissions
	PERM_DIRECTORY = 0755
	PERM_FILE      = 0644
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
	"pac": "Package",
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

// === Type
// ===

// Represents all information to create a project
type info struct {
	dirData    string // directory with templates
	dirProject string // directory of project created

	data map[string]interface{} // variables to pass to templates
}

func NewInfo(isFirstRun bool) *info {
	i := new(info)

	if isFirstRun {
		i.dirData = dirData()
	}
	i.dirProject = dirProject()
	i.data = templateData()

	return i
}

// ===

// Adds license file in directory `dir`.
func (i *info) addLicense(dir string) {
	dirTmpl := filepath.Join(i.dirData, "license")

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		i.renderFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, "bsd-3.txt"))
	default:
		copyFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, *fLicense+".txt"), PERM_FILE)

		// License LGPL must also add the GPL license text.
		if *fLicense == "lgpl-3" {
			copyFile(filepath.Join(dir, "LICENSE-GPL"),
				filepath.Join(dirTmpl, "gpl-3.txt"), PERM_FILE)
		}
	}
}

// Creates a new project.
func (i *info) CreateProject() {
	if err := os.MkdirAll(i.dirProject, PERM_DIRECTORY); err != nil {
		log.Fatal("directory error:", err)
	}

	setTmpl := i.parseTemplates(_CHAR_CODE_COMMENT, 0)

	// === Render project files
	if *fProjecType != "cmd" {
		i.renderSet(filepath.Join(i.dirProject, *fPackageName)+".go",
			setTmpl, "Pac")
		i.renderSet(filepath.Join(i.dirProject, *fPackageName)+"_test.go",
			setTmpl, "Test")
	} else {
		i.renderSet(filepath.Join(i.dirProject, *fPackageName)+".go",
			setTmpl, "Cmd")
	}

	// === Render common files
	dirTmpl := filepath.Join(i.dirData, "tmpl") // Base directory of templates

	i.renderFile(filepath.Join(*fProjectName, "CONTRIBUTORS.mkd"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.mkd"))
	i.renderFile(filepath.Join(*fProjectName, "NEWS.mkd"),
		filepath.Join(dirTmpl, "NEWS.mkd"))
	i.renderFile(filepath.Join(*fProjectName, "README.mkd"),
		filepath.Join(dirTmpl, "README.mkd"))

	// The file AUTHORS is for copyright holders.
	if !strings.HasPrefix(*fLicense, "cc0") {
		i.renderFile(filepath.Join(*fProjectName, "AUTHORS.mkd"),
			filepath.Join(dirTmpl, "AUTHORS.mkd"))
	}

	// === Add file related to VCS
	switch *fVCS {
	case "other", "none":
		break
	default:
		ignoreFile := "." + *fVCS + "ignore"

		if *fVCS == "hg" {
			tmplIgnore = hgIgnoreTop + tmplIgnore
		}

		if err := ioutil.WriteFile(filepath.Join(*fProjectName, ignoreFile),
			[]byte(tmplIgnore), PERM_FILE); err != nil {
			log.Fatal("write error:", err)
		}
	}

	// === License file
	i.addLicense(*fProjectName)

	// === Print messages
	if i.data["author_is_org"].(bool) {
		fmt.Print(`
  * The organization has been added as author.
    Update `)

		if i.data["license_is_cc0"].(bool) {
			fmt.Print("AUTHORS")
		} else {
			fmt.Print("CONTRIBUTORS")
		}
		fmt.Print(" file to add people.\n")
	}
}

// === Utility
// ===

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

	return filepath.Join(goEnv, SUBDIR_GOINSTALLED)
}

// Gets the project directory.
func dirProject() string {
	return filepath.Join(*fProjectName, *fPackageName)
}

// Creates data to pass them to templates. Used at creating a new project.
func templateData() map[string]interface{} {
	var value bool

	data := map[string]interface{}{
		"project_name":    *fProjectName,
		"package_name":    *fPackageName,
		"author":          *fAuthor,
		"author_email":    *fAuthorEmail,
		"license":         listLicense[*fLicense],
		"vcs":             *fVCS,
		"_project_header": createHeader(*fProjectName),
	}

	if *fAuthorIsOrg {
		value = true
	}
	data["author_is_org"] = value
	value = false

	if *fProjecType == "cgo" {
		value = true
	}
	data["project_is_cgo"] = value

	return data
}
