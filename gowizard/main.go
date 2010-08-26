// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)


// Exit status code if there is any error
const ERROR = 2

// Permissions
const (
	PERM_DIRECTORY = 0755
	PERM_FILE      = 0644
)

// Metadata to build the new project
var cfg *metadata


// === Main program execution
func main() {
	if !*fUpdate {
		createProject()
	} else {
		updateProject()
	}

	os.Exit(0)
}

// ===

func createProject() {
	tag := loadConfig()             // Tags for templates
	header := renderHeader(tag, "") // Header with copyright and license

	cfg = NewMetadata(*fProjecType, *fProjectName, *fPackageName,
		*fAuthor, *fAuthorEmail, *fLicense, configFile())

	// === Create directories in lower case

	// Get data directory from `$(GOROOT)/lib/$(TARG)`
	dirData := path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

	projectName := cfg.ProjectName // Stores the name before of change it
	cfg.ProjectName = strings.ToLower(cfg.ProjectName)

	dirApp := path.Join(cfg.ProjectName, cfg.PackageName)
	os.MkdirAll(dirApp, PERM_DIRECTORY)

	// === Render project files
	switch cfg.ProjectType {
	case "lib", "cgo":
		renderCode(dirApp+"/Makefile", tmplMakefile, header["makefile"], tag)
		renderCode(dirApp+"/main.go", tmplPkgMain, header["code"], tag)
		renderCode(dirApp+"/main_test.go", tmplTest, header["code"], tag)
	case "app", "tool":
		renderCode(dirApp+"/Makefile", tmplMakefile, header["makefile"], tag)
		renderCode(dirApp+"/main.go", tmplCmdMain, header["code"], tag)
	}

	// === Render common files
	dirTmpl := dirData + "/tmpl" // Templates base directory

	renderFile(cfg.ProjectName, dirTmpl+"/README.mkd", tag)

	if strings.HasPrefix(cfg.License, "cc0") {
		renderNewFile(cfg.ProjectName+"/AUTHORS.mkd",
			dirTmpl+"/AUTHORS-cc0.mkd", tag)
	} else {
		renderFile(cfg.ProjectName, dirTmpl+"/AUTHORS.mkd", tag)
		renderFile(cfg.ProjectName, dirTmpl+"/CONTRIBUTORS.mkd", tag)
	}

	// === Add file related to VCS
	switch *fVCS {
	case "other":
		break
	// CHANGES is only necessary when is not used a VCS.
	case "none":
		renderFile(cfg.ProjectName, dirTmpl+"/CHANGES.mkd", tag)
	default:
		fileIgnore := *fVCS + "ignore"

		if err := copyFile(path.Join(cfg.ProjectName, "."+fileIgnore),
			path.Join(dirTmpl, fileIgnore)); err != nil {
			log.Exit(err)
		}
	}

	// === Add license file
	dirTmpl = dirData + "/license"

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		renderNewFile(cfg.ProjectName+"/LICENSE", dirTmpl+"/bsd-3.txt",
			tag)
	default:
		if err := copyFile(cfg.ProjectName+"/LICENSE",
			path.Join(dirTmpl, cfg.License+".txt")); err != nil {
			log.Exit(err)
		}
	}

	// === Create file Metadata
	cfg.ProjectName = projectName
	cfg.WriteINI(strings.ToLower(projectName))

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


func updateProject() {
	//tag := loadConfig()

}

