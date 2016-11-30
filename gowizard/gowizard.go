// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tredoe/dat/question"
	"github.com/tredoe/dat/valid"
	"github.com/tredoe/goutil/cmdutil"
	"github.com/tredoe/wizard"
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
	flag.Var(&fImportPath, "import", "base of import path (i.e. github.com/tredoe); colon-separated list")
}

// * * *

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: gowizard -i [-cfg]

`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	cfg, err := initConfig()
	if err != nil {
		cmdutil.Fatal(err)
	}
	if cfg == nil {
		os.Exit(0)
	}

	p, err := wizard.NewProject(cfg)
	if err != nil {
		cmdutil.Fatal(err)
	}

	if err = p.Create(); err != nil {
		cmdutil.Fatal(err)
	}
}

// * * *

// initConfig loads configuration from flags and user configuration.
// Returns the configuration to nil when it is used the flag "cfg".
func initConfig() (*wizard.Conf, error) {
	var (
		fName    = flag.String("name", "", "project name")
		fLicense = flag.String("license", "", "license covering the program")
		fAuthor  = flag.String("author", "", "author's name")
		fEmail   = flag.String("email", "", "author's email")
		fVCS     = flag.String("vcs", "", "version control system")
		fOrg     = flag.String("org", "", "organization holder of the copyright")

		fConfig      = flag.Bool("cfg", false, "add the user configuration file")
		fInteractive = flag.Bool("i", false, "interactive mode")

		// Listing
		fListLicense = flag.Bool("ll", false, "list the available licenses (for license flag)")
		fListVCS     = flag.Bool("lv", false, "list the available version control systems (for vcs flag)")
	)

	// == Parse the flags
	flag.Usage = usage
	flag.Parse()

	if flag.NFlag() == 0 {
		usage()
	}

	// == Listing
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

	if *fListLicense || *fListVCS {
		return nil, nil
	}

	var err error
	cfg := &wizard.Conf{
		Program:     *fName,
		License:     *fLicense,
		Author:      *fAuthor,
		Email:       *fEmail,
		VCS:         *fVCS,
		ImportPaths: fImportPath,
		Org:         *fOrg,
	}

	// Get configuration per user, if any.
	if !*fConfig {
		if err = cfg.UserConfig(); err != nil {
			return nil, err
		}
	}
	cfg.Project = *fName

	if err = cfg.PreCheck(*fInteractive, *fConfig); err != nil {
		return nil, err
	}
	// Interactive mode
	if *fInteractive {
		if err = interactive(cfg, *fConfig); err != nil {
			return nil, err
		}
	}
	if err = cfg.PostCheck(*fInteractive, *fConfig); err != nil {
		return nil, err
	}

	// Add configuration.
	if *fConfig && *fInteractive {
		cfg.AddConfig()
		return nil, nil
	}

	return cfg, nil
}

// interactive uses the interactive mode.
func interactive(c *wizard.Conf, addConfig bool) (err error) {
	var sFlags []string
	var msg string

	// == Sorted flags
	if addConfig {
		msg = "New configuration"
		sFlags = []string{"author", "email", "license", "vcs", "import", "org"}
	} else {
		msg = "New project"
		sFlags = []string{
			"name",
			"org",
			"author",
			"email",
			"license",
			"vcs",
			"import",
		}
	}

	q := question.New()
	defer func() {
		err2 := q.Restore()
		if err2 != nil && err == nil {
			err = err2
		}
	}()

	fmt.Printf("\n  = Gowizard :: %s\n\n", msg)

	for _, k := range sFlags {
		f := flag.Lookup(k)

		if strings.Contains(f.Usage, ";") {
			f.Usage = strings.SplitN(f.Usage, ";", 2)[0]
		}
		f.Usage = strings.ToUpper(string(f.Usage[0])) + f.Usage[1:]

		switch k {
		case "name":
			q.Prompt(f.Usage,
				valid.String().SetStringCheck(valid.S_Strict),
				valid.NewScheme().Required().SetDefault(c.Project),
			)
			c.Project, err = q.ReadString()

			if err = c.SetNames(); err != nil {
				return err
			}
		case "org":
			isOrg := true

			if c.Org == "" {
				q.Prompt("Is for an organization?",
					valid.Bool(),
					valid.NewScheme().SetDefault(false),
				)
				isOrg, err = q.ReadBool()
			}

			if isOrg {
				q.Prompt(f.Usage,
					valid.String(),
					valid.NewScheme().Required().SetDefault(c.Org),
				)
				c.Org, err = q.ReadString()
			}
		case "author":
			q.Prompt(f.Usage,
				valid.String().SetStringCheck(valid.S_Strict),
				valid.NewScheme().Required().SetDefault(c.Author),
			)
			c.Author, err = q.ReadString()
		case "email":
			q.Prompt(f.Usage,
				valid.Email(),
				valid.NewScheme().Required().SetDefault(c.Email),
			)
			c.Email, err = q.ReadString()
		case "license":
			q.Prompt(f.Usage,
				valid.String(),
				valid.NewScheme().SetDefault(wizard.ListLowerLicense[c.License]),
			)
			c.License, err = q.ChoiceString(wizard.ListLicenseSorted)
			// It is got in upper case
			c.License = strings.ToLower(c.License)
		case "vcs":
			q.Prompt(f.Usage,
				valid.String(),
				valid.NewScheme().SetDefault(c.VCS),
			)
			c.VCS, err = q.ChoiceString(wizard.ListVCSsorted)
		case "import":
			if addConfig {
				q.Prompt(f.Usage,
					valid.String(),
					valid.NewScheme().Required(),
				)
				c.ImportPaths, err = q.ReadStringSlice()

			} else if len(c.ImportPaths) == 0 {
				tmp := ""

				q.Prompt(f.Usage,
					valid.String(),
					nil,
				)
				tmp, err = q.ReadString()

				if tmp != "" {
					c.ImportPaths = make([]string, 1)
					c.ImportPaths[0] = tmp
				}

			} else {
				q.Prompt(f.Usage,
					valid.String(),
					valid.NewScheme().SetDefault(c.ImportPaths[0]),
				)
				c.ImportPaths[0], err = q.ChoiceString(c.ImportPaths)
			}
		}

		if err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}
