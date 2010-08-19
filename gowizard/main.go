// Copyright 2010, The "gowizard" Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

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
	var header, tag map[string]string

	cfg, header, tag = loadMetadata()

	// === Creates directories in lower case

	// Gets the data directory from `$(GOROOT)/lib/$(TARG)`
	dirData := path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

	projectName := cfg.ProjectName // Stores the name before of change it
	cfg.ProjectName = strings.ToLower(cfg.ProjectName)

	dirApp := path.Join(cfg.ProjectName, cfg.ApplicationName)
	os.MkdirAll(dirApp, PERM_DIRECTORY)

	// Templates base directory
	dirTmpl := dirData + "/tmpl/pkg"

	// === Renders application files
	switch cfg.ApplicationType {
	case "pkg":
		renderCodeFile(header["makefile"], dirApp, dirTmpl+"/Makefile", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main.go", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main_test.go", tag)
	case "web.go":
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main.go", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main_test.go", tag)

		dirTmpl = dirData + "/tmpl/web.go"
		renderCodeFile(header["makefile"], dirApp, dirTmpl+"/Makefile", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/setup.go", tag)
	case "cmd":
		dirTmpl = dirData + "/tmpl/cmd"
		renderCodeFile(header["makefile"], dirApp, dirTmpl+"/Makefile", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main.go", tag)
		renderCodeFile(header["code"], dirApp, dirTmpl+"/main_test.go", tag)
	}

	// === Renders common files
	dirTmpl = dirData + "/tmpl/common"

	renderFile(cfg.ProjectName, dirTmpl+"/CHANGES.mkd", tag)
	renderFile(cfg.ProjectName, dirTmpl+"/README.mkd", tag)

	if strings.HasPrefix(cfg.License, "cc0") {
		renderNewFile(cfg.ProjectName+"/AUTHORS.mkd",
			dirTmpl+"/AUTHORS-cc0.mkd", tag)
	} else {
		renderFile(cfg.ProjectName, dirTmpl+"/AUTHORS.mkd", tag)
		renderFile(cfg.ProjectName, dirTmpl+"/CONTRIBUTORS.mkd", tag)
	}

	// Adds file related to VCS
	if tag["vcs"] != "" && tag["vcs"] != "n" {
		var fileIgnore string

		switch tag["vcs"] {
		case "git":
			fileIgnore = "gitignore"
		case "hg":
			fileIgnore = "hgignore"
		}

		if err := CopyFile(path.Join(cfg.ProjectName, "."+fileIgnore),
			path.Join(dirTmpl, fileIgnore)); err != nil {
			log.Exit(err)
		}
	}

	// === Adds license file
	dirTmpl = dirData + "/license"

	switch cfg.License {
	case "none":
		break
	case "bsd-3":
		renderNewFile(cfg.ProjectName+"/LICENSE", dirTmpl+"/bsd-3.txt",
			tag)
	default:
		if err := CopyFile(cfg.ProjectName+"/LICENSE",
			path.Join(dirTmpl, cfg.License+".txt")); err != nil {
			log.Exit(err)
		}
	}

	// === Creates file Metadata
	cfg.ProjectName = projectName
	cfg.WriteINI(strings.ToLower(projectName))

	// === Prints messages
	if tag["author_is_org"] != "" {
		if tag["license_is_cc0"] != "" {
			fmt.Print("\n  * Update the file AUTHORS")
		} else {
			fmt.Print("\n  * Update the file CONTRIBUTORS")
		}
	}
	fmt.Println("\n  * Warning: don't edit section 'default' in file Metadata\n")

	os.Exit(0)
}

