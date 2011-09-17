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
	"os"
	"strconv"
	"strings"
	"template"
	"time"
)

// For comments in source code files
const _CHAR_CODE_COMMENT = "//"

// Copyright and licenses
const (
	tmplCopyright = `{{define "Copyright"}}{{with .comment}}{{.}} {{end}}Copyright {{.year}}  The "{{.project_name}}" Authors{{end}}`

	tmplBSD = `{{define "Header"}}{{template "Copyright" .}}
{{.comment}}
{{.comment}} Use of this source code is governed by the {{.license}}
{{.comment}} that can be found in the LICENSE file.
{{.comment}}
{{.comment}} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{{.comment}} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{{.comment}} for more details.
{{end}}`

	tmplApache = `{{define "Header"}}{{template "Copyright" .}}
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
{{end}}`

	tmplGNU = `{{define "Header"}}{{template "Copyright" .}}
{{.comment}}
{{.comment}} This program is free software: you can redistribute it and/or modify
{{.comment}} it under the terms of the GNU {{with Affero}}{{.}} {{end}}{{with Lesser}}{{.}} {{end}}General Public License as published by
{{.comment}} the Free Software Foundation, either version 3 of the License, or
{{.comment}} (at your option) any later version.
{{.comment}}
{{.comment}} This program is distributed in the hope that it will be useful,
{{.comment}} but WITHOUT ANY WARRANTY; without even the implied warranty of
{{.comment}} MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
{{.comment}} GNU {{with Affero}}{{.}} {{end}}{{with Lesser}}{{.}} {{end}}General Public License for more details.
{{.comment}}
{{.comment}} You should have received a copy of the GNU {{with Affero}}{{.}} {{end}}{{with Lesser}}{{.}} {{end}}General Public License
{{.comment}} along with this program.  If not, see <http://www.gnu.org/licenses/>.
{{end}}`

	tmplNone = `{{define "Header"}}{{template "Copyright" .}}
{{end}}`

	tmplCC0 = `{{define "Header"}}{{.comment}} To the extent possible under law, Authors have waived all copyright and
{{.comment}} related or neighboring rights to "{{.project_name}}".
{{end}}`
)

// For source code files
const (
	tmplCmd = `{{define "Cmd"}}{{.Header}}
package main

import (

)


{{end}}`

	tmplPac = `{{define "Pkg"}}{{.Header}}
package {{.package_name}}
{{if .project_is_cgo}}
import "C"{{end}}
import (

)


{{end}}`

	tmplTest = `{{define "Test"}}{{.Header}}
package {{.package_name}}

import (
	"testing"
)

func Test(t *testing.T) {

}

{{end}}`
)

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
func (p *project) renderFile(dst, src string) {
	// === Create file.
	file, err := os.Create(dst)
	if err != nil {
		log.Fatal("file error:", err)
	}
	if err = file.Chmod(PERM_FILE); err != nil {
		log.Fatal("file error:", err)
	}

	tmpl, err := template.ParseFile(src)
	if err != nil {
		log.Fatal("parsing error:", err)
	}
	if err = tmpl.Execute(file, p.data); err != nil {
		log.Fatal("execution failed:", err)
	}
}

// Renders the template "tmplName" in "set" to the file "dst".
func (p *project) renderSet(dst string, set *template.Set, tmplName string) {
	// === Create file.
	file, err := os.Create(dst)
	if err != nil {
		log.Fatal("file error:", err)
	}
	if err = file.Chmod(PERM_FILE); err != nil {
		log.Fatal("file error:", err)
	}

	err = set.Execute(file, tmplName, p.data)
	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
}

// Parses the templates.
// "charComment" is the character used to comment in code files.
// If "year" is nil then gets the actual year.
func (p *project) parseTemplates(charComment string, year int) *template.Set {
	var err os.Error
	var fullSet *template.Set
	var tmplHeader string

	licenseName := strings.Split(*fLicense, "-")[0]
	set := new(template.Set)

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

	for _, t := range []string{tmplCopyright, tmplHeader, tmplCmd, tmplPac,
		tmplTest} {
		fullSet, err = set.Parse(t)
		if err != nil {
			log.Fatal("parse error in %q: %s", t, err)
		}
	}

	// Tag to render the copyright in README.
	//	p.data["comment"] = ""
	//	p.data["copyright"] = parse(tmplCopyright, data)

	// These tags are not used anymore.
	//	for _, t := range []string{"Affero", "comment", "year"} {
	//		p.data[t] = "", false
	//	}

	return fullSet
}
