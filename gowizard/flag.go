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

	conf "goconf.googlecode.com/hg"
)

// Global flag
var fUpdate = flag.Bool("u", false, "Updates metadata")


func loadMetadata() (*metadata, map[string]string) {

	// === Flags for the command line
	// ===

	// Metadata
	var (
		fProjectName = flag.String("Project-name", "",
			"The name of the project.")

		fApplicationName = flag.String("Application-name", "",
			"The name of the package.")

		fApplicationType = flag.String("Application-type", "pkg",
			"The application type.")

		/*fVersion = flag.String("Version", "",
			"A string containing the package's version number.")

		fSummary = flag.String("Summary", "",
			"A one-line summary of what the package does.")

		fDownloadURL = flag.String("Download-URL", "",
			"A string containing the URL from which this version of the\n"+
				"\tpackage can be downloaded.")*/

		fAuthor = flag.String("Author", "",
			"A string containing the author's name at a minimum.")

		fAuthorEmail = flag.String("Author-email", "",
			"A string containing the author's e-mail address.")

		fLicense = flag.String("License", "bsd-2",
			"The license covering the package.")

		/*fPlatform = flag.String("Platform", "",
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
				"\tfor the package.")*/
	)

	// Generic flags
	var (
		fDebug       = flag.Bool("d", false, "debug mode")
		fInteractive = flag.Bool("i", false, "interactive mode")
		fListLicense = flag.Bool("ll", false,
			"shows the list of licenses for the flag `license`")
		fListApp = flag.Bool("la", false,
			"shows the list of application types for the flag `Application-type`")
	)

	// Flags used on interactive mode
	var interactiveFlags = map[string]*string{
		"Application-type": fApplicationType,
		"Project-name":     fProjectName,
		"Application-name": fApplicationName,
		"Author":           fAuthor,
		"Author-email":     fAuthorEmail,
		"License":          fLicense,
	}

	// Sorted flags for interactive mode
	var sortedInteractiveFlags = []string{
		"Application-type",
		"Project-name",
		"Application-name",
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
Usage: gowizard -Application-type -Project-name -Author -Author-email -License
	[-Application-name]

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

		switch *fApplicationType {
		// The name of a tool for the command line is usually named as
		// the project name.
		case "cmd":
			if *fApplicationName == "" {
				*fApplicationName = strings.ToLower(*fProjectName)
			}
		default:
			if *fApplicationName == "" {
				// The package name is created:
				// getting the last string after of the dash ('-'), if any,
				// and removing 'go'. Finally, it's lower cased.
				pkg := strings.Split(*fProjectName, "-", -1)
				*fApplicationName = reGo.ReplaceAllString(
					strings.ToLower(pkg[len(pkg)-1]), "")
			} else {
				*fApplicationName = strings.ToLower(
					strings.TrimSpace(*fApplicationName))
			}
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
		for _, k := range sortedInteractiveFlags {
			f := flag.Lookup(k)
			fmt.Printf("  %s", strings.TrimRight(f.Usage, "."))

			switch k {
			case "Application-type", "License":
				fmt.Printf(" [%s]", f.Value)
			case "Application-name":
				setNames()
				fmt.Printf(" [%s]", *fApplicationName)
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
	if *fProjectName == "" || *fAuthor == "" || *fAuthorEmail == "" ||
		*fLicense == "" {
		usage()
	}

	// License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	// Application type
	*fApplicationType = strings.ToLower(*fApplicationType)
	if _, present := listApp[*fApplicationType]; !present {
		log.Exitf("unavailable application type %s", *fApplicationType)
	}

	// === Adds the tags to pass to the templates
	// ===
	projectHeader := make([]byte, len(*fProjectName))
	for i, _ := range projectHeader {
		projectHeader[i] = '='
	}

	tag := map[string]string{
		"project_name":     *fProjectName,
		"application_name": *fApplicationName,
		"full_author":      fmt.Sprint(*fAuthor, " <", *fAuthorEmail, ">"),
		"license":          listLicense[*fLicense],
		"_project_header":  string(projectHeader),
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

	var file *conf.ConfigFile
	var err os.Error

	if *fUpdate {
		if file, err = conf.ReadConfigFile(_FILE_NAME); err != nil {
			log.Exit(err)
		}
	} else {
		file = conf.NewConfigFile()
	}

	return NewMetadata(*fProjectName, *fApplicationName, *fApplicationType,
		*fAuthor, *fAuthorEmail, *fLicense, file),
		tag
}

