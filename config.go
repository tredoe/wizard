// Copyright 2010  The "Gowizard" Authors
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at
// http://mozilla.org/MPL/2.0/.

package wizard

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/kless/goconfig/config"
)

// Conf represents the configuration of the project.
type Conf struct {
	Type    string
	Project string
	Program string // to lower case
	License string
	Author  string
	Email   string
	VCS     string
	Org     string // the author develops the program for an organization

	// To pass to templates
	Comment       string
	FullLicense   string
	GNUextra      string
	ProjectHeader string
	Owner         string // organization or project's Authors; for patents file
	IsUnlicense   bool
	IsCmd         bool
	IsCgo         bool
	Year          int

	// Is a new project? If it is not then it is created a new program
	IsNewProject bool
}

// TODO: to be used by a GUI, in the first there is to get a new type Conf.
// Then, it is passed to ExtraConfig().

// AddTemplateData adds extra fields to pass to templates.
func (c *Conf) AddTemplateData() {
	if c.IsNewProject {
		c.ProjectHeader = strings.Repeat(_HEADER_CHAR, len(c.Project))
	}

	if c.License != "none" {
		c.FullLicense = fmt.Sprintf("[%s](%s)",
			ListLicense[ListLowerLicense[c.License]], ListLicenseURL[c.License])
	}
	if c.License == "unlicense" {
		c.IsUnlicense = true
	}

	if c.Type == "cgo" {
		c.IsCgo = true
	}
	if c.Type == "cmd" {
		c.IsCmd = true
	}
}

// SetNames sets names for both project and program.
//
// A program name is named as the project name but in lower case; and if it is
// not a command then it is removed the prefix or suffix related to "go", if any.
func (c *Conf) SetNames(addProgram bool) {
	if addProgram {
		c.Program = strings.ToLower(strings.TrimSpace(c.Program))

		// The first line of Readme file has the project name.
		if f, err := os.Open(_README); err == nil {
			buf := bufio.NewReader(f)
			defer f.Close()

			if line, _, err := buf.ReadLine(); err == nil {
				c.Project = string(line)
			}
		}
	} else {
		c.Project = strings.TrimSpace(c.Project)
		c.Program = strings.ToLower(c.Project)
	}

	// The program name is not changed in commands.
	if c.Type == "cmd" {
		return
	}

	// To remove them from the program name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd := regexp.MustCompile(`-go$`)

	if reStart1.MatchString(c.Program) {
		c.Program = reStart1.ReplaceAllString(c.Program, "")
	} else if reStart2.MatchString(c.Program) {
		c.Program = reStart2.ReplaceAllString(c.Program, "")
	} else if reEnd.MatchString(c.Program) {
		c.Program = reEnd.ReplaceAllString(c.Program, "")
	}
}

//
// == User configuration

// AddConfig creates the user configuration file.
func (cfg *Conf) AddConfig() error {
	tmpl := template.Must(template.New("Config").Parse(tmplUserConfig))

	home := os.Getenv("HOME")
	if home == "" {
		return errors.New("could not add user configuration file because $HOME is not set")
	}

	file, err := createFile(filepath.Join(home, _USER_CONFIG))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("execution failed: %s", err)
	}
	return nil
}

// UserConfig loads configuration per user, if any.
func (c *Conf) UserConfig() error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("environment variable $HOME is not set")
	}

	pathUserConfig := filepath.Join(home, _USER_CONFIG)

	// To know if the file exist.
	switch stat, err := os.Stat(pathUserConfig); {
	case os.IsNotExist(err):
		return nil
	case stat.Mode()&os.ModeType != 0:
		return fmt.Errorf("expected file: %s", _USER_CONFIG)
	}

	cfg, err := config.ReadDefault(pathUserConfig)
	if err != nil {
		return fmt.Errorf("error parsing configuration: %s", err)
	}

	// == Get values
	var errKeys []string
	ok := true

	if c.Org == "" {
		c.Org, err = cfg.String("DEFAULT", "org")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "org")
		}
	}
	if c.Author == "" {
		c.Author, err = cfg.String("DEFAULT", "author")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "author")
		}
	}
	if c.Email == "" {
		c.Email, err = cfg.String("DEFAULT", "email")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "email")
		}
	}
	if c.License == "" {
		c.License, err = cfg.String("DEFAULT", "license")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "license")
		}
	}
	if c.VCS == "" {
		c.VCS, err = cfg.String("DEFAULT", "vcs")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "vcs")
		}
	}

	if !ok {
		return fmt.Errorf("error at user configuration: %s\n",
			strings.Join(errKeys, ","))
	}
	return nil
}

// == Checking
//

// CheckAndSetNames checks values in the configuration, and set names of both
// project and program whether it is not on interactive mode.
func (c *Conf) CheckAndSetNames(interactive, addConfig, addProgram bool) error {
	var required []string

	if !interactive {
		if !addConfig {
			c.SetNames(addProgram)
		}

		// == Necessary fields
		if addConfig {
			required = []string{c.Author, c.Email, c.License, c.VCS}
		} else if addProgram {
			required = []string{c.Type, c.Program, c.License}
		} else {
			required = []string{c.Type, c.Project, c.Program, c.License, c.Author,
				c.Email, c.VCS}
		}

		for _, v := range required {
			if v == "" {
				return errors.New("missing required field")
			}
		}
	}

	// == Maps

	// Project type
	if c.Type != "" {
		c.Type = strings.ToLower(c.Type)

		if _, ok := ListType[c.Type]; !ok {
			return fmt.Errorf("unavailable project type: %q", c.Type)
		}
	}

	// License
	if c.License != "" {
		c.License = strings.ToLower(c.License)

		if _, ok := ListLowerLicense[c.License]; !ok {
			return fmt.Errorf("unavailable license: %q", c.License)
		}
	}

	// VCS
	if c.VCS != "" {
		c.VCS = strings.ToLower(c.VCS)

		if _, ok := ListVCS[c.VCS]; !ok {
			return fmt.Errorf("unavailable VCS: %q", c.VCS)
		}
	}

	return nil
}
