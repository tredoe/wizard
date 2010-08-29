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


// Permissions
const (
	PERM_DIRECTORY = 0755
	PERM_FILE      = 0644
)

// Comment characters
const (
	COMMENT_CODE     = "//"
	COMMENT_MAKEFILE = "#"
)

var (
	bCommentCode     = []byte(COMMENT_CODE)
	bCommentMakefile = []byte(COMMENT_MAKEFILE)
)

const ERROR = 2 // Exit status code if there is any error

var argv0 = os.Args[0] // Executable name
var cfg *Metadata


// === Main program execution
func main() {
	loadConfig()

	if !*fUpdate {
		createProject()
	} else {
		updateProject()
	}

	os.Exit(0)
}

/* Show data on 'tag'. */
func debug(tag map[string]string) {
	fmt.Println("  = Debug\n")

	for k, v := range tag {
		// Tags starting with '_' are not showed.
		if k[0] == '_' {
			continue
		}
		fmt.Printf("  %s: %s\n", k, v)
	}
	os.Exit(0)
}

// ===


/* Creates a new project. */
func createProject() {
	tag := tagsToCreate()
	if *fDebug {
		debug(tag)
	}

	headerCode, headerMakefile := renderAllHeaders(tag, "")

	cfg = NewMetadata(*fProjecType, *fProjectName, *fPackageName, *fLicense,
		*fAuthor, *fAuthorEmail)

	// === Create directories in lower case

	// Get data directory from `$(GOROOT)/lib/$(TARG)`
	dirData := path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

	projectName := cfg.ProjectName // Store the name before of change it
	cfg.ProjectName = strings.ToLower(cfg.ProjectName)

	dirApp := path.Join(cfg.ProjectName, cfg.PackageName)
	os.MkdirAll(dirApp, PERM_DIRECTORY)

	// === Render project files
	switch cfg.ProjectType {
	case "lib", "cgo":
		renderCode(dirApp+"/Makefile", tmplMakefile, headerMakefile, tag)
		renderCode(dirApp+"/main.go", tmplPkgMain, headerCode, tag)
		renderCode(dirApp+"/main_test.go", tmplTest, headerCode, tag)
	case "app", "tool":
		renderCode(dirApp+"/Makefile", tmplMakefile, headerMakefile, tag)
		renderCode(dirApp+"/main.go", tmplCmdMain, headerCode, tag)
	}

	// === Render common files
	dirTmpl := dirData + "/tmpl" // Templates base directory

	renderFile(cfg.ProjectName, dirTmpl+"/NEWS.mkd", tag)
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
	// File CHANGES is only necessary when is not used a VCS.
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
func updateProject() {
	var err os.Error

	if cfg, err = ReadMetadata(); err != nil {
		log.Exit(err)
	}

	tag, update := tagsToUpdate()
	if *fDebug {
		debug(tag)
	}

	// === Get all Go source files
	finderGo := newFinderGo()
	path.Walk(cfg.PackageName, finderGo, nil)

	if len(finderGo.files) == 0 {
		fmt.Fprintf(os.Stderr,
			"%s: no Go source files in %q\n", argv0, cfg.PackageName)
		os.Exit(ERROR)
	}

	// === Update Makefile and source code files
	bPackageName := []byte(tag["package_name"])

	if update["ProjectName"] || update["License"] || update["PackageInCode"] {
		for _, fname := range finderGo.files {
			file, err := replaceCode(fname, bPackageName, tag, update)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(file))
		}

		// Makefile
		file, err := replaceMakefile(path.Join(cfg.PackageName, "Makefile"),
			bPackageName, tag, update)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(file))
	}

}

/*
	// Rename directory named like the package name.
	if update["PackageName"] {
		if err := os.Rename(cfg.PackageName, *fPackageName); err != nil {
			log.Exit(err)
		}
	}
*/

