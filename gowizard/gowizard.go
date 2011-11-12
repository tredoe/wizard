// Copyright 2010  The "GoWizard" Authors
//
// Use of this source code is governed by the BSD 2-Clause License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kless/Go-Inline/quest"
	"github.com/kless/GoWizard/wizard"
)

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: gowizard -i

	+ -add-config -author -email -license -vcs [-org-name]
	+ -add-license
	+ -project-type -project-name -license -author -email -vcs
	  [-package-name -org -org-name]

`)
	flag.PrintDefaults()
	os.Exit(ERROR)
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		fatalf(err.Error())
	}

	p, err := wizard.NewProject(cfg)
	if err != nil {
		fatalf(err.Error())
	}

	p.Create()
}

// * * *

// Loads configuration from flags and user configuration.
func initConfig() (*wizard.Conf, error) {
	fProjecType := flag.String("project-type", "", "The project type.")
	fProjectName := flag.String("project-name", "", "The name of the project.")
	fPackageName := flag.String("package-name", "", "The name of the package.")
	fLicense := flag.String("license", "", "The license covering the package.")
	fAuthor := flag.String("author", "", "The author's name.")
	fEmail := flag.String("email", "", "The author's e-mail.")
	fVCS := flag.String("vcs", "", "Version control system.")
	fOrgName := flag.String("org-name", "", "The organization's name.")
	fIsForOrg := flag.Bool("org", false, "Does an organization is the copyright holder?")

	// === Generic flags
	fAddLicense := flag.String("add-license", "", "Add a license file.")
	fAddConfig := flag.Bool("add-config", false, "Add the user configuration file.")
	fInteractive := flag.Bool("i", false, "Interactive mode.")

	fListLicense := flag.Bool("ll", false,
		"Show the list of licenses (for flags \"license\" and \"add-license\").")
	fListProject := flag.Bool("lp", false,
		"Show the list of project types (for flag \"project-type\").")
	fListVCS := flag.Bool("lv", false,
		"Show the list of version control systems (for flag \"vcs\").")

	// === Parse the flags
	flag.Usage = usage
	flag.Parse()

	if len(os.Args) == 1 {
		usage()
	}

	// === Options
	if *fListProject {
		fmt.Println("  = Project types\n")
		for k, v := range wizard.ListProject {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	if *fListLicense {
		fmt.Println("  = Licenses\n")
		for _, v := range wizard.ListLicense {
			fmt.Printf("  %s: %s\n", v[0], v[1])
		}
	}
	if *fListVCS {
		fmt.Println("  = Version control systems\n")
		for k, v := range wizard.ListVCS {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	// * * *

	// Exit if it is not on interactive way
	if !*fInteractive && (*fListProject || *fListLicense || *fListVCS) {
		os.Exit(0)
	}

	cfg := &wizard.Conf{
		ProjecType:  *fProjecType,
		ProjectName: *fProjectName,
		PackageName: *fPackageName,
		License:     *fLicense,
		Author:      *fAuthor,
		Email:       *fEmail,
		VCS:         *fVCS,
		OrgName:     *fOrgName,
		IsForOrg:    *fIsForOrg,
	}

	// Add configuration.
	if *fAddConfig {
		wizard.AddConfig(cfg)
		os.Exit(0)
	}

	// New license for existent project.
	if *fAddLicense != "" {
		cfg.License = *fAddLicense

		// The project name is the name of the actual directory.
		wd, err := os.Getwd()
		if err != nil {
			fatalf(err.Error())
		}
		cfg.ProjectName = filepath.Base(wd)

		// Get year of project's creation
		year, err := wizard.ProjectYear("README.md")
		if err != nil {
			fatalf(err.Error())
		}

		project, err := wizard.NewProject(cfg)
		if err != nil {
			fatalf(err.Error())
		}
		project.ParseLicense(wizard.CHAR_COMMENT, year)

		if err = wizard.AddLicense(project, false); err != nil {
			os.Exit(ERROR)
		}
		os.Exit(0)
	}

	// Get configuration per user
	err := wizard.UserConfig(cfg)
	if err != nil {
		return nil, err
	}

	if *fInteractive {
		if err = interactive(cfg); err != nil {
			return nil, err
		}
	} else {
		wizard.SetNames(cfg)
	}

	if err = wizard.ExtraConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Interactive mode.
func interactive(c *wizard.Conf) error {
	var err error

	// Sorted flags
	interactiveFlags := []string{
		"project-type",
		"project-name",
		"package-name",
		"org",
		"org-name",
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
		case "project-type":
			c.ProjecType, err = prompt.ChoiceString(keys(wizard.ListProject))
		case "project-name":
			c.ProjectName, err = prompt.Mod(quest.REQUIRED).ReadString()
		case "package-name":
			wizard.SetNames(c)
			c.PackageName, err = prompt.ByDefault(c.PackageName).ReadString()
		case "org":
			c.IsForOrg, err = prompt.ByDefault(c.IsForOrg).ReadBool()
		case "org-name":
			if c.IsForOrg {
				if c.OrgName != "" {
					c.OrgName, err = prompt.ByDefault(c.OrgName).ReadString()
				} else {
					c.OrgName, err = prompt.Mod(quest.REQUIRED).ReadString()
				}
			}
		case "author":
			if c.Author != "" {
				c.Author, err = prompt.ByDefault(c.Author).ReadString()
			} else {
				c.Author, err = prompt.Mod(quest.REQUIRED).ReadString()
			}
		case "email":
			if c.Email != "" {
				c.Email, err = prompt.ByDefault(c.Email).ReadEmail()
			} else {
				c.Email, err = prompt.Mod(quest.REQUIRED).ReadEmail()
			}
		case "license":
			if c.License != "" {
				c.License, err = prompt.ByDefault(c.License).
					ChoiceString(extraKeys(wizard.ListLicense))
			} else {
				c.License, err = prompt.ChoiceString(extraKeys(wizard.ListLicense))
			}
		case "vcs":
			if c.VCS != "" {
				c.VCS, err = prompt.ByDefault(c.VCS).ChoiceString(keys(wizard.ListVCS))
			} else {
				c.VCS, err = prompt.ChoiceString(keys(wizard.ListVCS))
			}
		}

		if err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}

//
// === Utility

// Gets an array from map keys.
func keys(m map[string]string) []string {
	a := make([]string, len(m))
	i := 0

	for k, _ := range m {
		a[i] = k
		i++
	}
	return a
}

// Gets an array from the first slice in the map's value.
func extraKeys(m map[string][]string) []string {
	a := make([]string, len(m))
	i := 0

	for _, v := range m {
		a[i] = v[0]
		i++
	}
	return a
}

//
// === Error

const ERROR = 1

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "gowizard: "+format+"\n", a...)
	os.Exit(ERROR)
}
