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
	// Permissions
	PERM_DIRECTORY = 0755
	PERM_FILE      = 0644

	SUBDIR_GOINSTALLED = "src/pkg/github.com/kless/Go-Wizard/data"
	//SUBDIR_GOINSTALLED = "lib/gowizard"
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

// === Get data directory

var dirData string

func init() {
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
		fatalf("Environment variable GOROOT neither" +
			" GOROOT_FINAL has been set\n")
	}

	dirData = filepath.Join(goEnv, SUBDIR_GOINSTALLED)
}

// === Main program execution
// ===

func main() {
	loadConfig()

	createProject()
	os.Exit(0)
}

// ===

// Adds license file in directory `dir`.
func addLicense(dir string, tag map[string]interface{}) {
	dirTmpl := filepath.Join(dirData, "license")

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		renderFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, "bsd-3.txt"), tag)
	default:
		if err := copyFile(filepath.Join(dir, "LICENSE"),
			filepath.Join(dirTmpl, *fLicense+".txt"), PERM_FILE); err != nil {
			reportExit(err)
		}

		// License LGPL must also add the GPL license text.
		if *fLicense == "lgpl-3" {
			if err := copyFile(filepath.Join(dir, "LICENSE-GPL"),
				filepath.Join(dirTmpl, "gpl-3.txt"), PERM_FILE); err != nil {
				reportExit(err)
			}
		}
	}
}

// Creates a new project.
func createProject() {
	tag := tagsToCreate()
	if *fDebug {
		debug(tag)
	}

	// === Render project files
	// To create directories in lower case.
	dirApp := filepath.Join(*fProjectName, *fPackageName)
	if err := os.MkdirAll(dirApp, PERM_DIRECTORY); err != nil {
		log.Fatal("directory error:", err)
	}

	setTmpl := parseTemplates(tag, CHAR_CODE_COMMENT, 0)

	if *fProjecType != "cmd" {
		renderSet(filepath.Join(dirApp, *fPackageName)+".go",
			setTmpl, "Pac", tag)
		renderSet(filepath.Join(dirApp, *fPackageName)+"_test.go",
			setTmpl, "Test", tag)
	} else {
		renderSet(filepath.Join(dirApp, *fPackageName)+".go",
			setTmpl, "Cmd", tag)
	}

	// === Render common files
	dirTmpl := filepath.Join(dirData, "tmpl") // Base directory of templates

	renderFile(filepath.Join(*fProjectName, "CONTRIBUTORS.mkd"),
		filepath.Join(dirTmpl, "CONTRIBUTORS.mkd"), tag)
	renderFile(filepath.Join(*fProjectName, "NEWS.mkd"),
		filepath.Join(dirTmpl, "NEWS.mkd"), tag)
	renderFile(filepath.Join(*fProjectName, "README.mkd"),
		filepath.Join(dirTmpl, "README.mkd"), tag)

	// The file AUTHORS is for copyright holders.
	if !strings.HasPrefix(*fLicense, "cc0") {
		renderFile(filepath.Join(*fProjectName, "AUTHORS.mkd"),
			filepath.Join(dirTmpl, "AUTHORS.mkd"), tag)
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
			reportExit(err)
		}
	}

	// === License file
	addLicense(*fProjectName, tag)

	// === Print messages
	if tag["author_is_org"].(bool) {
		fmt.Print(`
  * The organization has been added as author.
    Update `)

		if tag["license_is_cc0"].(bool) {
			fmt.Print("AUTHORS")
		} else {
			fmt.Print("CONTRIBUTORS")
		}
		fmt.Print(" file to add people.\n")
	}
}
