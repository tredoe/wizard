// Copyright 2010  The "GoWizard" Authors
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

package wizard

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// To get the project's creation year
var (
	reCopyright = regexp.MustCompile(`(Copyright)[ ]+[0-9]{4}[ ].*Authors`)
	reCopyleft  = regexp.MustCompile(`(Written in)[ ]+[0-9]{4}[ ].*Authors`)
	reYear      = regexp.MustCompile(`[0-9]{4}`)
)

// Copyright and licenses
const (
	tmplCopyright = `Copyright {{.Year}}  The "{{.Project}}" Authors`
	tmplCopyleft  = `Written in {{.Year}} by the "{{.Project}}" Authors`

	tmplNone = `{{.Comment}} {{template "Copyright" .}}
`

	tmplApache = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} Licensed under the Apache License, Version 2.0 (the "License");
{{.Comment}} you may not use this file except in compliance with the License.
{{.Comment}} You may obtain a copy of the License at
{{.Comment}}
{{.Comment}}     http://www.apache.org/licenses/LICENSE-2.0
{{.Comment}}
{{.Comment}} Unless required by applicable law or agreed to in writing, software
{{.Comment}} distributed under the License is distributed on an "AS IS" BASIS,
{{.Comment}} WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
{{.Comment}} See the License for the specific language governing permissions and
{{.Comment}} limitations under the License.
`

	tmplBSD = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} Use of this source code is governed by the {{.FullLicense}}
{{.Comment}} that can be found in the LICENSE file.
{{.Comment}}
{{.Comment}} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{{.Comment}} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{{.Comment}} for more details.
`

	tmplGNU = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} This program is free software: you can redistribute it and/or modify
{{.Comment}} it under the terms of the GNU {{with .GNUextra}}{{.}} {{end}}General Public License as published by
{{.Comment}} the Free Software Foundation, either version 3 of the License, or
{{.Comment}} (at your option) any later version.
{{.Comment}}
{{.Comment}} This program is distributed in the hope that it will be useful,
{{.Comment}} but WITHOUT ANY WARRANTY; without even the implied warranty of
{{.Comment}} MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
{{.Comment}} GNU {{with .GNUextra}}{{.}} {{end}}General Public License for more details.
{{.Comment}}
{{.Comment}} You should have received a copy of the GNU {{with .GNUextra}}{{.}} {{end}}General Public License
{{.Comment}} along with this program.  If not, see <http://www.gnu.org/licenses/>.
`

	tmplCC0 = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} To the extent possible under law, the author(s) have waived all copyright
{{.Comment}} and related or neighboring rights to this work to the public domain worldwide.
{{.Comment}} This software is distributed without any warranty.
{{.Comment}}
{{.Comment}} You should have received a copy of the CC0 Public Domain Dedication along
{{.Comment}} with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
`

	tmplUnlicense = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} To the extent possible under law, the author(s) have waived all copyright
{{.Comment}} and related or neighboring rights to this software to the public domain worldwide.
{{.Comment}} This software is distributed without any warranty.
{{.Comment}}
{{.Comment}} You should have received a copy of the Unlicense along with this software.
{{.Comment}} If not, see <http://unlicense.org/>.
`
)

// For source code files
const (
	tmplCmd      = `{{template "Header" .}}
package main

import (

)

func main() {

}
`
	tmplPkg      = `{{template "Header" .}}
package {{.Program}}
{{if .IsCgo}}
import "C"{{end}}
import (

)


`
	tmplTest     = `{{template "Header" .}}
package {{.Program}}

import (
	"testing"
)

func Test(t *testing.T) {

}
`
)

// User configuration
const tmplUserConfig = `[DEFAULT]
org: {{.Org}}
author: {{.Author}}
email: {{.Email}}
license: {{.License}}
vcs: {{.VCS}}
`

// === File ignore for VCS
const hgIgnoreTop = "syntax: glob\n"

var tmplIgnore = `# Generic
*~
[._]*

# Go
*.[ao]
*.[568vq]
[568vq].out
main

# Cgo
*.cgo*
*.so
`

// Renders the template "src", creating a file in "dst".
func (p *project) parseFromFile(dst, src string) error {
	file, err := createFile(dst)
	if err != nil {
		return err
	}

	p.tmpl, err = p.tmpl.ParseFiles(src)
	if err != nil {
		fmt.Errorf("parsing error: %s", err)
	}

	if err = p.tmpl.ExecuteTemplate(file, filepath.Base(src), p.cfg); err != nil {
		fmt.Errorf("execution failed: %s", err)
	}

	return nil
}

// Renders the template "tmplName" in "set" to the file "dst".
func (p *project) parseFromVar(dst string, tmplName string) error {
	file, err := createFile(dst)
	if err != nil {
		return err
	}

	if err = p.tmpl.ExecuteTemplate(file, tmplName, p.cfg); err != nil {
		fmt.Errorf("execution failed: %s", err)
	}
	return nil
}

// Parses the license header.
// "charComment" is the character used to comment in code files.
func (p *project) parseLicense(charComment string) {
	var tmplHeader string

	licenseName := strings.Split(p.cfg.License, "-")[0]
	p.cfg.Comment = charComment

	// If "year" is nil then gets the actual year.
	if p.cfg.Year == 0 {
		p.cfg.Year = time.Now().Year()
	}

	switch licenseName {
	case "apache":
		tmplHeader = tmplApache
	case "bsd":
		tmplHeader = tmplBSD
	case "cc0":
		tmplHeader = tmplCC0
	case "gpl", "lgpl", "agpl":
		tmplHeader = tmplGNU

		if licenseName == "agpl" {
			p.cfg.GNUextra = "Affero"
		} else if licenseName == "lgpl" {
			p.cfg.GNUextra = "Lesser"
		}
	case "unlicense":
		tmplHeader = tmplUnlicense
	case "none":
		tmplHeader = tmplNone
	}

	p.tmpl = template.Must(p.tmpl.New("Header").Parse(tmplHeader))

	if licenseName != "unlicense" && licenseName != "cc0" {
		p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplCopyright))
	} else {
		p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplCopyleft))
	}
}

// Parses the templates for the project.
func (p *project) parseProject() {
	if p.cfg.Type == "cmd" {
		p.tmpl = template.Must(p.tmpl.New("Cmd").Parse(tmplCmd))
	} else {
		p.tmpl = template.Must(p.tmpl.New("Pkg").Parse(tmplPkg))
		p.tmpl = template.Must(p.tmpl.New("Test").Parse(tmplTest))
	}
}
