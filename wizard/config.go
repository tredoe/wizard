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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kless/goconfig/config"
)

// Represents the configuration of the project.
type Conf struct {
	ProjecType  string
	ProjectName string
	PackageName string
	License     string
	Author      string
	Email       string
	VCS         string
	OrgName     string // the author develops the program for an organization
	IsForOrg    bool

	// To pass to templates
	Comment       string
	FullLicense   string
	GNUextra      string
	ProjectHeader string
	HasCopyright  bool
	IsCmdProject  bool
	IsCgoProject  bool
	Year          int
}

// TODO: to be used by a GUI, in the first there is to get a new type Conf.
// Then, it is passed to ExtraConfig().

// Checks values in configuration and add extra fields to pass to templates.
func ExtraConfig(cfg *Conf) error {
	// === Checking
	if err := checkAtCreate(cfg); err != nil {
		return err
	}

	// === Extra for templates
	cfg.ProjectHeader = strings.Repeat(_CHAR_HEADER, len(cfg.ProjectName))

	if cfg.License != "none" {
		cfg.FullLicense = ListLicense[cfg.License][1]
	}
	if cfg.License != "cc0" {
		cfg.HasCopyright = true
	}
	if cfg.ProjecType == "cgo" {
		cfg.IsCgoProject = true
	}
	// For the Makefile
	if cfg.ProjecType == "cmd" {
		cfg.IsCmdProject = true
	}

	return nil
}

// Sets names for both project and package.
func SetNames(c *Conf) {
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

// Loads configuration per user, if any.
func UserConfig(c *Conf) error {
	home, err := os.Getenverror("HOME")
	if err != nil {
		return fmt.Errorf("no variable HOME: %s", err)
	}

	pathUserConfig := filepath.Join(home, _USER_CONFIG)

	// To know if the file exist.
	switch info, err := os.Stat(pathUserConfig); {
	case err != nil:
		return fmt.Errorf("user configuration does not exist: %s", err)
	case !info.IsRegular():
		return fmt.Errorf("not a file: %s", _USER_CONFIG)
	}

	cfg, err := config.ReadDefault(pathUserConfig)
	if err != nil {
		return fmt.Errorf("error parsing configuration: %s", err)
	}

	// === Get values
	var errKeys []string
	ok := true

	if c.OrgName == "" {
		c.OrgName, err = cfg.String("DEFAULT", "org-name")
		if err != nil {
			ok = false
			errKeys = append(errKeys, "org-name")
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
func checkAtCreate(c *Conf) error {
	ok := true

	// === Necessary fields
	if c.ProjecType == "" || c.ProjectName == "" || c.License == "" ||
		c.Author == "" || c.VCS == "" {
		return errors.New("missing required fields to create project")
	}

	// === Project type
	c.ProjecType = strings.ToLower(c.ProjecType)
	if _, present := ListProject[c.ProjecType]; !present {
		fmt.Fprintf(os.Stderr, "unavailable project type: %q\n", c.ProjecType)
		ok = false
	}

	// === VCS
	c.VCS = strings.ToLower(c.VCS)
	if _, present := ListVCS[c.VCS]; !present {
		fmt.Fprintf(os.Stderr, "unavailable version control system: %q\n", c.VCS)
		ok = false
	}

	// === License
	c.License = strings.ToLower(c.License)
	licenseOk := checkLicense(c.License)

	if !ok || !licenseOk {
		return errors.New("required field")
	}
	return nil
}

// Checks license.
func checkLicense(name string) bool {
	if _, ok := ListLicense[name]; !ok {
		fmt.Fprintf(os.Stderr, "unavailable license: %q\n", name)
		return false
	}
	return true
}
