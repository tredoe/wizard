// Copyright 2010, The "gowizard" Authors.  All rights reserved.
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
	"strconv"
	"time"

	"github.com/kless/go-readline/readline"
	conf "goconf.googlecode.com/hg"
)

// Global flag
var fUpdate = flag.Bool("u", false, "Updates metadata")


func loadMetadata() (data *metadata, header, tag map[string]string) {

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
		fDebug       = flag.Bool("d", false, "Debug mode")
		fInteractive = flag.Bool("i", false, "Interactive mode")

		fListLicense = flag.Bool("ll", false,
			"Shows the list of licenses for the flag 'License'")
		fListApp = flag.Bool("la", false,
			"Shows the list of application types for the flag 'Application-type'")

		fIsOrganization = flag.Bool("org", false,
			"Does the author is an organization?")
	)

	// Sorted flags for interactive mode
	var interactiveFlags = []string{
		"Application-type",
		"Project-name",
		"Application-name",
		"Author",
		"Author-email",
		"License",
	}

	// === Parses the flags
	// ===
	usage := func() {
		fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-name -Author -Author-email
       gowizard -Project-name -Author [-Author-email] -org
	[-Application-type -Application-name -License]

       gowizard -u [-ProjectName -ApplicationName -License]

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
		os.Exit(0)
	}

	if *fListLicense {
		fmt.Printf(`
  Licenses
  --------
`)
		for k, v := range listLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fInteractive {
		fmt.Printf(`
  Interactive
  -----------
`)
		// === fIsOrganization
		var err os.Error

		f := flag.Lookup("org")
		readline.OptPrompt.Indent = "  "

		*fIsOrganization, err = readline.PromptBool(f.Usage)
		if err != nil {
			log.Exit(err)
		}

		// === Flags for Metadata
		var input string

		for _, k := range interactiveFlags {
			f = flag.Lookup(k)
			text := strings.TrimRight(f.Usage, ".")

			switch k {
			case "Application-type", "License":
				input, err = readline.Prompt(text, f.Value.String())
			case "Application-name":
				setNames()
				input, err = readline.Prompt(text, *fApplicationName)
			case "Author-email":
				if *fIsOrganization {
					input, err = readline.Prompt(text, "")
				} else {
					input, err = readline.RepeatPrompt(text)
				}
			default:
				input, err = readline.RepeatPrompt(text)
			}

			if err != nil {
				log.Exit(err)
			}
			flag.Set(k, input)
		}

		fmt.Println()
	} else {
		setNames()
	}

	// === Checks
	// ===

	// === Necessary fields
	if *fProjectName == "" || *fAuthor == "" {
		usage()
	}
	if *fAuthorEmail == "" && !*fIsOrganization {
		log.Exit("The email address is required for people")
	}

	// === License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		log.Exitf("Unavailable license: '%s'", *fLicense)
	}

	if *fLicense == "bsd-3" && !*fIsOrganization {
		log.Exit("The license 'bsd-3' requires an organization as author")
	}

	// === Application type
	*fApplicationType = strings.ToLower(*fApplicationType)
	if _, present := listApp[*fApplicationType]; !present {
		log.Exitf("Unavailable application type: '%s'", *fApplicationType)
	}

	// === Adds the tags to pass to the templates
	// ===
	projectHeader := make([]byte, len(*fProjectName))
	for i, _ := range projectHeader {
		projectHeader[i] = '='
	}

	var org string
	if *fIsOrganization {
		org = "ok"
	} else {
		org = ""
	}

	tag = map[string]string{
		"is_organization":  org,
		"project_name":     *fProjectName,
		"application_name": *fApplicationName,
		"author":           *fAuthor,
		"author_email":     *fAuthorEmail,
		"license":          listLicense[*fLicense],
		"_project_header":  string(projectHeader),
	}

	// === Adds the headers
	header = renderHeader(tag, fLicense)

	// === Shows data on 'tag' and license header, if 'fDebug' is set
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
		fmt.Printf("\n  header:\n%s\n", header["code"])
		os.Exit(0)
	}

	// ===

	data = NewMetadata(*fProjectName, *fApplicationName, *fApplicationType,
		*fAuthor, *fAuthorEmail, *fLicense, config())

	return data, header, tag
}

/* Returns the INI configuration file. */
func config() (file *conf.ConfigFile) {
	var err os.Error

	if *fUpdate {
		if file, err = conf.ReadConfigFile(_FILE_NAME); err != nil {
			log.Exit(err)
		}
	} else {
		file = conf.NewConfigFile()
	}

	return
}

/* Renders the headers of source code files according to the license. */
func renderHeader(tag map[string]string, fLicense *string) map[string]string {
	const (
		COMMENT_CODE     = "//"
		COMMENT_MAKEFILE = "#"
	)

	var header, headerMakefile, headerCode string
	var isGnu bool

	tag["year"] = strconv.Itoa64(time.LocalTime().Year)

	if strings.HasPrefix(*fLicense, "gpl") || strings.HasPrefix(*fLicense, "agpl") {
		isGnu = true

		tag["version"] = strings.Split(*fLicense, "-", -1)[1]
		header = fmt.Sprint(t_COPYRIGHT, t_LICENSE_LINE_1, t_VERSION_GNU,
			t_LICENSE_LAST_LINE)

		tag["comment"] = "#"
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)

	} else if strings.HasPrefix(*fLicense, "cc0") {
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(t_LICENSE_CC0, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(t_LICENSE_CC0, tag)

	} else if *fLicense == "none" {
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(t_COPYRIGHT, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(t_COPYRIGHT, tag)

	} else {
		header = fmt.Sprint(t_COPYRIGHT, t_LICENSE_LINE_1, t_LICENSE_LAST_LINE)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	}

	// === Now, it's needed without comments to be passed to the README template.
	tag["comment"] = "", false
	tag["copyright"] = parse(t_COPYRIGHT, tag)

	// Adds the version to GNU licenses.
	if isGnu {
		header = fmt.Sprint(tag["license"], t_VERSION_GNU)
		tag["license"] = parse(header, tag)
	}

	 // This tag is not used anymore.
	tag["year"] = "", false

	return map[string]string{
		"makefile": headerMakefile,
		"code":     headerCode,
	}
}


/* TODO

Update
	"project_name"
	"application_name"
	"license"

*/

