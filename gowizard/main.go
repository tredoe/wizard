// Copyright 2010, The "gowizard" Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"strconv"
	"strings"
	"time"
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
	var licenseRender string
	var tag map[string]string

	cfg, tag = loadMetadata()

	// === Renders the header
	if strings.HasPrefix(cfg.License, "cc0") {
		licenseRender = parse(t_LICENSE_CC0, tag)
	} else {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
		licenseRender = parse(t_LICENSE, tag)
	}

	// === Creates directories in lower case

	// Gets the data directory from `$(GOROOT)/lib/$(TARG)`
	dirData := path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

	projectName := cfg.ProjectName // Stores the name before of change it
	cfg.ProjectName = strings.ToLower(cfg.ProjectName)

	dirApp := path.Join(cfg.ProjectName, cfg.ApplicationName)
	os.MkdirAll(dirApp, PERM_DIRECTORY)

	// === Renders application files
	switch cfg.ApplicationType {
	case "pkg":
		renderCodeFile(&licenseRender, dirApp, dirData+"/tmpl/pkg/main.go", tag)
		renderCodeFile(&licenseRender, dirApp, dirData+"/tmpl/pkg/Makefile", tag)
	case "cmd":
		renderCodeFile(&licenseRender, dirApp, dirData+"/tmpl/cmd/main.go", tag)
		renderCodeFile(&licenseRender, dirApp, dirData+"/tmpl/cmd/Makefile", tag)
	case "web.go":
		renderCodeFile(&licenseRender, dirApp, dirData+"/tmpl/web.go/setup.go", tag)
	}

	// === Renders common files
	renderFile(cfg.ProjectName, dirData+"/tmpl/common/AUTHORS.txt", tag)
	renderFile(cfg.ProjectName, dirData+"/tmpl/common/CONTRIBUTORS.txt", tag)
	renderFile(cfg.ProjectName, dirData+"/tmpl/common/README.rst", tag)

	// === Adds license file
	switch cfg.License {
	case "bsd-3":
		renderNewFile(cfg.ProjectName+"/LICENSE.txt",
			dirData+"/license/bsd-3.txt", tag)
	default:
		copy(cfg.ProjectName+"/LICENSE.txt",
			path.Join(dirData, "license", cfg.License+".txt"))
	}

	// === Creates file Metadata
	cfg.ProjectName = projectName
	cfg.WriteINI(strings.ToLower(projectName))

	os.Exit(0)
}

