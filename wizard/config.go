// Copyright 2010  The "Go-Wizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package wizard

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kless/goconfig/config"
	"github.com/kless/Go-Inline/inline"
)

// Flags for the command line
var (
	fProjecType  = flag.String("Project-type", "", "The project type.")
	fProjectName = flag.String("Project-name", "", "The name of the project.")
	fPackageName = flag.String("Package-name", "", "The name of the package.")
	fLicense     = flag.String("License", "", "The license covering the package.")
	fAuthor      = flag.String("Author", "",
		"A string containing the author's name at a minimum.")
	fAuthorEmail = flag.String("Author-email", "",
		"A string containing the author's e-mail address.")

	fAuthorIsOrg = flag.Bool("org", false, "Does the author is an organization?")
	fVCS         = flag.String("vcs", "", "Version control system")
	fAddConfig   = flag.Bool("config", false, "Add the user configuration file")
)

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-type -Project-name -License -Author -Author-email -vcs
	[-Package-name -org -config]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

// Loads configuration from flags and it returns data for templates.
func LoadFlags() {
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

	if len(os.Args) == 1 { // flag.NArg()
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
	// * * *

	// Exit if it is not on interactive way
	if !*fInteractive && (*fListProject || *fListLicense || *fListVCS) {
		os.Exit(0)
	}

	// Get configuration per user
	userConfig()

	if *fInteractive {
		interactive()
	} else {
		setNames()
	}

	// === Checking
	checkAtCreate()
}

//
// === Checking

// Common checking.
func checkCommon(errors bool) {
	// === License
	*fLicense = strings.ToLower(*fLicense)
	if _, present := listLicense[*fLicense]; !present {
		fmt.Fprintf(os.Stderr, "unavailable license: %q\n", *fLicense)
		errors = true
	}

	if *fLicense == "bsd-3" && !*fAuthorIsOrg {
		fmt.Fprintf(os.Stderr,
			"license 'bsd-3' requires an organization as author\n")
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
		fmt.Fprintf(os.Stderr, "missing required fields to create project\n")
		usage()
	}
	if *fAuthorEmail == "" && !*fAuthorIsOrg {
		fmt.Fprintf(os.Stderr, "the email address is required for people\n")
		errors = true
	}

	// === Project type
	*fProjecType = strings.ToLower(*fProjecType)
	if _, present := listProject[*fProjecType]; !present {
		fmt.Fprintf(os.Stderr, "unavailable project type: %q\n", *fProjecType)
		errors = true
	}

	// === VCS
	*fVCS = strings.ToLower(*fVCS)
	if _, present := listVCS[*fVCS]; !present {
		fmt.Fprintf(os.Stderr, "unavailable version control system: %q\n", *fVCS)
		errors = true
	}

	checkCommon(errors)
}

//
// === Utility

// Loads configuration per user, if any.
func userConfig() {
	home, err := os.Getenverror("HOME")
	if err != nil {
		log.Print("no variable HOME:", err)
		return
	}

	pathUserConfig := filepath.Join(home, _USER_CONFIG)

	// To know if the file exist.
	info, err := os.Stat(pathUserConfig)
	if err != nil {
		log.Print("user configuration does not exist:", err)
		return
	}

	if !info.IsRegular() {
		log.Fatal("not a file:", _USER_CONFIG)
		return
	}

	cfg, err := config.ReadDefault(pathUserConfig)
	if err != nil {
		log.Fatal("error parsing configuration:", err)
		return
	}

	// === Get values
	var errors bool
	var errKeys []string

	if *fAuthor == "" {
		*fAuthor, err = cfg.String("DEFAULT", "author")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "author")
		}
	}
	if *fAuthorEmail == "" {
		*fAuthorEmail, err = cfg.String("DEFAULT", "author-email")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "author-email")
		}
	}
	if *fLicense == "" {
		*fLicense, err = cfg.String("DEFAULT", "license")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "license")
		}
	}
	if *fVCS == "" {
		*fVCS, err = cfg.String("DEFAULT", "vcs")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "vcs")
		}
	}

	if errors {
		log.Fatalf("%s: %s\n", err, strings.Join(errKeys, ","))
	}
}

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

	q := inline.NewQuestionByDefault()
	q.ExitAtCtrlC(0) // Exit with Ctrl-C
	defer q.Close()

	fmt.Println("\n  = Go Wizard\n")

	for _, k := range interactiveFlags {
		f := flag.Lookup(k)
		text := strings.TrimRight(f.Usage, ".")

		switch k {
		case "org":
			*fAuthorIsOrg, err = q.ReadBoolDefault(text, false, inline.NONE)
		case "Project-type":
			input, err = q.ReadChoice(text, arrayKeys(listProject),
				inline.NONE)
		case "Project-name":
			input, err = q.ReadString(text, inline.REQUIRED)
		case "Package-name":
			setNames()
			input, err = q.ReadStringDefault(text, *fPackageName, inline.REQUIRED)
		case "Author":
			if *fAuthor != "" {
				input, err = q.ReadStringDefault(text, f.Value.String(), inline.REQUIRED)
				break
			}
			input, err = q.ReadString(text, inline.REQUIRED)
		case "Author-email":
			if *fAuthorIsOrg {
				input, err = q.ReadString(text, inline.NONE)
				break
			}

			if *fAuthorEmail != "" {
				input, err = q.ReadStringDefault(text, f.Value.String(), inline.REQUIRED)
				break
			}
			input, err = q.ReadString(text, inline.REQUIRED)
		case "License":
			if *fLicense != "" {
				input, err = q.ReadChoiceDefault(text, arrayKeys(listLicense), f.Value.String(), inline.NONE)
				break
			}
			input, err = q.ReadChoice(text, arrayKeys(listLicense), inline.NONE)
		case "vcs":
			if *fVCS != "" {
				input, err = q.ReadChoiceDefault(text, arrayKeys(listVCS), f.Value.String(), inline.NONE)
				break
			}
			input, err = q.ReadChoice(text, arrayKeys(listVCS), inline.NONE)
		}

		if err != nil {
			log.Fatal(err)
		}

		if k != "org" {
			flag.Set(k, input)
		}
	}

	fmt.Println()
}

// Sets names for both project and package.
func setNames() {
	// === To remove them from the project name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd := regexp.MustCompile(`-go$`)

	*fProjectName = strings.TrimSpace(*fProjectName)

	// A program is usually named as the project name.
	// It is created removing prefix or suffix related to "go".
	if *fPackageName == "" {
		*fPackageName = strings.ToLower(*fProjectName)

		if reStart1.MatchString(*fPackageName) {
			*fPackageName = reStart1.ReplaceAllString(*fPackageName, "")
		} else if reStart2.MatchString(*fPackageName) {
			*fPackageName = reStart2.ReplaceAllString(*fPackageName, "")
		} else if reEnd.MatchString(*fPackageName) {
			*fPackageName = reEnd.ReplaceAllString(*fPackageName, "")
		}

	} else {
		*fPackageName = strings.ToLower(strings.TrimSpace(*fPackageName))
	}
}
