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
	fListLicense = flag.Bool("ll", false,
		"shows the list of licenses for the flag `license`")
	fListApp = flag.Bool("la", false,
		"shows the list of application types for the flag `app`")

	fApp = flag.String("app", "pkg", "The application type.")
)

// Available application types
var listApp = map[string]string{
	"cmd":    "command line",
	"pkg":    "package",
	"web.go": "web environment",
}

// Licenses available
var listLicense = map[string]string{
	"apache": "Apache (version 2.0)",
	"bsd-2":  "Simplified BSD",
	"bsd-3":  "New BSD",
	"cc0":    "Creative Commons CC0 1.0 Universal",
}

// Flags used in interactive mode
var interactiveFlags = map[string]*string{
	"app":          fApp,
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
	"app",
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
		fmt.Fprintf(os.Stderr, `
Usage: gowizard -app -Project-name -Version -Summary -Download-URL
	-Author -Author-email -License [-Package-name -Platform -Description
	-Keywords -Home-page -Classifier]

`)

		flag.PrintDefaults()
		os.Exit(_ERROR)
	}

	flag.Usage = usage
	flag.Parse()

	// === Options
	// ===

	if *fListApp {
		fmt.Printf(`
  Applications
  ------------
`)
		for k, v := range listApp {
			fmt.Printf("  %s: %s\n", k, v)
		}
		//os.Exit(0)
	}

	if *fListLicense {
		fmt.Printf(`
  Licenses
  --------
`)
		for k, v := range listLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
		//os.Exit(0)
	}

	if *fInteractive {
		fmt.Printf(`
  Interactive
  -----------
`)

		for i, k := range sortedInteractiveFlags {
			f := flag.Lookup(k)
			fmt.Printf("  %s", strings.TrimRight(f.Usage, "."))

			switch i {
			case 0, 8: // app, License
				fmt.Printf(" [%s]", f.Value)
			case 2: // Package-name
				setNames()
				fmt.Printf(" [%s]", *fPackageName)
			}

			fmt.Print(": ")

			if input := read(); input != "" {
				*interactiveFlags[k] = input
			}
		}

		fmt.Println()
	} else {
		setNames()
	}

	// === Checks necessary fields
	if *fProjectName == "" || *fVersion == "" || *fSummary == "" ||
		*fDownloadURL == "" || *fAuthor == "" || *fAuthorEmail == "" ||
		*fLicense == "" {
		usage()
	}

	// === Checks license
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	// === Checks application type
	*fApp = strings.ToLower(*fApp)
	if _, present := listApp[*fApp]; !present {
		log.Exitf("unavailable application type %s", *fApp)
	}

	return
}

/* Sets names for both project and package. */
func setNames() {

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
}


// === Main program execution

func main() {
	checkFlags()

	// === Tags to pass to the templates
	projectNameHeader := make([]byte, len(*fProjectName))
	for i, _ := range projectNameHeader {
		projectNameHeader[i] = '='
	}

	tag := map[string]string{
		"projectName":        *fProjectName,
		"packageName":        *fPackageName,
		"author":             fmt.Sprint(*fAuthor, " <", *fAuthorEmail, ">"),
		"license":            listLicense[*fLicense],
		"_projectNameHeader": string(projectNameHeader),
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
		fmt.Printf(`
  Debug
  -----
`)
		for k, v := range tag {
			if k[0] == '_' {
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

	// === Creates Metadata file
	metadata := newMetadata_1_1(*fProjectName, *fPackageName, *fVersion,
		*fSummary, *fDownloadURL, *fAuthor, *fAuthorEmail, tag["license"])
	metadata.writeJSON(*fProjectName)

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

