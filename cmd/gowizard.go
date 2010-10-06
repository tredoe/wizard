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
	"container/vector"
	"fmt"
	"io/ioutil"
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

// Characters
const (
	CHAR_CODE_COMMENT = "//" // For comments in source code files
	CHAR_MAKE_COMMENT = "#"  // For comments in file Makefile
	CHAR_HEADER       = '='  // Header under the project name
)

const ERROR = 2 // Exit status code if there is any error
const README = "README.mkd"

// Get data directory from `$(GOROOT)/lib/$(TARG)`
var dirData = path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

var argv0 = os.Args[0] // Executable name


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

/* Add license file in directory `dir`. */
func addLicense(dir string, tag map[string]string) {
	dirTmpl := dirData + "/license"

	switch *fLicense {
	case "none":
		break
	case "bsd-3":
		renderNewFile(dir+"/LICENSE", dirTmpl+"/bsd-3.txt", tag)
	default:
		if err := copyFile(dir+"/LICENSE",
			path.Join(dirTmpl, *fLicense+".txt")); err != nil {
			log.Exit(err)
		}
	}
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

	headerCodeFile, headerMakefile := renderAllHeaders(tag, "")

	// === Render project files
	var dirApp string // To create directories in lower case.

	switch *fProjecType {
	case "lib", "cgo":
		dirApp = path.Join(*fProjectName, *fPackageName)
		os.MkdirAll(dirApp, PERM_DIRECTORY)

		renderNesting(dirApp+"/main.go", headerCodeFile, tmplPkgMain, tag)
		renderNesting(dirApp+"/main_test.go", headerCodeFile, tmplTest, tag)
		renderNesting(dirApp+"/Makefile", headerMakefile, tmplPkgMakefile, tag)
	case "app", "tool":
		dirApp = path.Join(*fProjectName, "cmd")
		os.MkdirAll(dirApp, PERM_DIRECTORY)

		renderNesting(path.Join(dirApp, *fPackageName)+".go",
			headerCodeFile, tmplCmdMain, tag)
		renderNesting(dirApp+"/Makefile", headerMakefile, tmplCmdMakefile, tag)
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
	case "other":
		break
	// File CHANGES is only necessary when is not used a VCS.
	case "none":
		renderFile(*fProjectName, dirTmpl+"/CHANGES.mkd", tag)
	default:
		ignoreFile := "." + *fVCS + "ignore"

		if *fVCS == "hg" {
			tmplIgnore = hgIgnoreTop + tmplIgnore
		}

		if err := ioutil.WriteFile(path.Join(*fProjectName, ignoreFile),
		[]byte(tmplIgnore), PERM_FILE); err != nil {
			log.Exit(err)
		}
	}

	// === License file
	addLicense(*fProjectName, tag)

	// === Create file Metadata
	// tag["project_name"] has the original name (no in lower case).
	cfg := NewMetadata(*fProjecType, tag["project_name"], *fPackageName,
		*fLicense, *fAuthor, *fAuthorEmail, *fVCS)

	if err := cfg.WriteINI(*fProjectName); err != nil {
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
	var updatedFiles, errorFiles vector.StringVector

	// VCS configuration files to push to a server.
	var configVCS = map[string]string{
		"bzr": ".bzr/branch/branch.conf",
		"git": ".git/config",
		"hg":  ".hg/hgrc",
	}

	// 'cfg' has the old values.
	cfg, err := ReadMetadata()
	if err != nil {
		log.Exit(err)
	}

	// 'tag' and the flags have the new values.
	tag, update := tagsToUpdate(cfg)
	if *fDebug {
		debug(tag)
	}

	// === Rename directories
	if *fVerbose && (update["ProjectName"] || update["PackageName"]) {
		fmt.Println("  = Directories renamed\n")
	}

	if update["ProjectName"] {
		if err := os.Chdir(".."); err != nil {
			log.Exit(err)
		}

		oldProjectName := strings.ToLower(cfg.ProjectName)

		if err := os.Rename(oldProjectName, *fProjectName); err != nil {
			log.Exit(err)
		} else if *fVerbose {
			fmt.Printf(" * Project: %q -> %q\n", oldProjectName, *fProjectName)
		}

		// Do 'chdir' in new project directory.
		if err := os.Chdir(*fProjectName); err != nil {
			log.Exit(err)
		}

		// === Rename URL in the VCS
		fname := configVCS[cfg.VCS]

		if cfg.VCS != "other" && cfg.VCS != "none" && backup(fname) {
			if err := replaceVCS_URL(fname, strings.ToLower(cfg.ProjectName),
			*fProjectName, cfg.VCS); err != nil {
				log.Exit(err)
			}

			if *fVerbose {
				updatedFiles.Push(fname)
			}
		}
	}

	if update["PackageName"] {
		if err := os.Rename(cfg.PackageName, *fPackageName); err != nil {
			log.Exit(err)
		} else if *fVerbose {
			fmt.Printf(" * Package: %q -> %q\n", cfg.PackageName, *fPackageName)
		}
	}

	// === Update source code files
	if update["ProjectName"] || update["License"] || update["PackageInCode"] {
		packageName := []byte(tag["package_name"])
		files := finderGo(*fPackageName)

		for _, fname := range files {
			if backup(fname) {

				if err := replaceGoFile(
				fname, packageName, cfg, tag, update); err != nil {
					fmt.Fprintf(os.Stderr,
						"%s: file %q not updated: %s\n", argv0, fname, err)
				} else if *fVerbose {
					updatedFiles.Push(fname)
				}
			} else {
				errorFiles.Push(fname)
			}
		}

		// === Update Makefile
		fname := path.Join(*fPackageName, "Makefile")

		if backup(fname) {
			if err := replaceMakefile(
			fname, packageName, cfg, tag, update); err != nil {
				fmt.Fprintf(os.Stderr,
					"%s: file %q not updated: %s\n", argv0, fname, err)
			} else if *fVerbose {
				updatedFiles.Push(fname)
			}
		} else {
			errorFiles.Push(fname)
		}
	}

	// === Update text files with extension 'mkd'
	if update["ProjectName"] || update["License"] {
		projectName := []byte(tag["project_name"])
		files := finderMkd(".")

		for _, fname := range files {
			if backup(fname) {

				if err := replaceTextFile(
				fname, projectName, cfg, tag, update); err != nil {
					fmt.Fprintf(os.Stderr,
						"%s: file %q not updated: %s\n", argv0, fname, err)
				} else if *fVerbose {
					updatedFiles.Push(fname)
				}
			} else {
				errorFiles.Push(fname)
			}
		}
	}

	// === License file
	if update["License"] {
		addLicense(".", tag)

		if *fVerbose {
			updatedFiles.Push("LICENSE")
		}

		cfg.License = *fLicense // Metadata
	}

	// === Metadata file
	if backup(_META_FILE) {
		if update["ProjectName"] {
			cfg.ProjectName = tag["project_name"]
		}
		if update["PackageName"] {
			cfg.PackageName = *fPackageName
		}

		if err := cfg.WriteINI("."); err != nil {
			log.Exit(err)
		}
	} else {
		errorFiles.Push(_META_FILE)
	}

	// === Print messages
	if *fVerbose {
		updatedFiles.Push(_META_FILE)
		fmt.Println("\n  = Files updated\n")

		for _, file := range updatedFiles {
			fmt.Printf(" * %s\n", file)
		}
	}

	if len(errorFiles) != 0 {
		files := ""

		for i, file := range errorFiles {
			if i == 0 {
				files = file
			} else {
				files += ", " + file
			}
		}

		fmt.Fprintf(os.Stderr, "%s: could not be backed up: %s\n", argv0, files)
	}
}

