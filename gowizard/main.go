// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)


// Exit status code if there is any error
const ERROR = 2

// Permissions
const (
	_PERM_DIRECTORY = 0755
	_PERM_FILE      = 0644
)

// Licenses available
var license = map[string]string{
	"apache": "Apache (version 2.0)",
	"bsd-2":  "Simplified BSD",
	"bsd-3":  "New BSD",
	"cc0":    "Creative Commons CC0 1.0 Universal",
}

// Gets the data directory from `$(GOROOT)/lib/$(TARG)`
var dataDir = path.Join(os.Getenv("GOROOT"), "lib", "gowizard")


// === Flags for the command line
// ===

var (
	fDebug = flag.Bool("d", false, "debug mode")
	fList  = flag.Bool("l", false, "show the list of licenses for the flag `license`")
	fWeb   = flag.Bool("w", false, "web application")

	fLicense = flag.String("license", "bsd-2", "kind of license")
)

func checkFlags() {
	usage := func() {
		fmt.Fprintf(os.Stderr,
			"Usage: gowizard -Project-name -Version -Summary -Download-URL -Author\n"+
				"\t\t-Author-email [-Package-name -Platform -Description -Keywords\n"+
				"\t\t-Home-page -Classifier]\n\n")

		flag.PrintDefaults()
		os.Exit(ERROR)
	}
	flag.Usage = usage
	flag.Parse()

	reGo := regexp.MustCompile(`^go`) // To remove it from the project name

	if *fList {
		fmt.Printf("Licenses\n\n")
		for k, v := range license {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	*fLicense = strings.ToLower(*fLicense)
	if _, present := license[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	if *fProjectName == "" {
		usage()
	}
	*fProjectName = strings.TrimSpace(*fProjectName)

	if *fPackageName == "" {
		// The package name is created:
		// getting the last string after of the dash ('-'), if any,
		// and removing 'go'. Finally, it's lower cased.
		pkg := strings.Split(*fProjectName, "-", -1)
		*fPackageName = reGo.ReplaceAllString(strings.ToLower(pkg[len(pkg)-1]), "")
	} else {
		*fPackageName = strings.ToLower(strings.TrimSpace(*fPackageName))
	}

	return
}


// === Main program execution

func main() {
	checkFlags()

	// === Tags to pass to the templates
	line := make([]byte, len(*fProjectName))
	for i, _ := range line {
		line[i] = '='
	}

	tag := map[string]string{
		"license":    license[*fLicense],
		"pkg":        *fPackageName,
		"project":    *fProjectName,
		"headerLine": string(line),
	}

	// === Renders the header
	var licenseRender string

	if *fLicense == "cc0" {
		licenseRender = parse(t_LICENSE_CC0, tag)
	} else {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
		licenseRender = parse(t_LICENSE, tag)
	}

	if *fDebug {
		fmt.Printf("Debug\n\n")
		for k, v := range tag {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	// Creates directories in lower case
	*fProjectName = strings.ToLower(*fProjectName)
	os.MkdirAll(path.Join(*fProjectName, *fPackageName), _PERM_DIRECTORY)

	// Copies license
	copy("/LICENSE.txt", fmt.Sprint("/license/", *fLicense, ".txt"))

	// === Renders common files
	renderFile("/tmpl/common/AUTHORS.txt", tag)
	renderFile("/tmpl/common/README.txt", tag)

	// === Renders files for normal project

	renderCodeFile(&licenseRender, "/tmpl/web.go/setup.go", tag)

	if !*fWeb {

	} else {

	}
}

