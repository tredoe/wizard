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

	"github.com/kless/go-readin/readin"
)


// Flags for the command line
var (
	// === Metadata
	fProjecType  = flag.String("Project-type", "", "The project type.")
	fProjectName = flag.String("Project-name", "", "The name of the project.")
	fPackageName = flag.String("Package-name", "", "The name of the package.")
	fLicense     = flag.String("License", "", "The license covering the package.")
	fAuthor      = flag.String("Author", "",
		"A string containing the author's name at a minimucfg.")
	fAuthorEmail = flag.String("Author-email", "",
		"A string containing the author's e-mail address.")

	fAuthorIsOrg = flag.Bool("org", false, "Does the author is an organization?")
	fDebug       = flag.Bool("d", false, "Debug mode")
	fUpdate      = flag.Bool("u", false, "Updates")
	fVCS         = flag.String("vcs", "", "Version control system")
)

// Available version control systems
var listVCS = map[string]string{
	"git":   "Git",
	"hg":    "Mercurial",
	"other": "other VCS",
	"none":  "none",
}


func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-type -Project-name -License -Author -Author-email -vcs
	[-Package-name -org]

       gowizard -u [-Project-name -Package-name -License]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

/* Loads configuration from flags.

Return tags for templates.
*/
func loadConfig() {
	// Generic flags
	var (
		fInteractive = flag.Bool("i", false, "Interactive mode")

		fListLicense = flag.Bool("ll", false,
			"Shows the list of licenses for the flag 'License'")
		fListProject = flag.Bool("lp", false,
			"Shows the list of project types for the flag 'Project-type'")
		fListVCS = flag.Bool("lv", false,
			"Shows the list of version control systems")
	)

	// === Parse the flags
	flag.Usage = usage
	flag.Parse()

	if len(os.Args) == 1 {
		usage()
	}

	// === Options
	if *fListProject {
		fmt.Println("  = Project types\n")
		for k, v := range listProject {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if *fListLicense {
		fmt.Println("  = Licenses\n")
		for k, v := range listLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if *fListVCS {
		fmt.Println("  = Version control systems\n")
		for k, v := range listVCS {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if !*fUpdate {
		if *fInteractive {
			interactive()
		} else {
			setNames()
		}
	}

	// === Checking and add tags
	if !*fUpdate {
		checkAtCreate()
		//tag = tagsToCreate()
	} else {
		checkAtUpdate()
		//tag = tagsToUpdate()
	}

	//return tag
}


// === Checking
// ===

/* Common checking. */
func checkCommon(errors bool) {
	// === License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		fmt.Fprintf(os.Stderr,
			"%s: unavailable license: %q\n", argv0, *fLicense)
		errors = true
	}

	if *fLicense == "bsd-3" && !*fAuthorIsOrg {
		fmt.Fprintf(os.Stderr,
			"%s: license 'bsd-3' requires an organization as author\n", argv0)
		errors = true
	}

	if errors {
		os.Exit(ERROR)
	}
}

/* Checking at create project. */
func checkAtCreate() {
	var errors bool

	// === Necessary fields
	if *fProjecType == "" || *fProjectName == "" || *fLicense == "" ||
		*fAuthor == "" || *fVCS == "" {
		fmt.Fprintf(os.Stderr,
			"%s: missing required fields to create project\n", argv0)
		usage()
	}
	if *fAuthorEmail == "" && !*fAuthorIsOrg {
		fmt.Fprintf(os.Stderr,
			"%s: the email address is required for people\n", argv0)
		errors = true
	}

	// === Project type
	*fProjecType = strings.ToLower(*fProjecType)
	if _, present := listProject[*fProjecType]; !present {
		fmt.Fprintf(os.Stderr,
			"%s: unavailable project type: %q\n", argv0, *fProjecType)
		errors = true
	}

	// === VCS
	*fVCS = strings.ToLower(*fVCS)
	if _, present := listVCS[*fVCS]; !present {
		fmt.Fprintf(os.Stderr,
			"%s: unavailable version control system: %q\n", argv0, *fVCS)
		errors = true
	}

	checkCommon(errors)
}

/* Checking at update project. */
func checkAtUpdate() {
	var errors bool

	// === Necessary fields
	if *fProjectName == "" && *fPackageName == "" && *fLicense == "" {
		fmt.Fprintf(os.Stderr,
			"%s: missing required fields to update\n", argv0)
		usage()
	}

	if *fLicense != "" {
		checkCommon(errors)
	}
}

// ===


/* Interactive mode. */
func interactive() {
	var input string
	var err os.Error

	// Sorted flags
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
}

/* Set names for both project and package. */
func setNames() {
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

/* Create tags to pass to the templates. Used at creating new project. */
func tagsToCreate() map[string]string {
	var value string

	tag := map[string]string{
		"project_name":    *fProjectName,
		"package_name":    *fPackageName,
		"author":          *fAuthor,
		"author_email":    *fAuthorEmail,
		"license":         listLicense[*fLicense],
		"vcs":             *fVCS,
		"_project_header": header(*fProjectName),
	}

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
	tag["project_is_lib"] = value

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

	return tag
}

/* Create tags to pass to the templates. Used at updating a project.

If the flags are empty then they are set with metadata values.
If flags are not empty then is indicated in map 'update' since they are the
values to change.
*/
func tagsToUpdate() (tag map[string]string, update map[string]bool) {
	var value string

	tag = map[string]string{
		"project_name": *fProjectName,
		"package_name": *fPackageName,
	}
	update = map[string]bool{}

	if *fProjectName == "" {
		tag["project_name"] = cfg.ProjectName
	} else if *fProjectName != cfg.ProjectName {
		update["ProjectName"] = true
		tag["_project_header"] = header(*fProjectName)
	}

	if *fPackageName == "" {
		tag["package_name"] = cfg.PackageName
	} else if *fPackageName != cfg.PackageName {
		update["PackageName"] = true

		if cfg.ProjectType == "lib" || cfg.ProjectType == "cgo" {
			update["PackageInCode"] = true
		}
	}

	if *fLicense == "" {
		*fLicense = cfg.License
	} else if *fLicense != cfg.License {
		update["License"] = true
	}

	if strings.HasPrefix(*fLicense, "cc0") {
		value = "ok"
	} else {
		value = ""
	}
	tag["license_is_cc0"] = value
	tag["license"] = listLicense[*fLicense]

	return tag, update
}

