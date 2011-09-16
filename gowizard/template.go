// Copyright 2010  The "Go-Wizard" Authors
//
// Use of this source code is governed by the BSD-2 Clause license
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"template"
	"time"
)

// For comments in source code files
const CHAR_CODE_COMMENT = "//"

// Copyright and licenses
const (
	tmplCopyright = `{{define "Copyright"}}{{with .comment}}{{.}} {{end}}Copyright {{.year}}  The "{{.project_name}}" Authors{{end}}`

	tmplBSD = `{{define "BSD"}}{{template "Copyright" .}}
{{.comment}}
{{.comment}} Use of this source code is governed by the {{.license}}
{{.comment}} that can be found in the LICENSE file.
{{.comment}}
{{.comment}} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{{.comment}} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{{.comment}} for more details.
{{end}}`

	tmplApache = `{{define "Apache"}}{{template "Copyright" .}}
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

	tmplGNU = `{{define "GNU"}}{{template "Copyright" .}}
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

	tmplNone = `{{define "None"}}{{template "Copyright" .}}
{{end}}`

	tmplCC0 = `{{define "CC0"}}{{.comment}} To the extent possible under law, Authors have waived all copyright and
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

	tmplPac = `{{define "Pac"}}{{.Header}}
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
func renderFile(dst, src string, data interface{}) {
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
		reportExit(err)
	}
	if err = tmpl.Execute(file, data); err != nil {
		reportExit(err)
	}
}

// Renders the template "tmplName" in "set" to the file "dst".
func renderSet(dst string, set *template.Set, tmplName string, data interface{}) {
	// === Create file.
	file, err := os.Create(dst)
	if err != nil {
		log.Fatal("file error:", err)
	}
	if err = file.Chmod(PERM_FILE); err != nil {
		log.Fatal("file error:", err)
	}

	err = set.Execute(file, tmplName, data)
	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
}


// Parses the templates.
// "charComment" is the character used to comment in code files.
// If "year" is nil then gets the actual year.
func parseTemplates(data map[string]interface{}, charComment string, year int) *template.Set {
	var tmplHeader string
	licenseName := strings.Split(*fLicense, "-")[0]

	data["comment"] = charComment

	if year == 0 {
		data["year"] = strconv.Itoa64(time.LocalTime().Year)
	} else {
		data["year"] = year
	}

	switch licenseName {
	case "apache":
		tmplHeader = tmplApache
	case "bsd":
		tmplHeader = tmplBSD
	case "cc0":
		tmplHeader = tmplCC0
	case "gpl", "lgpl", "agpl":
		data["Affero"] = ""
		data["Lesser"] = ""

		if licenseName == "agpl" {
			data["Affero"] = "Affero"
		} else if licenseName == "lgpl" {
			data["Lesser"] = "Lesser"
		}

		tmplHeader = tmplGNU
	case "none":
		tmplHeader = tmplNone
	}

	set := new(template.Set)
	var fullSet *template.Set
	var err os.Error

//	for _, t := range []string{tmplApache, tmplBSD, tmplCC0, tmplGNU, tmplNone,
	for _, t := range []string{tmplCopyright, tmplHeader, tmplCmd, tmplPac,
			tmplTest} {
		fullSet, err = set.Parse(t)
		if err != nil {
			log.Fatal("parse error in %q: %s", t, err)
		}
	}


	// Tag to render the copyright in README.
//	data["comment"] = ""
//	data["copyright"] = parse(tmplCopyright, data)

	// These tags are not used anymore.
//	for _, t := range []string{"Affero", "comment", "year"} {
//		data[t] = "", false
//	}

	return fullSet
}

