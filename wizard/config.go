// Copyright 2010  The "GoWizard" Authors
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
	"github.com/kless/Go-Inline/quest"
)

// Represents the configuration of the project.
type conf struct {
	projecType  string
	ProjectName string
	PackageName string
	license     string
	Author      string
	Email       string
	AuthorIsOrg bool
	vcs         string

	addUserConf bool

	// To pass to templates
	Comment       string
	FullLicense   string
	GNUextra      string
	ProjectHeader string
	IsCmdProject  bool
	IsCgoProject  bool
	Year          int
}

// * * *

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -project-type -project-name -license -author -email -vcs
	[-package-name -org -config]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

// Loads configuration from flags and user configuration.
func initConfig() *conf {
	fProjecType := flag.String("project-type", "", "The project type.")
	fProjectName := flag.String("project-name", "", "The name of the project.")
	fPackageName := flag.String("package-name", "", "The name of the package.")
	fLicense := flag.String("license", "", "The license covering the package.")
	fAuthor := flag.String("author", "", "The author's name.")
	fEmail := flag.String("email", "", "The author's e-mail address.")
	fAuthorIsOrg := flag.Bool("org", false, "Does the author is an organization?")
	fVCS := flag.String("vcs", "", "Version control system")

	// === Generic flags
	fAddUserConf := flag.Bool("config", false, "Add the user configuration file")
	fInteractive := flag.Bool("i", false, "Interactive mode")

	fListLicense := flag.Bool("ll", false,
		"Show the list of licenses for the flag 'license'")
	fListProject := flag.Bool("lp", false,
		"Show the list of project types for the flag 'project-type'")
	fListVCS := flag.Bool("lv", false,
		"Show the list of version control systems")

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
	// * * *

	// Exit if it is not on interactive way
	if !*fInteractive && (*fListProject || *fListLicense || *fListVCS) {
		os.Exit(0)
	}

	cfg := &conf{
		projecType:  *fProjecType,
		ProjectName: *fProjectName,
		PackageName: *fPackageName,
		license:     *fLicense,
		Author:      *fAuthor,
		Email:       *fEmail,
		AuthorIsOrg: *fAuthorIsOrg,
		vcs:         *fVCS,
		addUserConf: *fAddUserConf,
	}

	// Get configuration per user
	userConfig(cfg)

	if *fInteractive {
		interactive(cfg)
	} else {
		setNames(cfg)
	}

	// === Checking
	checkAtCreate(cfg)

	// === Extra for templates
	cfg.ProjectHeader = createHeader(cfg.ProjectName)

	if cfg.license != "none" {
		cfg.FullLicense = listLicense[cfg.license]
	}
	if cfg.projecType == "cgo" {
		cfg.IsCgoProject = true
	}
	// For the Makefile
	if cfg.projecType == "cmd" {
		cfg.IsCmdProject = true
	}

	return cfg
}

//
// === Checking

// Common checking.
func checkCommon(c *conf, errors bool) {
	// === License
	c.license = strings.ToLower(c.license)
	if _, present := listLicense[c.license]; !present {
		fmt.Fprintf(os.Stderr, "unavailable license: %q\n", c.license)
		errors = true
	}

	if c.license == "bsd-3" && !c.AuthorIsOrg {
		fmt.Fprintf(os.Stderr,
			"license 'bsd-3' requires an organization as author\n")
		errors = true
	}

	if errors {
		os.Exit(ERROR)
	}
}

// Checks at creating project.
func checkAtCreate(c *conf) {
	var errors bool

	// === Necessary fields
	if c.projecType == "" || c.ProjectName == "" || c.license == "" ||
		c.Author == "" || c.vcs == "" {
		fmt.Fprintf(os.Stderr, "missing required fields to create project\n")
		usage()
	}
	if c.Email == "" && !c.AuthorIsOrg {
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

	checkCommon(c, errors)
}

//
// === Utility

// Loads configuration per user, if any.
func userConfig(c *conf) {
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

	if c.Author == "" {
		c.Author, err = cfg.String("DEFAULT", "author")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "author")
		}
	}
	if c.Email == "" {
		c.Email, err = cfg.String("DEFAULT", "email")
		if err != nil {
			errors = true
			errKeys = append(errKeys, "email")
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
func interactive(c *conf) {
	var err os.Error

	// Sorted flags
	interactiveFlags := []string{
		"org",
		"project-type",
		"project-name",
		"package-name",
		"author",
		"email",
		"license",
		"vcs",
	}

	q := quest.NewQuestionByDefault()
	defer q.Close()
	q.ExitAtCtrlC(0)

	fmt.Println("\n  = Go Wizard\n")

	for _, k := range interactiveFlags {
		f := flag.Lookup(k)
		prompt := q.NewPrompt(strings.TrimRight(f.Usage, "."))

		switch k {
		case "org":
			c.AuthorIsOrg, err = prompt.ByDefault(false).ReadBool()
		case "project-type":
			c.projecType, err = prompt.ChoiceString(arrayKeys(listProject))
		case "project-name":
			c.ProjectName, err = prompt.Mod(quest.REQUIRED).ReadString()
		case "package-name":
			setNames(c)
			c.PackageName, err = prompt.ByDefault(c.PackageName).ReadString()
		case "author":
			if c.Author != "" {
				c.Author, err = prompt.ByDefault(c.Author).ReadString()
				break
			}
			c.Author, err = prompt.Mod(quest.REQUIRED).ReadString()
		case "email":
			if c.AuthorIsOrg {
				c.Email, err = prompt.ReadEmail()
				break
			}

			if c.Email != "" {
				c.Email, err = prompt.ByDefault(c.Email).ReadEmail()
				break
			}
			c.Email, err = prompt.Mod(quest.REQUIRED).ReadEmail()
		case "license":
			if c.license != "" {
				c.license, err = prompt.ByDefault(c.license).ChoiceString(arrayKeys(listLicense))
				break
			}
			c.license, err = prompt.ChoiceString(arrayKeys(listLicense))
		case "vcs":
			if c.vcs != "" {
				c.vcs, err = prompt.ByDefault(c.vcs).ChoiceString(arrayKeys(listVCS))
				break
			}
			c.vcs, err = prompt.ChoiceString(arrayKeys(listVCS))
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println()
}

// Sets names for both project and package.
func setNames(c *conf) {
	// === To remove them from the project name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd := regexp.MustCompile(`-go$`)

	c.ProjectName = strings.TrimSpace(c.ProjectName)

	// A program is usually named as the project name.
	// It is created removing prefix or suffix related to "go".
	if c.PackageName == "" {
		c.PackageName = strings.ToLower(c.ProjectName)

		if reStart1.MatchString(c.PackageName) {
			c.PackageName = reStart1.ReplaceAllString(c.PackageName, "")
		} else if reStart2.MatchString(c.PackageName) {
			c.PackageName = reStart2.ReplaceAllString(c.PackageName, "")
		} else if reEnd.MatchString(c.PackageName) {
			c.PackageName = reEnd.ReplaceAllString(c.PackageName, "")
		}

	} else {
		c.PackageName = strings.ToLower(strings.TrimSpace(c.PackageName))
	}
}
