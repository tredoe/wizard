// Copyright 2010  The "gowizard" Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gowizard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/kless/goconfig/config"
)

// Represents the configuration of the project.
type Conf struct {
	Type    string
	Project string
	Program string
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
	HasCopyright  bool
	IsUnlicense   bool
	IsCmd         bool
	IsCgo         bool
	Year          int

	// Is a new project? If it is not then it is created a new program
	IsNewProject bool
}

// TODO: to be used by a GUI, in the first there is to get a new type Conf.
// Then, it is passed to ExtraConfig().

// Adds extra fields to pass to templates.
func (c *Conf) AddTemplateData() {
	c.ProjectHeader = strings.Repeat(_CHAR_HEADER, len(c.Project))

	if c.License != "none" {
		c.FullLicense = ListLicense[ListLowerLicense[c.License]]
	}
	if c.License == "unlicense" {
		c.IsUnlicense = true
	} else if c.License != "cc0" {
		c.HasCopyright = true
	}

	if c.Type == "cgo" {
		c.IsCgo = true
	}
	if c.Type == "cmd" {
		c.IsCmd = true
	}
}

// Sets names for both project and package.
func (c *Conf) SetNames(addProgram bool) {
	if addProgram {
		c.Program = strings.ToLower(strings.TrimSpace(c.Program))
		return
	}

	// === To remove them from the project name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd := regexp.MustCompile(`-go$`)

	c.Project = strings.TrimSpace(c.Project)

	// A program is usually named as the project name.
	// It is created removing prefix or suffix related to "go".
	if c.Program == "" {
		c.Program = strings.ToLower(c.Project)

		if reStart1.MatchString(c.Program) {
			c.Program = reStart1.ReplaceAllString(c.Program, "")
		} else if reStart2.MatchString(c.Program) {
			c.Program = reStart2.ReplaceAllString(c.Program, "")
		} else if reEnd.MatchString(c.Program) {
			c.Program = reEnd.ReplaceAllString(c.Program, "")
		}

	} else {
		c.Program = strings.ToLower(strings.TrimSpace(c.Program))
	}
}

//
// === User configuration

// Creates the user configuration file.
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

// Loads configuration per user, if any.
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

	// === Get values
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

//
// === Checking

// Checks values in the configuration.
func (c *Conf) Checking(interactive, addConfig, addProgram bool) error {
	var required []string

	if !interactive {
		if !addConfig {
			c.SetNames(addProgram)
		}

		// === Necessary fields
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

	// === Maps

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
