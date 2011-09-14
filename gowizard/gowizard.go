// Copyright 2010  The "Go-Wizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

	dirData = path.Join(goEnv, SUBDIR_GOINSTALLED)
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
func addLicense(dir string, tag map[string]string) {
	dirTmpl := dirData + "/license"

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		renderNewFile(dir+"/LICENSE", dirTmpl+"/bsd-3.txt", tag)
	default:
		if err := copyFile(dir+"/LICENSE",
			path.Join(dirTmpl, *fLicense+".txt"), PERM_FILE); err != nil {
			reportExit(err)
		}

		// License LGPL must also add the GPL license text.
		if *fLicense == "lgpl-3" {
			if err := copyFile(dir+"/LICENSE-GPL",
				path.Join(dirTmpl, "gpl-3.txt"), PERM_FILE); err != nil {
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

	headerCodeFile, headerMakefile := renderAllHeaders(tag, "")

	// === Render project files

	// To create directories in lower case.
	dirApp := path.Join(*fProjectName, *fPackageName)
	os.MkdirAll(dirApp, PERM_DIRECTORY)

	renderNesting(path.Join(dirApp, *fPackageName)+".go", headerCodeFile,
		tmplPkgMain, tag)
	renderNesting(dirApp+"/Makefile", headerMakefile, tmplPkgMakefile, tag)

	if *fProjecType != "cmd" {
		renderNesting(path.Join(dirApp, *fPackageName)+"_test.go",
			headerCodeFile, tmplTest, tag)
	}

	// === Render common files
	dirTmpl := dirData + "/tmpl" // Templates base directory

	renderFile(*fProjectName, dirTmpl+"/NEWS.mkd", tag)
	renderFile(*fProjectName, dirTmpl+"/README.mkd", tag)

	if strings.HasPrefix(*fLicense, "cc0") {
		renderNewFile(*fProjectName+"/AUTHORS.mkd",
			dirTmpl+"/AUTHORS-cc0.mkd", tag)
	} else {
		renderFile(*fProjectName, dirTmpl+"/AUTHORS.mkd", tag)
		renderFile(*fProjectName, dirTmpl+"/CONTRIBUTORS.mkd", tag)
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

		if err := ioutil.WriteFile(path.Join(*fProjectName, ignoreFile),
			[]byte(tmplIgnore), PERM_FILE); err != nil {
			reportExit(err)
		}
	}

	// === License file
	addLicense(*fProjectName, tag)

	// === Print messages
	if tag["author_is_org"] != "" {
		fmt.Print(`
  * The organization has been added as author.
    Update `)

		if tag["license_is_cc0"] != "" {
			fmt.Print("AUTHORS")
		} else {
			fmt.Print("CONTRIBUTORS")
		}
		fmt.Print(" file to add people.\n")
	}
}
