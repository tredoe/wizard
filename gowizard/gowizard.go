// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kless/terminal/quest"
	"github.com/kless/wizard"
)

type importPaths []string

func (i *importPaths) String() string {
	if len(*i) == 0 {
		return `""`
	}
	return fmt.Sprint(*i)
}

func (i *importPaths) Set(value string) error {
	*i = make([]string, 0)

	for _, v := range strings.Split(value, ":") {
		*i = append(*i, strings.TrimSpace(v))
	}
	return nil
}

var fImportPath importPaths

func init() {
	flag.Var(&fImportPath, "import", "import path; colon-separated list")
}

// * * *

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: gowizard -i [-cfg | -add]

 * Configuration: -cfg -author -email -license -vcs [-org]
 * Project: -type -name -license -author -email -vcs -import [-org]
 * Program: -type -name -license -add

Add new files to current project:

 * File: (-go | -c | -t) name...
 * Installer: -installer
 * Command tester: -tester

`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("ERROR: ")

	cfg, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	if cfg == nil {
		os.Exit(0)
	}

	p, err := wizard.NewProject(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err = p.Create(); err != nil {
		log.Fatal(err)
	}
}

// * * *

// initConfig loads configuration from flags and user configuration.
// Returns the configuration to nil when it is used the flag "cfg".
func initConfig() (*wizard.Conf, error) {
	var (
		fType    = flag.String("type", "", "type of project")
		fName    = flag.String("name", "", "name of the project or program")
		fLicense = flag.String("license", "", "license covering the program")
		fAuthor  = flag.String("author", "", "author's name")
		fEmail   = flag.String("email", "", "author's e-mail")
		fVCS     = flag.String("vcs", "", "version control system")
		fOrg     = flag.String("org", "", "organization holder of the copyright")

		fConfig      = flag.Bool("cfg", false, "add the user configuration file")
		fAdd         = flag.Bool("add", false, "add a program")
		fInteractive = flag.Bool("i", false, "interactive mode")

		fGo   = flag.Bool("go", false, "new Go source file")
		fCgo  = flag.Bool("c", false, "new Cgo source file")
		fTest = flag.Bool("t", false, "new test file")

		fInstaller = flag.Bool("installer", false, "add a manager related to the system")
		fTester    = flag.Bool("tester", false, "add test file for testing the command output")

		// Listing
		fListType    = flag.Bool("lt", false, "list the available project types (for type flag)")
		fListLicense = flag.Bool("ll", false, "list the available licenses (for license flag)")
		fListVCS     = flag.Bool("lv", false, "list the available version control systems (for vcs flag)")
	)

	// == Parse the flags
	flag.Usage = usage
	flag.Parse()

	if flag.NFlag() == 0 || (*fAdd && *fConfig) {
		usage()
	}

	// == Listing
	if *fListType {
		maxLen := 0
		for _, v := range wizard.ListTypeSorted {
			if len(v) > maxLen {
				maxLen = len(v)
			}
		}

		fmt.Print("  = Project types\n\n")
		for _, v := range wizard.ListTypeSorted {
			fmt.Printf("  %s: %s%s\n",
				v, strings.Repeat(" ", maxLen-len(v)), wizard.ListType[v],
			)
		}
	}
	if *fListLicense {
		maxLen := 0
		for _, v := range wizard.ListLicenseSorted {
			if len(v) > maxLen {
				maxLen = len(v)
			}
		}

		fmt.Print("  = Licenses\n\n")
		for _, v := range wizard.ListLicenseSorted {
			fmt.Printf("  %s: %s%s\n",
				v, strings.Repeat(" ", maxLen-len(v)), wizard.ListLicense[v],
			)
		}
	}
	if *fListVCS {
		maxLen := 0
		for _, v := range wizard.ListVCSsorted {
			if len(v) > maxLen {
				maxLen = len(v)
			}
		}

		fmt.Print("  = Version control systems\n\n")
		for _, v := range wizard.ListVCSsorted {
			fmt.Printf("  %s: %s%s\n",
				v, strings.Repeat(" ", maxLen-len(v)), wizard.ListVCS[v],
			)
		}
	}

	if *fListType || *fListLicense || *fListVCS {
		return nil, nil
	}

	// == New file
	var err error

	if *fGo || *fCgo || *fTest || *fInstaller || *fTester {
		err = wizard.NewFile(*fGo, *fCgo, *fTest, *fInstaller, *fTester, flag.Args()...)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// * * *

	cfg := &wizard.Conf{
		Type:        *fType,
		Program:     *fName,
		License:     *fLicense,
		Author:      *fAuthor,
		Email:       *fEmail,
		VCS:         *fVCS,
		ImportPaths: fImportPath,
		Org:         *fOrg,

		IsNewProject: !*fAdd,
	}

	// Get configuration per user, if any.
	if !*fConfig {
		if err = cfg.UserConfig(); err != nil {
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
	if err = cfg.CheckAndSetNames(*fInteractive, *fConfig, *fAdd); err != nil {
		return nil, err
	}

	// Interactive mode
	if *fInteractive {
		if err = interactive(cfg, *fConfig, *fAdd); err != nil {
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
		sFlags = []string{"author", "email", "license", "vcs", "import", "org"}
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
			"import",
		}
	}

	q := quest.NewDefault()
	defer q.Restore()
	q.ExitAtCtrlC(0)

	fmt.Printf("\n  = Gowizard :: %s\n\n", msg)

	for _, k := range sFlags {
		f := flag.Lookup(k)

		if strings.Contains(f.Usage, ";") {
			f.Usage = strings.SplitN(f.Usage, ";", 2)[0]
		}
		f.Usage = strings.ToUpper(string(f.Usage[0])) + f.Usage[1:]
		prompt := q.NewPrompt(f.Usage)

		switch k {
		case "type":
			if c.Type == "" {
				if c.IsNewProject {
					c.Type = "pkg"
				} else {
					c.Type = "cmd"
				}
			}

			c.Type, err = prompt.Default(c.Type).ChoiceString(wizard.ListTypeSorted)
		case "name":
			if addProgram {
				c.Program, err = prompt.Default(c.Program).Mod(quest.REQUIRED).ReadString()
			} else {
				c.Project, err = prompt.Default(c.Project).Mod(quest.REQUIRED).ReadString()
			}
			if err = c.SetNames(addProgram); err != nil {
				return err
			}
		case "org":
			isOrg := true

			if c.Org == "" {
				prompt := q.NewPrompt("Is for an organization?")
				isOrg, err = prompt.Default(false).ReadBool()
			}
			if isOrg {
				prompt := q.NewPrompt(f.Usage)
				c.Org, err = prompt.Default(c.Org).ReadString()
			}
		case "author":
			c.Author, err = prompt.Default(c.Author).Mod(quest.REQUIRED).ReadString()
		case "email":
			c.Email, err = prompt.Default(c.Email).Mod(quest.REQUIRED).ReadEmail()
		case "license":
			// It is got in upper case
			c.License, err = prompt.Default(wizard.ListLowerLicense[c.License]).
				ChoiceString(wizard.ListLicenseSorted)
			c.License = strings.ToLower(c.License)
		case "vcs":
			c.VCS, err = prompt.Default(c.VCS).ChoiceString(wizard.ListVCSsorted)
		case "import":
			if addConfig {
				c.ImportPaths, err = prompt.ReadMultipleString()
			} else {
				c.ImportPaths[0], err = prompt.Default(c.ImportPaths[0]).ChoiceString(c.ImportPaths)
			}
		}

		if err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}
