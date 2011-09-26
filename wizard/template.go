// Copyright 2010  The "Go-Wizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package wizard

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"template"
	"time"
)

// Copyright and licenses
const (
	tmplCopyright = `Copyright {{.year}}  The "{{.project_name}}" Authors`
	tmplCopyleft  = `Written in {{.year}} by the "{{.project_name}}" Authors`

	tmplBSD = `{{.comment}} {{template "Copyright" .}}
{{.comment}}
{{.comment}} Use of this source code is governed by the {{.full_license}}
{{.comment}} that can be found in the LICENSE file.
{{.comment}}
{{.comment}} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{{.comment}} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{{.comment}} for more details.
`

	tmplApache = `{{.comment}} {{template "Copyright" .}}
{{.comment}}
{{.comment}} Licensed under the Apache License, Version 2.0 (the "License");
{{.comment}} you may not use this file except in compliance with the License.
{{.comment}} You may obtain a copy of the License at
{{.comment}}
{{.comment}}     http://www.apache.org/licenses/LICENSE-2.0
{{.comment}}
{{.comment}} Unless required by applicable law or agreed to in writing, software
{{.comment}} distributed under the License is distributed on an "AS IS" BASIS,
{{.comment}} WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
{{.comment}} See the License for the specific language governing permissions and
{{.comment}} limitations under the License.
`

	tmplGNU = `{{.comment}} {{template "Copyright" .}}
{{.comment}}
{{.comment}} This program is free software: you can redistribute it and/or modify
{{.comment}} it under the terms of the GNU {{with .Affero}}{{.}} {{end}}{{with .Lesser}}{{.}} {{end}}General Public License as published by
{{.comment}} the Free Software Foundation, either version 3 of the License, or
{{.comment}} (at your option) any later version.
{{.comment}}
{{.comment}} This program is distributed in the hope that it will be useful,
{{.comment}} but WITHOUT ANY WARRANTY; without even the implied warranty of
{{.comment}} MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
{{.comment}} GNU {{with .Affero}}{{.}} {{end}}{{with .Lesser}}{{.}} {{end}}General Public License for more details.
{{.comment}}
{{.comment}} You should have received a copy of the GNU {{with .Affero}}{{.}} {{end}}{{with .Lesser}}{{.}} {{end}}General Public License
{{.comment}} along with this program.  If not, see <http://www.gnu.org/licenses/>.
`

	tmplNone = `{{.comment}} {{template "Copyright" .}}
`

	tmplCC0 = `{{.comment}} {{template "Copyright" .}}
{{.comment}}
{{.comment}} To the extent possible under law, the author(s) have waived all copyright
{{.comment}} and related or neighboring rights to this software to the public domain worldwide.
{{.comment}} This software is distributed without any warranty.
{{.comment}}
{{.comment}} You should have received a copy of the CC0 Public Domain Dedication along
{{.comment}} with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
`
)

// For source code files
const (
	tmplCmd      = `{{template "Header" .}}
package main

import (

)


`
	tmplPkg      = `{{template "Header" .}}
package {{.package_name}}
{{if .is_cgo_project}}
import "C"{{end}}
import (

)


`
	tmplTest     = `{{template "Header" .}}
package {{.package_name}}

import (
	"testing"
)

func Test(t *testing.T) {

}

`
	tmplMakefile = `include $(GOROOT)/src/Make.inc

TARG={{if .is_cmd_project}}{{else}}<< IMPORT PATH >>/{{end}}{{.package_name}}
GOFILES=\
	{{.package_name}}.go\

include $(GOROOT)/src/Make.{{if .is_cmd_project}}cmd{{else}}pkg{{end}}

`
)

// User configuration
const tmplUserConfig = `[DEFAULT]
author: {{.author}}
author-email: {{.author_email}}
license: {{.license}}
vcs: {{.vcs}}
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
func (p *project) parseFromFile(dst, src string, useNest bool) {
	file := createFile(dst)

	tmpl, err := template.ParseFile(src)
	if err != nil {
		log.Fatal("parsing error:", err)
	}

	if !useNest {
		if err = tmpl.Execute(file, p.data); err != nil {
			log.Fatal("execution failed:", err)
		}
	} else {
		p.set.Add(tmpl)

		if err = p.set.Execute(file, filepath.Base(src), p.data); err != nil {
			log.Fatal("execution failed:", err)
		}
	}
}

// Renders the template "tmplName" in "set" to the file "dst".
func (p *project) parseFromVar(dst string, tmplName string) {
	file := createFile(dst)

	if err := p.set.Execute(file, tmplName, p.data); err != nil {
		log.Fatal("execution failed:", err)
	}
}

// Parses the templates.
// "charComment" is the character used to comment in code files.
// If "year" is nil then gets the actual year.
func (p *project) parseTemplates(charComment string, year int) {
	var tmplHeader string

	licenseName := strings.Split(p.cfg.license, "-")[0]
	p.data["comment"] = charComment

	if year == 0 {
		p.data["year"] = strconv.Itoa64(time.LocalTime().Year)
	} else {
		p.data["year"] = year
	}

	switch licenseName {
	case "apache":
		tmplHeader = tmplApache
	case "bsd":
		tmplHeader = tmplBSD
	case "cc0":
		tmplHeader = tmplCC0
	case "gpl", "lgpl", "agpl":
		p.data["Affero"] = ""
		p.data["Lesser"] = ""

		if licenseName == "agpl" {
			p.data["Affero"] = "Affero"
		} else if licenseName == "lgpl" {
			p.data["Lesser"] = "Lesser"
		}

		tmplHeader = tmplGNU
	case "none":
		tmplHeader = tmplNone
	}

	// === Add all templates
	if licenseName != "cc0" {
		p.set.Add(template.Must(template.New("Copyright").
			Parse(tmplCopyright)))
	} else {
		p.set.Add(template.Must(template.New("Copyright").
			Parse(tmplCopyleft)))
	}

	if p.cfg.projecType == "cmd" {
		p.set.Add(template.Must(template.New("Cmd").Parse(tmplCmd)))
	} else {
		tPkg := template.Must(template.New("Pkg").Parse(tmplPkg))
		tTest := template.Must(template.New("Test").Parse(tmplTest))
		p.set.Add(tPkg, tTest)
	}

	if p.cfg.addUserConf {
		p.set.Add(template.Must(template.New("Config").Parse(tmplUserConfig)))
	}

	tHeader := template.Must(template.New("Header").Parse(tmplHeader))
	tMakefile := template.Must(template.New("Makefile").Parse(tmplMakefile))
	p.set.Add(tHeader, tMakefile)

	// These tags are not used anymore.
	//for _, t := range []string{"Affero", "comment", "year"} {
	//	p.data[t] = "", false
	//}
}
