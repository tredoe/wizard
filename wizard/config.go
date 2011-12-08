// Copyright 2010  The "GoWizard" Authors
//
// Use of this source code is governed by the BSD 2-Clause License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package wizard

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
	HasCopyright  bool
	IsCmd         bool
	IsCgo         bool
	Year          int

	// Is a new project? If it is not then it is created a new program
	IsNewProject bool
}

// TODO: to be used by a GUI, in the first there is to get a new type Conf.
// Then, it is passed to ExtraConfig().

// Checks values in configuration and add extra fields to pass to templates.
func (cfg *Conf) ExtraConfig() error {
	// === Checking
	if err := cfg.check(); err != nil {
		return err
	}

	// === Extra for templates
	cfg.ProjectHeader = strings.Repeat(_CHAR_HEADER, len(cfg.Project))

	if cfg.License != "none" {
		cfg.FullLicense = ListLicense[ListLowerLicense[cfg.License]]
	}
	if cfg.License != "unlicense" && cfg.License != "cc0" {
		cfg.HasCopyright = true
	}
	if cfg.Type == "cgo" {
		cfg.IsCgo = true
	}
	// For the Makefile
	if cfg.Type == "cmd" {
		cfg.IsCmd = true
	}

	return nil
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

	envHome := os.Getenv("HOME")
	if envHome == "" {
		return errors.New("could not add user configuration file because $HOME is not set")
	}

	file, err := createFile(filepath.Join(envHome, _USER_CONFIG))
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
	home, err := os.Getenverror("HOME")
	if err != nil {
		return fmt.Errorf("no variable HOME: %s", err)
	}

	pathUserConfig := filepath.Join(home, _USER_CONFIG)

	// To know if the file exist.
	switch stat, err := os.Stat(pathUserConfig); {
	case err != nil: // not exist
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

// Checks at creating project.
func (c *Conf) check() error {
	ok := true

	// === Necessary fields
	if c.Type == "" || c.Program == "" || c.License == "" {
		return errors.New("missing required fields")
	}

	if c.IsNewProject {
		if c.Author == "" || c.VCS == "" {
			return errors.New("missing required fields to create project")
		}

		// === VCS
		c.VCS = strings.ToLower(c.VCS)
		if _, present := ListVCS[c.VCS]; !present {
			fmt.Fprintf(os.Stderr, "unavailable version control system: %q\n", c.VCS)
			ok = false
		}
	}

	// === Project type
	c.Type = strings.ToLower(c.Type)
	if _, present := ListType[c.Type]; !present {
		fmt.Fprintf(os.Stderr, "unavailable project type: %q\n", c.Type)
		ok = false
	}

	// === License
	c.License = strings.ToLower(c.License)

	// * * *

	if !ok {
		return errors.New("required field")
	}
	if err := CheckLicense(c.License); err != nil {
		return err
	}
	return nil
}

// Checks license.
func CheckLicense(name string) error {
	if _, ok := ListLowerLicense[name]; !ok {
		return fmt.Errorf("unavailable license: %s", name)
	}
	return nil
}
