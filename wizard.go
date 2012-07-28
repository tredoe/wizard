// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package wizard allows to create the base of new Go projects and add new
// packages or commands to the project.
package wizard

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const (
	// Permissions
	_DIR_PERM  = 0755
	_FILE_PERM = 0644

	_COMMENT_CHAR = "//" // For comments in source code files
	_HEADER_CHAR  = "="  // Header under the project name

	// Subdirectory where is installed through "go get"
	_DATA_PATH = "github.com/kless/wizard/data"

	_DOC_DIR      = "Doc"
	_INSTALL_DIR  = "Installer"
	_INSTALL_FILE = "install.go"
	_README       = "README.md"
	_USER_CONFIG  = ".gowizard" // Configuration file per user
)

// Project types
var (
	ListTypeSorted = []string{"cgo", "cmd", "pkg"}

	ListType = map[string]string{
		"cmd": "command line program",
		"pkg": "package",
		"cgo": "package that calls C code",
	}
)

// Version control systems (VCS)
var (
	ListVCSsorted = []string{"bzr", "git", "hg", "none"}

	ListVCS = map[string]string{
		"bzr":  "Bazaar",
		"git":  "Git",
		"hg":   "Mercurial",
		"none": "none",
	}

	/*// VCS configuration files
	listConfigVCS = map[string]string{
		"bzr": ".bzr/branch/branch.conf",
		"git": ".git/config",
		"hg":  ".hg/hgrc",
	}*/
)

// Available licenses
var (
	ListLicenseSorted = []string{"AGPL", "Apache", "CC0", "GPL", "MPL", "none"}

	ListLicense = map[string]string{
		"AGPL":   "GNU Affero General Public License, version 3 or later",
		"Apache": "Apache License, version 2.0",
		"CC0":    "Creative Commons CC0, version 1.0 Universal",
		"GPL":    "GNU General Public License, version 3 or later",
		"MPL":    "Mozilla Public License, version 2.0",
		"none":   "proprietary license",
	}
	ListLowerLicense = map[string]string{
		"agpl":   "AGPL",
		"apache": "Apache",
		"cc0":    "CC0",
		"gpl":    "GPL",
		"mpl":    "MPL",
		"none":   "none",
	}

	listLicenseURL = map[string]string{
		"agpl":   "http://www.gnu.org/licenses/agpl.html",
		"apache": "http://www.apache.org/licenses/LICENSE-2.0",
		"cc0":    "http://creativecommons.org/publicdomain/zero/1.0/",
		"gpl":    "http://www.gnu.org/licenses/gpl.html",
		"mpl":    "http://mozilla.org/MPL/2.0/",
	}
	listLicenseFaq = map[string]string{
		"agpl":   "http://www.gnu.org/licenses/gpl-faq.html",
		"apache": "http://www.apache.org/foundation/license-faq.html",
		"cc0":    "http://creativecommons.org/about/cc0",
		"gpl":    "http://www.gnu.org/licenses/gpl-faq.html",
		"mpl":    "http://www.mozilla.org/MPL/2.0/FAQ.html",
	}
)

// project represents all information to create a project.
type project struct {
	dataDir string             // directory with templates
	tmpl    *template.Template // set of templates
	cfg     *Conf
}

// NewFile creates a new file in the current directory.
// If addInstall is used then the name is set to a value fixed.
func NewFile(addGo, addCgo, addTest, addInstall bool, name ...string) error {
	cfg := new(Conf)
	pkg, err := build.ImportDir(".", 0)

	if !addInstall {
		if err == nil {
			cfg.Program = pkg.Name
		} else {
			cfg.Program = "main"
		}
		if pkg.IsCommand() {
			cfg.IsCmd = true
		}
		if addCgo {
			cfg.IsCgo = true
		}
	}

	goFiles, err := filepath.Glob("*.go")
	if err != nil {
		return err
	}
	if len(goFiles) == 0 {
		return errors.New("no found any Go source file in current directory")
	}

	// == Get the header from a Go file
	var file *os.File
	bComment := []byte("//")
	bPackage := []byte("package")
	found := false
	header := ""
	reYear := regexp.MustCompile(`[[:blank:]]\d{4}[[:blank:]]`)

	for _, f := range goFiles {
		if file != nil {
			file.Close()
		}
		file, err = os.Open(f)
		if err != nil {
			return err
		}
		defer file.Close()

		buf := bufio.NewReader(file)
		headerBuf := new(bytes.Buffer)
		isFirst := true

		for {
			line, _, err := buf.ReadLine()
			if err == io.EOF {
				break
			}

			if bytes.HasPrefix(line, bComment) {
				if isFirst {
					line = reYear.ReplaceAll(line, []byte(fmt.Sprintf(" %d ", time.Now().Year())))
					found = true
					isFirst = false
				}
				headerBuf.Write(line)
				headerBuf.WriteRune('\n')

			} else if headerBuf.Len() != 0 || bytes.HasPrefix(line, bPackage) {
				break
			}
		}

		if headerBuf.Len() != 0 {
			header = headerBuf.String()
			break
		}
	}

	if !found {
		return errors.New("any Go source file has a copyright & license notice")
	}
	proj := &project{"", new(template.Template), cfg}
	proj.tmpl = template.Must(proj.tmpl.New("Header").Parse(header))

	// == Render files

	if addGo || addCgo {
		proj.tmpl = template.Must(proj.tmpl.New("Go").Parse(tmplGo))
	}
	if addTest {
		proj.tmpl = template.Must(proj.tmpl.New("Test").Parse(tmplTest))
	}

	for _, n := range name {
		if addGo || addCgo {
			if err = proj.parseFromVar(n+".go", "Go"); err != nil {
				return err
			}
		}
		if addTest {
			if err = proj.parseFromVar(n+"_test.go", "Test"); err != nil {
				return err
			}
		}
	}

	if addInstall {
		if _, err = os.Stat(_README); os.IsNotExist(err) {
			return errors.New("file README not found; maybe current directory is not a Go project")
		}
		if err = os.Mkdir(_INSTALL_DIR, _DIR_PERM); err != nil {
			return err
		}

		proj.tmpl = template.Must(proj.tmpl.New(_INSTALL_DIR).Parse(tmplInstall))
		if err = proj.parseFromVar(filepath.Join(_INSTALL_DIR, _INSTALL_FILE),
			_INSTALL_DIR); err != nil {
			return err
		}
	}
	return nil
}

