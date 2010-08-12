// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Global flag
var fApp = flag.String("app", "pkg", "The application type.")


func loadMetadata() (*metadata, map[string]string) {

	// === Flags for the command line
	// ===

	// Metadata
	var (
		fProjectName = flag.String("Project-name", "",
			"The name of the project.")

		fPackageName = flag.String("Package-name", "",
			"The name of the package.")

		fVersion = flag.String("Version", "",
			"A string containing the package's version number.")

		fSummary = flag.String("Summary", "",
			"A one-line summary of what the package does.")

		fDownloadURL = flag.String("Download-URL", "",
			"A string containing the URL from which this version of the\n"+
				"\tpackage can be downloaded.")

		fAuthor = flag.String("Author", "",
			"A string containing the author's name at a minimum.")

		fAuthorEmail = flag.String("Author-email", "",
			"A string containing the author's e-mail address.")

		fLicense = flag.String("License", "bsd-2",
			"The license covering the package.")

		fPlatform = flag.String("Platform", "",
			"A comma-separated list of platform specifications, summarizing\n"+
				"\tthe operating systems supported by the package which are not listed\n"+
				"\tin the \"Operating System\" Trove classifiers.")

		fDescription = flag.String("Description", "",
			"A longer description of the package that can run to several\n"+
				"\tparagraphs.")

		fKeywords = flag.String("Keywords", "",
			"A list of additional keywords to be used to assist searching for\n"+
				"\tthe package in a larger catalog.")

		fHomePage = flag.String("Home-page", "",
			"A string containing the URL for the package's home page.")

		fClassifier = flag.String("Classifier", "",
			"Each entry is a string giving a single classification value\n"+
				"\tfor the package.")
	)

	// Generic flags
	var (
		fDebug       = flag.Bool("d", false, "debug mode")
		fInteractive = flag.Bool("i", false, "interactive mode")
		fListLicense = flag.Bool("ll", false,
			"shows the list of licenses for the flag `license`")
		fListApp = flag.Bool("la", false,
			"shows the list of application types for the flag `app`")
	)

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

	// Sorted flags for interactive mode
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

	// Available Application types
	var listApp = map[string]string{
		"cmd":    "command line",
		"pkg":    "package",
		"web.go": "web environment",
	}

	// Available licenses
	var listLicense = map[string]string{
		"apache": "Apache (version 2.0)",
		"bsd-2":  "Simplified BSD",
		"bsd-3":  "New BSD",
		"cc0":    "Creative Commons CC0 1.0 Universal",
	}

	// === Parses the flags
	// ===
	usage := func() {
		fmt.Fprintf(os.Stderr, `
Usage: gowizard -app -Project-name -Version -Summary -Download-URL
	-Author -Author-email -License [-Package-name -Platform -Description
	-Keywords -Home-page -Classifier]

`)
		flag.PrintDefaults()
		os.Exit(ERROR)
	}

	flag.Usage = usage
	flag.Parse()

	// === Sets names for both project and package
	setNames := func() {
		reGo := regexp.MustCompile(`^go`) // To remove it from the project name
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
	}

	if *fListLicense {
		fmt.Printf(`
  Licenses
  --------
`)
		for k, v := range listLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
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

	// === Checks
	// ===

	// Necessary fields
	if *fProjectName == "" || *fVersion == "" || *fSummary == "" ||
		*fDownloadURL == "" || *fAuthor == "" || *fAuthorEmail == "" ||
		*fLicense == "" {
		usage()
	}

	// License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	// Application type
	*fApp = strings.ToLower(*fApp)
	if _, present := listApp[*fApp]; !present {
		log.Exitf("unavailable application type %s", *fApp)
	}

	// === Adds the tags to pass to the templates
	// ===
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

	// ===

	classifier := strings.Split(*fClassifier, ",", -1)

	return NewMetadata(*fProjectName, *fPackageName, *fVersion,
		*fSummary, *fDownloadURL, *fAuthor, *fAuthorEmail, *fLicense,
		*fPlatform, *fDescription, *fKeywords, *fHomePage, classifier),
		tag
}

