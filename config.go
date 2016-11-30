// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package wizard

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/tredoe/dat/valid"
	"gopkg.in/yaml.v1"
)

// Conf represents the configuration of the project.
type Conf struct {
	Project     string
	Program     string // to lower case
	License     string
	Author      string
	Email       string
	VCS         string
	Org         string // the author develops the program for an organization
	Import      string // To get data from user configuration; then is sent to ImportPaths
	ImportPaths []string

	// To pass to templates
	ImportPath    string
	Comment       string
	FullLicense   string
	GNUextra      string
	ProjectHeader string
	Year          int
}

// SetNames sets names for both project and program.
//
// A program name is named as the project name but in lower case.
func (c *Conf) SetNames() error {
	c.Project = strings.TrimSpace(c.Project)
	c.Program = strings.ToLower(c.Project)

	// To remove them from the program name, if any.
	reStart1 := regexp.MustCompile(`^go-`)
	reStart2 := regexp.MustCompile(`^go`)
	reEnd1 := regexp.MustCompile(`-go$`)
	reEnd2 := regexp.MustCompile(`go$`)

	if reStart1.MatchString(c.Program) {
		c.Program = reStart1.ReplaceAllString(c.Program, "")
	} else if reStart2.MatchString(c.Program) {
		c.Program = reStart2.ReplaceAllString(c.Program, "")
	} else if reEnd1.MatchString(c.Program) {
		c.Program = reEnd1.ReplaceAllString(c.Program, "")
	} else if reEnd2.MatchString(c.Program) {
		c.Program = reEnd2.ReplaceAllString(c.Program, "")
	}

	return nil
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

	cfg.ImportPath = strings.Join(cfg.ImportPaths, ":")

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
	switch info, err := os.Stat(pathUserConfig); {
	case os.IsNotExist(err):
		return nil
	case !info.Mode().IsRegular():
		return fmt.Errorf("expected regular file: %s", _USER_CONFIG)
	}

	data, err := ioutil.ReadFile(pathUserConfig)
	if err != nil {
		return err
	}

	cfg := Conf{}
	if err = yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return fmt.Errorf("error parsing configuration: %s", err)
	}

	// == Get values

	if c.Org == "" && cfg.Org != "" {
		c.Org = cfg.Org
	}
	if c.Author == "" && cfg.Author != "" {
		c.Author = cfg.Author
	}
	if c.Email == "" && cfg.Email != "" {
		c.Email = cfg.Email
	}
	if c.License == "" && cfg.License != "" {
		c.License = cfg.License
	}
	if c.VCS == "" && cfg.VCS != "" {
		c.VCS = cfg.VCS
	}
	if len(c.ImportPaths) == 0 && cfg.Import != "" {
		c.ImportPaths = strings.Split(cfg.Import, ":")
	}

	return nil
}

// == Checking
//

// PreCheck checks the initial configuration, setting some values.
func (c *Conf) PreCheck(interactive, addConfig bool) error {
	var required []string

	if !interactive {
		if !addConfig {
			if err := c.SetNames(); err != nil {
				return err
			}
		}

		// == Necessary fields
		if addConfig {
			required = []string{c.Author, c.Email, c.License, c.VCS}
		} else {
			required = []string{c.Project, c.Program, c.License, c.Author,
				c.Email, c.VCS}
		}

		for _, v := range required {
			if v == "" {
				return errors.New("missing required field")
			}
		}
	}

	// == Maps

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

// PostCheck checks and sets to be run after of.get configuration.
func (c *Conf) PostCheck(interactive, addConfig bool) error {
	// Email
	if !interactive {
		_, err := valid.Email().
			SetScheme(valid.NewScheme().Required()).
			Check(c.Email)
		if err != nil {
			return err
		}
	}

	c.Email = fmt.Sprintf("%s <%s>",
		c.Author, strings.Replace(c.Email, "@", " AT ", -1))

	// Adds extra fields to pass to templates.
	if !addConfig {
		c.ProjectHeader = strings.Repeat(_HEADER_CHAR, len(c.Project))

		if c.License != "none" {
			c.FullLicense = ListLicense[ListLowerLicense[c.License]]
		}
	}

	return nil
}
