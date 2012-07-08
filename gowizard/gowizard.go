// Copyright 2012  The "Gowizard" Authors
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kless/Go-Inline/quest"
	"github.com/kless/wizard"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Tool to create skeleton of Go projects.
Usage: gowizard -i [-cfg | -add]

 * Configuration: -cfg -author -email -license -vcs [-org]
 * Project: -type -name -license -author -email -vcs [-org]
 * Program: -type -name -license -add

`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		fatalf("%s", err)
	}
	if cfg == nil { // flag "-cfg"
		os.Exit(0)
	}

	p, err := wizard.NewProject(cfg)
	if err != nil {
		fatalf("%s", err)
	}

	p.Create()
}

// * * *

// initConfig loads configuration from flags and user configuration.
// Returns the configuration to nil when it is used the flag "cfg".
func initConfig() (*wizard.Conf, error) {
	var (
		fType    = flag.String("type", "", "The type of project.")
		fName    = flag.String("name", "", "The name of the project or program.")
		fLicense = flag.String("license", "", "The license covering the program.")
		fAuthor  = flag.String("author", "", "The author's name.")
		fEmail   = flag.String("email", "", "The author's e-mail.")
		fVCS     = flag.String("vcs", "", "Version control system.")
		fOrg     = flag.String("org", "", "The organization holder of the copyright.")

		fAdd         = flag.Bool("add", false, "Add a program.")
		fConfig      = flag.Bool("cfg", false, "Add the user configuration file.")
		fInteractive = flag.Bool("i", false, "Interactive mode.")

		// Listing
		fListType    = flag.Bool("lt", false, "Show the list of project types (for flag \"type\").")
		fListLicense = flag.Bool("ll", false, "Show the list of licenses (for flag \"license\").")
		fListVCS     = flag.Bool("lv", false, "Show the list of version control systems (for flag \"vcs\").")
	)

	// == Parse the flags
	flag.Usage = usage
	flag.Parse()

	if flag.NFlag() == 0 || (*fAdd && *fConfig) {
		usage()
	}

	// == Listing
	if *fListType {
		fmt.Print("  = Project types\n\n")
		for k, v := range wizard.ListType {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	if *fListLicense {
		fmt.Print("  = Licenses\n\n")
		for k, v := range wizard.ListLicense {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	if *fListVCS {
		fmt.Print("  = Version control systems\n\n")
		for k, v := range wizard.ListVCS {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if *fListType || *fListLicense || *fListVCS {
		os.Exit(0)
	}
	// * * *

	cfg := &wizard.Conf{
		Type:    *fType,
		Program: *fName,
		License: *fLicense,
		Author:  *fAuthor,
		Email:   *fEmail,
		VCS:     *fVCS,
		Org:     *fOrg,

		IsNewProject: !*fAdd,
	}

	// Get configuration per user, if any.
	if !*fConfig {
		if err := cfg.UserConfig(); err != nil {
			return nil, err
		}
	}

	// New program for existent project.
	if *fAdd {
		if *fLicense != "" {
			cfg.License = *fLicense
		}

		// The project's name is the name of the actual directory.
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cfg.Project = filepath.Base(wd)
	} else {
		cfg.Project = *fName
	}

	// Check flags
	if err := cfg.CheckAndSetNames(*fInteractive, *fConfig, *fAdd); err != nil {
		return nil, err
	}

	// Interactive mode
	if *fInteractive {
		if err := interactive(cfg, *fConfig, *fAdd); err != nil {
			return nil, err
		}
	}

	// Add configuration.
	if *fConfig {
		cfg.AddConfig()
		return nil, nil
	}

	cfg.AddTemplateData()
	return cfg, nil
}

// interactive uses the interactive mode.
func interactive(c *wizard.Conf, addConfig, addProgram bool) error {
	var sFlags []string
	var msg string
	var err error

	// == Sorted flags
	if addConfig {
		msg = "New configuration"
		sFlags = []string{"author", "email", "license", "vcs", "org"}
	} else if addProgram {
		msg = "Add program"
		sFlags = []string{"type", "name", "license"}
	} else {
		msg = "New project"
		sFlags = []string{
			"type",
			"name",
			"org",
			"author",
			"email",
			"license",
			"vcs",
		}
	}

	q := quest.NewQuestionByDefault()
	defer q.Restore()
	q.ExitAtCtrlC(0)

	fmt.Printf("\n  = Gowizard :: %s\n\n", msg)

	for _, k := range sFlags {
		f := flag.Lookup(k)
		prompt := q.NewPrompt(strings.TrimRight(f.Usage, "."))

		switch k {
		case "type":
			if c.Type == "" {
				if c.IsNewProject {
					c.Type = "pkg"
				} else {
					c.Type = "cmd"
				}
			}

			c.Type, err = prompt.ByDefault(c.Type).ChoiceString(keys(wizard.ListType))
		case "name":
			if addProgram {
				c.Program, err = prompt.ByDefault(c.Program).Mod(quest.REQUIRED).ReadString()
			} else {
				c.Project, err = prompt.ByDefault(c.Project).Mod(quest.REQUIRED).ReadString()
			}
			c.SetNames(addProgram)
		case "org":
			c.Org, err = prompt.ByDefault(c.Org).ReadString()
		case "author":
			c.Author, err = prompt.ByDefault(c.Author).Mod(quest.REQUIRED).ReadString()
		case "email":
			c.Email, err = prompt.ByDefault(c.Email).Mod(quest.REQUIRED).ReadEmail()
		case "license":
			// It is got in upper case
			c.License, err = prompt.ByDefault(wizard.ListLowerLicense[c.License]).
				ChoiceString(keys(wizard.ListLicense))
			c.License = strings.ToLower(c.License)
		case "vcs":
			c.VCS, err = prompt.ByDefault(c.VCS).ChoiceString(keys(wizard.ListVCS))
		}

		if err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}

// == Utility
//

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// keys gets an array from map keys.
func keys(m map[string]string) []string {
	a := make([]string, len(m))
	i := 0

	for k, _ := range m {
		a[i] = k
		i++
	}
	return a
}