// NewProject initializes information for a new project.
func NewProject(cfg *Conf) (*project, error) {
	// To get the path of the templates directory.
	pkg, err := build.Import(_DATA_PATH, build.Default.GOPATH, build.FindOnly)
	if err != nil {
		return nil, fmt.Errorf("NewProject: data directory not found: %s", err)
	}

	return &project{pkg.Dir, new(template.Template), cfg}, nil
}

// Create creates a new project.
func (p *project) Create() error {
	err := os.Mkdir(p.cfg.Program, _DIR_PERM)
	if err != nil {
		return fmt.Errorf("directory error: %s", err)
	}
	if p.cfg.IsNewProject {
		if err := os.Mkdir(filepath.Join(p.cfg.Program, _DOC_DIR),
			_DIR_PERM); err != nil {
			return fmt.Errorf("directory error: %s", err)
		}
	}

	p.parseLicense(_COMMENT_CHAR)
	p.parseProject()

	// == Render project files
	err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+".go", "Go")
	if err != nil {
		return err
	}
	if p.cfg.Type != "cmd" {
		err = p.parseFromVar(filepath.Join(p.cfg.Program, p.cfg.Program)+"_test.go", "Test")
		if err != nil {
			return err
		}
	}

	// == -add flag
	if !p.cfg.IsNewProject {
		if err = p.addLicense("."); err != nil { // actual directory
			return err
		}
		// Append the command name into the ignore file.
		if p.cfg.IsCmd {
			ignoreFile, err := filepath.Glob(".*ignore")
			if err != nil {
				return err
			}
			if len(ignoreFile) == 0 {
				return nil
			}

			file, err := os.OpenFile(ignoreFile[0], os.O_WRONLY|os.O_APPEND, 0)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err = file.WriteString(p.cfg.Program + "/" + p.cfg.Program + "\n"); err != nil {
				return err
			}
		}
		return nil
	}
	// ==

	if len(p.cfg.ImportPaths) != 0 {
		p.cfg.ImportPath = strings.Replace(p.cfg.ImportPaths[0], "$", p.cfg.Program, 1)
	}

	// License file
	if err = p.addLicense(p.cfg.Program); err != nil {
		return err
	}

	// Render common files
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, _DOC_DIR, "CONTRIBUTORS.md"),
		"Contributors"); err != nil {
		return err
	}
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, _DOC_DIR, "NEWS.md"),
		"News"); err != nil {
		return err
	}
	if err = p.parseFromVar(filepath.Join(p.cfg.Program, _README),
		"Readme"); err != nil {
		return err
	}

	// The file AUTHORS is for copyright holders.
	if p.cfg.License != "cc0" {
		if err = p.parseFromVar(filepath.Join(p.cfg.Program, _DOC_DIR, "AUTHORS.md"),
			"Authors"); err != nil {
			return err
		}
	}

	// Add file related to VCS
	if p.cfg.VCS != "none" {
		ignoreFile := "." + p.cfg.VCS + "ignore"
		if err = p.parseFromVar(filepath.Join(p.cfg.Program, ignoreFile),
			"Ignore"); err != nil {
			return err
		}

		// Initialize VCS
		out, err := exec.Command(p.cfg.VCS, "init", p.cfg.Program).CombinedOutput()
		if err != nil {
			return err
		}
		if out != nil {
			fmt.Print(string(out))
		}
	}

	return nil
}

// * * *

// addLicense creates a license file.
func (p *project) addLicense(dir string) error {
	if p.cfg.License == "none" {
		return nil
	}

	license := ListLowerLicense[p.cfg.License]
	licenseDst := filepath.Join(dir, _DOC_DIR, "LICENSE_"+license+".txt")

	// Check if it exist.
	if !p.cfg.IsNewProject {
		if _, err := os.Stat(licenseDst); !os.IsNotExist(err) {
			return nil
		}
	}

	if err := copyFile(licenseDst, filepath.Join(p.dataDir, license+".txt")); err != nil {
		return err
	}
	return nil
}
