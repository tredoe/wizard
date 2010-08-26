// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

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

	"github.com/kless/go-readin/readin"
)

// Global flag
var fUpdate = flag.Bool("u", false, "Updates metadata")


func loadMetadata() (data *metadata, header, tag map[string]string) {

	// === Flags for the command line
	// ===

	// Metadata
	var (
		fProjecType = flag.String("Project-type", "",
			"The project type.")

		fProjectName = flag.String("Project-name", "",
			"The name of the project.")

		fPackageName = flag.String("Package-name", "",
			"The name of the package.")

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
		fListProject = flag.Bool("lp", false,
			"Shows the list of project types for the flag 'Project-type'")
		fListVCS = flag.Bool("lv", false,
			"Shows the list of version control systems")

		fAuthorIsOrg = flag.Bool("org", false,
			"Does the author is an organization?")

		fVCS = flag.String("vcs", "", "Version control system")
	)

	// Sorted flags for interactive mode
	var interactiveFlags = []string{
		"org",
		"Project-type",
		"Project-name",
		"Package-name",
		"Author",
		"Author-email",
		"License",
		"vcs",
	}

	// Available version control systems
	var listVCS = map[string]string{
		"git":   "Git",
		"hg":    "Mercurial",
		"other": "other VCS",
		"none":  "none",
	}

	// === Parse the flags
	// ===
	usage := func() {
		fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-type -Project-name -Author [-Author-email] -vcs
	[-Package-name -License -org]

       gowizard -u [-Project-name -Package-name -License]

`)
		flag.PrintDefaults()
		os.Exit(ERROR)
	}

	flag.Usage = usage
	flag.Parse()

	// === Set names for both project and package
	setNames := func() {
		reGo := regexp.MustCompile(`^go`) // To remove it from the project name
		*fProjectName = strings.TrimSpace(*fProjectName)

		switch *fProjecType {
		// A program is usually named as the project name.
		case "app", "tool":
			if *fPackageName == "" {
				*fPackageName = strings.ToLower(*fProjectName)
			}
		default:
			if *fPackageName == "" {
				// The package name is created:
				// getting the last string after of the dash ('-'), if any,
				// and removing 'go'. Finally, it's lower cased.
				pkg := strings.Split(*fProjectName, "-", -1)
				*fPackageName = reGo.ReplaceAllString(
					strings.ToLower(pkg[len(pkg)-1]), "")
			} else {
				*fPackageName = strings.ToLower(
					strings.TrimSpace(*fPackageName))
			}
		}
	}

	// === Options
	// ===

	if *fListProject {
		fmt.Println("  = Project types\n")
		for k, v := range listProject {
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
		readin.DefaultIndent = "  "

		for _, k := range interactiveFlags {
			f := flag.Lookup(k)
			text := fmt.Sprintf("+ %s", strings.TrimRight(f.Usage, "."))

			switch k {
			case "Package-name":
				setNames()
				input, err = readin.Prompt(text, *fPackageName)
			case "Project-type":
				input, err = readin.PromptChoice(text, arrayKeys(listProject),
					f.Value.String())
			case "Author-email":
				if *fAuthorIsOrg {
					input, err = readin.Prompt(text, "")
				} else {
					input, err = readin.RepeatPrompt(text)
				}
			case "License":
				input, err = readin.PromptChoice(text, arrayKeys(listLicense),
					f.Value.String())
			case "org":
				*fAuthorIsOrg, err = readin.PromptBool(text)
			case "vcs":
				input, err = readin.PromptChoice(text, arrayKeys(listVCS),
					f.Value.String())
			default:
				input, err = readin.RepeatPrompt(text)
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

	// === Checking
	// ===

	// === Necessary fields
	if *fProjecType == "" || *fProjectName == "" || *fAuthor == "" || *fVCS == "" {
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

	// === Project type
	*fProjecType = strings.ToLower(*fProjecType)
	if _, present := listProject[*fProjecType]; !present {
		log.Exitf("Unavailable project type: %q", *fProjecType)
	}

	// === VCS
	*fVCS = strings.ToLower(*fVCS)
	if _, present := listVCS[*fVCS]; !present {
		log.Exitf("Unavailable version control system: %q", *fVCS)
	}

	// === Add the tags to pass to the templates
	// ===
	var value string

	tag = map[string]string{
		"project_name": *fProjectName,
		"package_name": *fPackageName,
		"author":       *fAuthor,
		"author_email": *fAuthorEmail,
		"license":      listLicense[*fLicense],
		"vcs":          *fVCS,
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

	if *fProjecType == "cgo" {
		value = "ok"
	} else {
		value = ""
	}
	tag["project_is_cgo"] = value

	if *fProjecType == "lib" {
		value = "ok"
	} else {
		value = ""
	}
	tag["project_is_lib"] = value

	if strings.HasPrefix(*fLicense, "cc0") {
		value = "ok"
	} else {
		value = ""
	}
	tag["license_is_cc0"] = value

	if *fVCS == "none" {
		value = "ok"
	} else {
		value = ""
	}
	tag["vcs_is_none"] = value

	// === Add headers
	header = renderHeader(tag, fLicense)

	// === Show data on 'tag' and license header, if 'fDebug' is set
	if *fDebug {
		fmt.Println("  = Debug\n")

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

	data = NewMetadata(*fProjecType, *fProjectName, *fPackageName,
		*fAuthor, *fAuthorEmail, *fLicense, configFile())

	return data, header, tag
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
		header := fmt.Sprint(tmplCopyright, tmplApache)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "bsd":
		header := fmt.Sprint(tmplCopyright, tmplBSD)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "cc0":
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(tmplCC0, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(tmplCC0, tag)
	case "gpl", "agpl":
		header := fmt.Sprint(tmplCopyright, tmplGNU)
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
		header := fmt.Sprint(tmplCopyright, "\n")

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	}

	// Tag to render the copyright in README.
	tag["comment"] = ""
	tag["copyright"] = parse(tmplCopyright, tag)

	// These tags are not used anymore.
	for _, t := range []string{"Affero", "comment", "year"} {
		tag[t] = "", false
	}

	return map[string]string{
		"makefile": headerMakefile,
		"code":     headerCode,
	}
}

