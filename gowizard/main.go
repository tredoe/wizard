// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
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

// Gets the data directory from `$(GOROOT)/lib/$(TARG)`
var dataDir = path.Join(os.Getenv("GOROOT"), "lib", "gowizard")

// Metadata to build the new project
var cfg *metadata


// === Main program execution

func main() {
	var tag map[string]string

	cfg, tag = loadMetadata()

	// === Renders the header
	var licenseRender string

	if cfg.License == "cc0" {
		licenseRender = parse(t_LICENSE_CC0, tag)
	} else {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
		licenseRender = parse(t_LICENSE, tag)
	}

	// === Creates directories in lower case
	cfg.ProjectName = strings.ToLower(cfg.ProjectName)
	os.MkdirAll(path.Join(cfg.ProjectName, cfg.PackageName), PERM_DIRECTORY)

	// === Copies the license
	copy(cfg.ProjectName+"/LICENSE.txt",
		path.Join(dataDir, "license", cfg.License+".txt"))

	// === Renders common files
	renderFile(dataDir+"/tmpl/common/AUTHORS.txt", tag)
	renderFile(dataDir+"/tmpl/common/CONTRIBUTORS.txt", tag)
	renderFile(dataDir+"/tmpl/common/README.txt", tag)

	// === Creates Metadata file
	cfg.License = tag["license"]
	cfg.writeJSON(cfg.ProjectName)

	// === Renders source code files
	switch *fApp {
	case "pkg":
		renderCodeFile(&licenseRender, dataDir+"/tmpl/pkg/main.go", tag)
		renderCodeFile(&licenseRender, dataDir+"/tmpl/pkg/Makefile", tag)
	case "cmd":
		renderCodeFile(&licenseRender, dataDir+"/tmpl/cmd/main.go", tag)
		renderCodeFile(&licenseRender, dataDir+"/tmpl/cmd/Makefile", tag)
	case "web.go":
		renderCodeFile(&licenseRender, dataDir+"/tmpl/web.go/setup.go", tag)
	}

	os.Exit(0)
}

