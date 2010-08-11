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
const _ERROR = 2

// Permissions
const (
	_PERM_DIRECTORY = 0755
	_PERM_FILE      = 0644
)

// Gets the data directory from `$(GOROOT)/lib/$(TARG)`
var dataDir = path.Join(os.Getenv("GOROOT"), "lib", "gowizard")


// === Flags for the command line
// ===

// To remove it from the project name
var reGo = regexp.MustCompile(`^go`)

var (
	fDebug       = flag.Bool("d", false, "debug mode")
	fInteractive = flag.Bool("i", false, "interactive mode")
	fList        = flag.Bool("l", false,
		"shows the list of licenses for the flag `license`")
	fWeb = flag.Bool("w", false, "web application")
)

// Flags used in interactive mode
var interactiveFlags = map[string]*string{
	"Project-name": fProjectName,
	"Package-name": fPackageName,
	"Version":      fVersion,
	"Summary":      fSummary,
	"Download-URL": fDownloadURL,
	"Author":       fAuthor,
	"Author-email": fAuthorEmail,
	"License":      fLicense,
}

// Sorted flags
var sortedInteractiveFlags = []string{
	"Project-name",
	"Package-name",
	"Version",
	"Summary",
	"Download-URL",
	"Author",
	"Author-email",
	"License",
}

func checkFlags() {
	usage := func() {
		fmt.Fprintf(os.Stderr,
			"Usage: gowizard -Project-name -Version -Summary -Download-URL -Author\n"+
				"\t\t-Author-email -License [-Package-name -Platform -Description\n"+
				"\t\t-Keywords -Home-page -Classifier]\n\n")

		flag.PrintDefaults()
		os.Exit(_ERROR)
	}

	flag.Usage = usage
	flag.Parse()

	// === Options
	if *fList {
		fmt.Printf("Licenses\n\n")
		for k, v := range license {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fInteractive {
		for _, k := range sortedInteractiveFlags {
			f := flag.Lookup(k)

			fmt.Printf("\n  %s: ", strings.TrimRight(f.Usage, "."))
			*interactiveFlags[k] = read()
		}

		fmt.Println()
	}

	// === Checks necessary fields
	if *fProjectName == "" || *fVersion == "" || *fSummary == "" ||
		*fDownloadURL == "" || *fAuthor == "" || *fAuthorEmail == "" ||
		*fLicense == "" {
		usage()
	}

	// === Checks license
	*fLicense = strings.ToLower(*fLicense)
	if _, present := license[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	// === Sets names for both project and package
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
		"project":    *fProjectName,
		"pkg":        *fPackageName,
		"license":    license[*fLicense],
		"headerLine": string(line),
		"author":     fmt.Sprint(*fAuthor, " <", *fAuthorEmail, ">"),
	}

	// === Renders the header
	var licenseRender string

	if *fLicense == "cc0" {
		licenseRender = parse(t_LICENSE_CC0, tag)
	} else {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
		licenseRender = parse(t_LICENSE, tag)
	}

	// === Shows data on 'tag', if 'fDebug' is set
	if *fDebug {
		fmt.Printf("Debug\n\n")
		for k, v := range tag {
			if k == "headerLine" {
				continue
			}
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	// === Creates directories in lower case
	*fProjectName = strings.ToLower(*fProjectName)
	os.MkdirAll(path.Join(*fProjectName, *fPackageName), _PERM_DIRECTORY)

	// === Copies the license
	copy(*fProjectName+"/LICENSE.txt",
		fmt.Sprint(dataDir, "/license/", *fLicense, ".txt"))

	// === Renders common files
	renderFile(dataDir+"/tmpl/common/AUTHORS.txt", tag)
	renderFile(dataDir+"/tmpl/common/CONTRIBUTORS.txt", tag)
	renderFile(dataDir+"/tmpl/common/README.txt", tag)

	// === Renders source code files
	if *fWeb {
		renderCodeFile(&licenseRender, dataDir+"/tmpl/web.go/setup.go", tag)
	} else {

	}

	os.Exit(0)
}

