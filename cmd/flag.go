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
	"container/vector"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/kless/go-readin/readin"
	"github.com/kless/goconfig/config"
)


const USER_CONFIG = ".gowizard" // Configuration file per user

// Flags for the command line
var (
	// === Metadata
	fProjecType  = flag.String("Project-type", "", "The project type.")
	fProjectName = flag.String("Project-name", "", "The name of the project.")
	fPackageName = flag.String("Package-name", "", "The name of the package.")
	fLicense     = flag.String("License", "", "The license covering the package.")
	fAuthor      = flag.String("Author", "",
		"A string containing the author's name at a minimum.")
	fAuthorEmail = flag.String("Author-email", "",
		"A string containing the author's e-mail address.")

	fAuthorIsOrg = flag.Bool("org", false, "Does the author is an organization?")
	fDebug       = flag.Bool("d", false, "Debug mode")
	fUpdate      = flag.Bool("u", false, "Update data on project created")
	fVerbose     = flag.Bool("v", false, "Show files updated")
	fVCS         = flag.String("vcs", "", "Version control system")
)


func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-type -Project-name -License -Author -Author-email -vcs
	[-Package-name -org]

       gowizard -u [-Project-name -Package-name -License] [-org]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

// Loads configuration from flags and it returns tags for templates.
func loadConfig() {
	// === Generic flags
	fInteractive := flag.Bool("i", false, "Interactive mode")

	fListLicense := flag.Bool("ll", false,
		"Show the list of licenses for the flag 'License'")
	fListProject := flag.Bool("lp", false,
		"Show the list of project types for the flag 'Project-type'")
	fListVCS := flag.Bool("lv", false,
		"Show the list of version control systems")

	// === Parse the flags
	flag.Usage = usage
	flag.Parse()

	if len(os.Args) == 1 {
		usage()
	}

	// Get configuration per user
	if !*fUpdate {
		userConfig()
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

	// Exit if it is not on interactive way
	if !*fInteractive && (*fListProject || *fListLicense || *fListVCS) {
		os.Exit(0)
	}

	if !*fUpdate {
		if *fInteractive {
			interactive()
		} else {
			setNames()
		}
	}

	// === Checking
	if !*fUpdate {
		checkAtCreate()
	} else {
		checkAtUpdate()
	}
}


// === Checking
// ===

// Common checking.
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
		if *fUpdate {
			fmt.Fprintf(os.Stderr,
				"\tThe author name for the license is got from metadata file\n")
		}
		errors = true
	}

	if errors {
		os.Exit(ERROR)
	}
}

// Checks at creating project.
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

// Checks at updating project.
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


// Interactive mode.
func interactive() {
	var input string
	var err os.Error

	// Sorted flags
	interactiveFlags := []string{
		"org",
		"Project-type",
		"Project-name",
		"Package-name",
		"Author",
		"Author-email",
		"License",
		"vcs",
	}

	fmt.Println("\n  = Interactive\n")
	readin.DefaultIndent = "  "

	for _, k := range interactiveFlags {
		f := flag.Lookup(k)
		text := "+ " + strings.TrimRight(f.Usage, ".")

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

// Sets names for both project and package.
func setNames() {
	reGo := regexp.MustCompile(`^go`) // To remove it from the project name

	// The project name is not changed to lower case to add it so to the tag.
	*fProjectName = strings.TrimSpace(*fProjectName)

	switch *fProjecType {
	// A program is usually named as the project name.
	case "app", "tool":
		if *fPackageName == "" {
			*fPackageName = strings.ToLower(*fProjectName)
		}
	// For libraries (packages)
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

// Creates tags to pass them to templates. Used at creating a new project.
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

	// Now, the project name can be changed to lower case.
	*fProjectName = strings.ToLower(*fProjectName)

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

	if *fVCS == "none" {
		value = "ok"
	} else {
		value = ""
	}
	tag["vcs_is_none"] = value

	return tag
}

// Creates tags to pass them to templates. Used at updating a project.
//
// If the flags are empty then they are set with metadata values.
// If flags are not empty then is indicated in map 'update' since they are the
// values to change.
func tagsToUpdate(cfg *Metadata) (tag map[string]string, update map[string]bool) {
	update = map[string]bool{}

	// setNames() Needs to know the 'ProjectType'
	*fProjecType = cfg.ProjectType
	setNames()

	tag = map[string]string{
		"project_name": *fProjectName,
		"package_name": *fPackageName,
	}

	if *fProjectName == "" {
		tag["project_name"] = cfg.ProjectName
	} else if *fProjectName != cfg.ProjectName {
		update["ProjectName"] = true

		*fProjectName = strings.ToLower(*fProjectName)
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

		if *fLicense == "bsd-3" {
			tag["author"] = cfg.Author
		}
	}
	tag["license"] = listLicense[*fLicense]

	return tag, update
}

// Loads configuration per user, if any.
func userConfig() {
	home, err := os.Getenverror("HOME")
	if err != nil {
		if *fDebug {
			fmt.Fprintf(os.Stderr, "\n%s: userConfig(): %s: HOME\n\n", argv0, err)
		}
		return
	}

	pathUserConfig := path.Join(home, USER_CONFIG)

	// To know if the file exist.
	info, err := os.Stat(pathUserConfig)
	if err != nil {
		if *fDebug {
			fmt.Fprintf(os.Stderr, "\n%s: userConfig(): %s\n\n", argv0, err)
		}
		return
	}

	if !info.IsRegular() {
		if *fDebug {
			fmt.Fprintf(os.Stderr, "\n%s: userConfig(): %s is not a file\n\n",
				argv0, pathUserConfig)
		}
		return
	}

	cfg, err := config.ReadDefault(pathUserConfig)
	if err != nil {
		if *fDebug {
			fmt.Fprintf(os.Stderr, "\n%s: userConfig(): %s\n\n", argv0, err)
		}
		return
	}

	// === Get values
	var errors bool
	errKeys := new(vector.StringVector)

	if *fAuthor == "" {
		*fAuthor, err = cfg.String("DEFAULT", "author")
		if err != nil {
			errors = true
			errKeys.Push("author")
		}
	}
	if *fAuthorEmail == "" {
		*fAuthorEmail, err = cfg.String("DEFAULT", "author-email")
		if err != nil {
			errors = true
			errKeys.Push("author-email")
		}
	}
	if *fLicense == "" {
		*fLicense, err = cfg.String("DEFAULT", "license")
		if err != nil {
			errors = true
			errKeys.Push("license")
		}
	}
	if *fVCS == "" {
		*fVCS, err = cfg.String("DEFAULT", "vcs")
		if err != nil {
			errors = true
			errKeys.Push("vcs")
		}
	}

	if errors {
		s := ""

		for i, val := range *errKeys {
			if i == 0 {
				s = val
			} else {
				s += ", " + val
			}
		}

		fmt.Fprintf(os.Stderr, "%s: %s: %s\n", argv0, err, s)
		os.Exit(ERROR)
	}
}
