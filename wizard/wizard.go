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
	"template"
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

// === Type
// ===

// Represents all information to create a project
type project struct {
	dirData    string // directory with templates
	dirProject string // directory of project created

	set  *template.Set          // set of templates
	data map[string]interface{} // variables to pass to templates
}

// Creates information for the project.
// "isFirstRun" indicates if it is the first time in be called.
func NewProject(isFirstRun bool) *project {
	p := new(project)

	if isFirstRun {
		p.dirData = dirData()
	}
	p.dirProject = dirProject()
	p.data = templateData()
	p.set = new(template.Set)

	return p
}

// ===

// Adds license file in directory `dir`.
func (p *project) addLicense(dir string) {
	dirTmpl := filepath.Join(p.dirData, "license")

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		p.toTextFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, "bsd-3.txt"), false)
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
func (p *project) Create() {
	if err := os.MkdirAll(p.dirProject, PERM_DIRECTORY); err != nil {
		log.Fatal("directory error:", err)
	}

	p.parseTemplates(_CHAR_CODE_COMMENT, 0)

	// === Render project files
	if *fProjecType != "cmd" {
		p.toGoFile(filepath.Join(p.dirProject, *fPackageName)+".go",
			"Pkg")
		p.toGoFile(filepath.Join(p.dirProject, *fPackageName)+"_test.go",
			"Test")
	} else {
		p.toGoFile(filepath.Join(p.dirProject, *fPackageName)+".go",
			"Cmd")
	}

	// === Render common files
	dirTmpl := filepath.Join(p.dirData, "tmpl") // Base directory of templates

	p.toTextFile(filepath.Join(*fProjectName, "CONTRIBUTORS.mkd"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.mkd"), false)
	p.toTextFile(filepath.Join(*fProjectName, "NEWS.mkd"),
		filepath.Join(dirTmpl, "NEWS.mkd"), false)
	p.toTextFile(filepath.Join(*fProjectName, "README.mkd"),
		filepath.Join(dirTmpl, "README.mkd"), true)

	// The file AUTHORS is for copyright holders.
	if !strings.HasPrefix(*fLicense, "cc0") {
		p.toTextFile(filepath.Join(*fProjectName, "AUTHORS.mkd"),
			filepath.Join(dirTmpl, "AUTHORS.mkd"), false)
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
	p.addLicense(*fProjectName)

	// === Print messages
	if p.data["org"].(bool) {
		fmt.Print(`
  * The organization has been added as author.
    Update the CONTRIBUTORS file to add people.
`)
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
	data := map[string]interface{}{
		"project_name":    *fProjectName,
		"package_name":    *fPackageName,
		"org":             *fAuthorIsOrg,
		"author":          *fAuthor,
		"author_email":    *fAuthorEmail,
		"vcs":             *fVCS,
		"_project_header": createHeader(*fProjectName),
	}

	if *fLicense != "none" {
		data["license"] = listLicense[*fLicense]
	}
	if *fProjecType == "cgo" {
		data["is_cgo_project"] = true
	}

	return data
}
