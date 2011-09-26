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

// Represents the configuration of the project.
type conf struct {
	projecType  string
	projectName string
	packageName string
	license     string
	author      string
	authorEmail string
	authorIsOrg bool
	vcs         string

	addUserConf bool
}

// * * *

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -Project-type -Project-name -License -Author -Author-email -vcs
	[-Package-name -org -config]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

// Loads configuration from flags and user configuration.
func initConfig() *conf {
	fProjecType := flag.String("Project-type", "", "The project type.")
	fProjectName := flag.String("Project-name", "", "The name of the project.")
	fPackageName := flag.String("Package-name", "", "The name of the package.")
	fLicense := flag.String("License", "", "The license covering the package.")
	fAuthor := flag.String("Author", "",
		"A string containing the author's name at a minimum.")
	fAuthorEmail := flag.String("Author-email", "",
		"A string containing the author's e-mail address.")

	fAuthorIsOrg := flag.Bool("org", false, "Does the author is an organization?")
	fVCS := flag.String("vcs", "", "Version control system")

	// === Generic flags
	fAddUserConf := flag.Bool("config", false, "Add the user configuration file")
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

	cfg := &conf{
		projecType:  *fProjecType,
		projectName: *fProjectName,
		packageName: *fPackageName,
		license:     *fLicense,
		author:      *fAuthor,
		authorEmail: *fAuthorEmail,
		authorIsOrg: *fAuthorIsOrg,
		vcs:         *fVCS,
		addUserConf: *fAddUserConf,
	}

	// Get configuration per user
	cfg.userConfig()

	if *fInteractive {
		cfg.interactive()
	} else {
		cfg.setNames()
	}

	// === Checking
	cfg.checkAtCreate()

	return cfg
}

//
// === Checking

// Common checking.
func (c *conf) checkCommon(errors bool) {
	// === License
	c.license = strings.ToLower(c.license)
	if _, present := listLicense[c.license]; !present {
		fmt.Fprintf(os.Stderr, "unavailable license: %q\n", c.license)
		errors = true
	}

	if c.license == "bsd-3" && !c.authorIsOrg {
		fmt.Fprintf(os.Stderr,
			"license 'bsd-3' requires an organization as author\n")
		errors = true
	}

	if errors {
		os.Exit(ERROR)
	}
}

// Checks at creating project.
func (c *conf) checkAtCreate() {
	var errors bool

	// === Necessary fields
	if c.projecType == "" || c.projectName == "" || c.license == "" ||
		c.author == "" || c.vcs == "" {
		fmt.Fprintf(os.Stderr, "missing required fields to create project\n")
		usage()
	}
	if c.authorEmail == "" && !c.authorIsOrg {
		fmt.Fprintf(os.Stderr, "the email address is required for people\n")
		errors = true
	}

	// === Project type
	c.projecType = strings.ToLower(c.projecType)
	if _, present := listProject[c.projecType]; !present {
		fmt.Fprintf(os.Stderr, "unavailable project type: %q\n", c.projecType)
		errors = true
	}

	// === VCS
	c.vcs = strings.ToLower(c.vcs)
	if _, present := listVCS[c.vcs]; !present {
		fmt.Fprintf(os.Stderr, "unavailable version control system: %q\n", c.vcs)
		errors = true
	}

	c.checkCommon(errors)
}

//
// === Utility

// Loads configuration per user, if any.
func (c *conf) userConfig() {
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

	if c.author == "" {
		c.author, err = cfg.String("DEFAULT", "author")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "author")
		}
	}
	if c.authorEmail == "" {
		c.authorEmail, err = cfg.String("DEFAULT", "author-email")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "author-email")
		}
	}
	if c.license == "" {
		c.license, err = cfg.String("DEFAULT", "license")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "license")
		}
	}
	if c.vcs == "" {
		c.vcs, err = cfg.String("DEFAULT", "vcs")
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
func (c *conf) interactive() {
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
			c.authorIsOrg, err = q.ReadBoolDefault(text, false, inline.NONE)
		case "Project-type":
			input, err = q.ReadChoice(text, arrayKeys(listProject),
				inline.NONE)
		case "Project-name":
			input, err = q.ReadString(text, inline.REQUIRED)
		case "Package-name":
			c.setNames()
			input, err = q.ReadStringDefault(text, c.packageName, inline.REQUIRED)
		case "Author":
			if c.author != "" {
				input, err = q.ReadStringDefault(text, f.Value.String(), inline.REQUIRED)
				break
			}
			input, err = q.ReadString(text, inline.REQUIRED)
		case "Author-email":
			if c.authorIsOrg {
				input, err = q.ReadString(text, inline.NONE)
				break
			}

			if c.authorEmail != "" {
				input, err = q.ReadStringDefault(text, f.Value.String(), inline.REQUIRED)
				break
			}
			input, err = q.ReadString(text, inline.REQUIRED)
		case "License":
			if c.license != "" {
				input, err = q.ReadChoiceDefault(text, arrayKeys(listLicense), f.Value.String(), inline.NONE)
				break
			}
			input, err = q.ReadChoice(text, arrayKeys(listLicense), inline.NONE)
		case "vcs":
			if c.vcs != "" {
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
func (c *conf) setNames() {
	// === To remove them from the project name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd := regexp.MustCompile(`-go$`)

	c.projectName = strings.TrimSpace(c.projectName)

	// A program is usually named as the project name.
	// It is created removing prefix or suffix related to "go".
	if c.packageName == "" {
		c.packageName = strings.ToLower(c.projectName)

		if reStart1.MatchString(c.packageName) {
			c.packageName = reStart1.ReplaceAllString(c.packageName, "")
		} else if reStart2.MatchString(c.packageName) {
			c.packageName = reStart2.ReplaceAllString(c.packageName, "")
		} else if reEnd.MatchString(c.packageName) {
			c.packageName = reEnd.ReplaceAllString(c.packageName, "")
		}

	} else {
		c.packageName = strings.ToLower(strings.TrimSpace(c.packageName))
	}
}
