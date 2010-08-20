// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for more details.

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

		/*fDownloadURL = flag.String("Download-URL", "",
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
		fListVCS = flag.Bool("lv", false,
			"Shows the list of version control systems")

		fAuthorIsOrg = flag.Bool("org", false,
			"Does the author is an organization?")

		fVCS = flag.String("vcs", "none", "Version control system")
	)

	// Sorted flags for interactive mode
	var interactiveFlags = []string{
		"org",
		"Application-type",
		"Project-name",
		"Application-name",
		"Author",
		"Author-email",
		"License",
		"vcs",
	}

	// Available version control systems
	var listVCS = map[string]string{
		"git":  "Git",
		"hg":   "Mercurial",
		"none": "other/none",
	}

	// === Parses the flags
	// ===
	usage := func() {
		fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-name -Author -Author-email
       gowizard -Project-name -Author [-Author-email] -org
	[-Application-type -Application-name -License -vcs]

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
		fmt.Println("  = Application types\n")
		for k, v := range listApp {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fListLicense {
		fmt.Println("  = Licenses\n")
		for k, v := range listLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fListVCS {
		fmt.Println("  = Version control systems\n")
		for k, v := range listVCS {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fInteractive {
		var input string
		var err os.Error

		fmt.Println("  = Interactive\n")
		readline.OptPrompt.Indent = "  "

		for _, k := range interactiveFlags {
			f := flag.Lookup(k)
			text := fmt.Sprintf("+ %s", strings.TrimRight(f.Usage, "."))

			switch k {
			case "Application-name":
				setNames()
				input, err = readline.Prompt(text, *fApplicationName)
			case "Application-type":
				input, err = readline.PromptChoice(text, arrayKeys(listApp),
					f.Value.String())
			case "Author-email":
				if *fAuthorIsOrg {
					input, err = readline.Prompt(text, "")
				} else {
					input, err = readline.RepeatPrompt(text)
				}
			case "License":
				input, err = readline.PromptChoice(text, arrayKeys(listLicense),
					f.Value.String())
			case "org":
				*fAuthorIsOrg, err = readline.PromptBool(text)
			case "vcs":
				input, err = readline.PromptChoice(text, arrayKeys(listVCS),
					f.Value.String())
			default:
				input, err = readline.RepeatPrompt(text)
			}

			if err != nil {
				log.Exit(err)
			}

			if k != "org" {
				flag.Set(k, input)
			}
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
	if *fAuthorEmail == "" && !*fAuthorIsOrg {
		log.Exit("The email address is required for people")
	}

	// === License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		log.Exitf("Unavailable license: %q", *fLicense)
	}

	if *fLicense == "bsd-3" && !*fAuthorIsOrg {
		log.Exit("The license 'bsd-3' requires an organization as author")
	}

	// === Application type
	*fApplicationType = strings.ToLower(*fApplicationType)
	if _, present := listApp[*fApplicationType]; !present {
		log.Exitf("Unavailable application type: %q", *fApplicationType)
	}

	// === VCS
	*fVCS = strings.ToLower(*fVCS)
	if *fVCS != "none" {
		if _, present := listVCS[*fVCS]; !present {
			log.Exitf("Unavailable version control system: %q", *fVCS)
		}
	}

	// === Adds the tags to pass to the templates
	// ===
	var value string

	tag = map[string]string{
		"application_name": *fApplicationName,
		"author":           *fAuthor,
		"author_email":     *fAuthorEmail,
		"license":          listLicense[*fLicense],
		"project_name":     *fProjectName,
		"vcs":              *fVCS,
	}

	projectHeader := make([]byte, len(*fProjectName))
	for i, _ := range projectHeader {
		projectHeader[i] = '='
	}
	tag["_project_header"] = string(projectHeader)

	if *fAuthorIsOrg {
		value = "ok"
	} else {
		value = ""
	}
	tag["author_is_org"] = value

	if *fApplicationType == "pkg" {
		value = "ok"
	} else {
		value = ""
	}
	tag["app_is_pkg"] = value

	if strings.HasPrefix(*fLicense, "cc0") {
		value = "ok"
	} else {
		value = ""
	}
	tag["license_is_cc0"] = value

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

/* Gets an array from the map keys. */
func arrayKeys(m map[string]string) []string {
	a := make([]string, len(m))

	i := 0
	for k, _ := range m {
		a[i] = k
		i++
	}

	return a
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

	var headerMakefile, headerCode string

	tag["year"] = strconv.Itoa64(time.LocalTime().Year)
	licenseName := strings.Split(*fLicense, "-", -1)[0]

	switch licenseName {
	case "apache":
		header := fmt.Sprint(t_COPYRIGHT, t_APACHE)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "bsd":
		header := fmt.Sprint(t_COPYRIGHT, t_BSD)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "cc0":
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(t_CC0, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(t_CC0, tag)
	case "gpl", "agpl":
		header := fmt.Sprint(t_COPYRIGHT, t_GNU)
		if licenseName == "agpl" {
			tag["Affero"] = "Affero"
		} else {
			tag["Affero"] = ""
		}

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "none":
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(t_COPYRIGHT, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(t_COPYRIGHT, tag)
	}

	// Tag to render the copyright in README.
	tag["comment"] = ""
	tag["copyright"] = parse(t_COPYRIGHT, tag)

	// These tags are not used anymore.
	for _, t := range []string{"Affero", "comment", "year"} {
		tag[t] = "", false
	}

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

