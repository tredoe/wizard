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

const (
	// Exit status code if there is any error
	ERROR = 2

	// Permissions
	PERM_DIRECTORY = 0755
	PERM_FILE      = 0644
)

var (
	argv0 = os.Args[0] // Executable name
	cfg   *metadata
)


// === Main program execution
func main() {
	tag := loadConfig()

	if !*fUpdate {
		createProject(tag)
	} else {
		updateProject(tag)
	}

	os.Exit(0)
}

// ===

/* Creates a new project. */
func createProject(tag map[string]string) {
	header := renderHeader(tag, "") // Header with copyright and license

	cfg = NewMetadata(*fProjecType, *fProjectName, *fPackageName, *fLicense,
		*fAuthor, *fAuthorEmail)

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
	if err := cfg.WriteINI(strings.ToLower(projectName)); err != nil {
		log.Exit(err)
	}

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

/* Updates some values from a project already created. */
func updateProject(tag map[string]string) {
	var updateProjectName, updatePackageName, updateLicense bool

	metadata, err := ReadMetadata()
	if err != nil {
		log.Exit(err)
	}

	// === See what updating
	if *fProjectName != "" && *fProjectName != metadata.ProjectName {
		metadata.ProjectName = *fProjectName
		updateProjectName = true
	}

	if *fPackageName != "" && *fPackageName != metadata.PackageName {
		metadata.PackageName = *fPackageName
		updatePackageName = true
	}

	if *fLicense != "" && *fLicense != metadata.License {
		metadata.License = *fLicense
		updateLicense = true
	}

	// === Get all Go source files
	finderGo := newFinderGo()
	path.Walk(metadata.PackageName, finderGo, nil)

	if len(finderGo.files) == 0 {
		fmt.Fprintf(os.Stderr,
			"%s: no Go source files in %q\n", argv0, metadata.PackageName)
	}

	// === Update license
	// ===
	if updateLicense || updateProjectName {
		println()
	}

	fmt.Println(finderGo.files, updatePackageName)
}

