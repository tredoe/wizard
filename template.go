// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package wizard

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Copyright
const (
	tmplCopyright = `Copyright {{.Year}} {{.Author}}`
	tmplCopyleft  = `Written in {{.Year}} by {{.Author}}`

	tmplOrgCopyright = `Copyright {{.Year}} The {{.Project}} Authors`
	tmplOrgCopyleft  = `Written in {{.Year}} by the {{.Project}} Authors`
)

// Licenses
const (
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

	tmplMPL = `{{.Comment}} {{template "Copyright" .}}
{{.Comment}}
{{.Comment}} This Source Code Form is subject to the terms of the Mozilla Public
{{.Comment}} License, v. 2.0. If a copy of the MPL was not distributed with this
{{.Comment}} file, You can obtain one at http://mozilla.org/MPL/2.0/.
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
)

// Base of source files
const (
	tmplGo = `{{template "Header" .}}
package {{.Program}}

import (
	
)


`

	tmplTest = `{{template "Header" .}}
package {{.Program}}

import "testing"

func Test(t *testing.T) {
	
}
`

	tmplExample = `{{template "Header" .}}
package {{.Program}}_test

import (
	"fmt"

	"{{.ImportPath}}"
)

func Example() {
	fmt.Println()
	// Output:
	// 
}
`
)

// User configuration
const tmplUserConfig = `
org: {{.Org}}
author: {{.Author}}
email: {{.Email}}
license: {{.License}}
vcs: {{.VCS}}
import: {{.ImportPath}}
`

// Ignore file for VCS
const hgIgnoreTop = "syntax: glob\n"

var tmplIgnore = `## Special files
*~
[._]*

# Compiled Object files, Static and Dynamic libs (Shared Objects)
*.[ao]
*.so
*.dll

# Folders
_obj
_test

# Architecture specific extensions/prefixes
*.[568vq]
[568vq].out

*.cgo?.*

_testmain.go

# Compiled Go source
*.exe
*.test
*.prof

## Data
*.bin

## Packages
# It's better to unpack these files and commit the raw source since
# git has its own built in compression methods
*.7z
*.dmg
*.gz
*.iso
*.jar
*.rar
*.tar
*.zip

## Logs and databases
*.db
*.log
*.sql
*.sqlite

## OS generated files
Icon?

# * * *
{{.Program}}
`

// Information files
const (
	tmplAuthors = `
This is the official list of **{{.Project}}** authors for copyright purposes.  
This file is distinct from the 'CONTRIBUTORS' file. See the latter for an explanation.

Names should be added to this file as:

	Name or Organization <email address>

(The email address is not required for organizations)

Please keep the list sorted.
* * *

{{with .Org}}{{.}}{{else}}{{.Email}}{{end}}

`

	tmplContributors = `
This is the official list of people who can contribute (and typically have
contributed) code to the **{{.Project}}** repository.

The 'AUTHORS' file lists the copyright holders; this file lists people. For
example, the employees of an organization are listed here but not in 'AUTHORS',
because the organization holds the copyright.

Names should be added to this file as:

	Name <email address>

Please keep the list sorted.
* * *

{{.Email}}

`

	tmplChangelog = `
This file documents the changes in **{{.Project}}** versions that are listed below.

Items should be added to this file as:

	#### YYYY-MM-DD Release
	One change.
	Other change.

* * *


`

	tmplReadme = `{{.Project}}
{{.ProjectHeader}}
<< PROJECT SYNOPSIS >>

[Documentation online](http://godoc.org/{{with .ImportPath}}{{.}}{{else}}<< IMPORT PATH >>{{end}})

## Installation

	go get {{with .ImportPath}}{{.}}{{else}}<< IMPORT PATH >>{{end}}
{{if .FullLicense}}
## License

Unless otherwise noted:

+ The source files are distributed under the *{{.FullLicense}}*
{{end}}
* * *
*Generated by [Gowizard](https://github.com/tredoe/wizard)*
`
)

// * * *

// parseFromFile renders the template "src", creating a file in "dst".
func (p *project) parseFromFile(dst, src string) error {
	file, err := createFile(dst)
	if err != nil {
		return err
	}

	p.tmpl, err = p.tmpl.ParseFiles(src)
	if err != nil {
		return fmt.Errorf("parsing error: %s", err)
	}
	if err = p.tmpl.ExecuteTemplate(file, filepath.Base(src), p.cfg); err != nil {
		return fmt.Errorf("execution failed: %s", err)
	}

	return nil
}

// parseFromVar renders the template "tmplName" to the file "dst".
func (p *project) parseFromVar(dst string, tmplName string) error {
	file, err := createFile(dst)
	if err != nil {
		return err
	}

	if err = p.tmpl.ExecuteTemplate(file, tmplName, p.cfg); err != nil {
		return fmt.Errorf("execution failed: %s", err)
	}
	return nil
}

// parseLicense parses the license header.
// charComment is the character used to comment in code files.
func (p *project) parseLicense(charComment string) {
	licenseName := strings.Split(p.cfg.License, "-")[0]
	tmplHeader := ""

	p.cfg.Comment = charComment
	p.cfg.Year = time.Now().Year()

	switch licenseName {
	case "mpl":
		tmplHeader = tmplMPL
	case "apache":
		tmplHeader = tmplApache
	case "cc0":
		tmplHeader = tmplCC0
	case "gpl", "agpl":
		tmplHeader = tmplGNU

		if licenseName == "agpl" {
			p.cfg.GNUextra = "Affero"
		}
	case "none":
		tmplHeader = tmplNone
	}

	p.tmpl = template.Must(p.tmpl.New("Header").Parse(tmplHeader))

	if licenseName != "cc0" {
		if p.cfg.Org == "" {
			p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplCopyright))
		} else {
			p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplOrgCopyright))
		}
	} else {
		if p.cfg.Org == "" {
			p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplCopyleft))
		} else {
			p.tmpl = template.Must(p.tmpl.New("Copyright").Parse(tmplOrgCopyleft))
		}
	}
}

// parseProject parses the templates for the project.
func (p *project) parseProject() {
	p.tmpl = template.Must(p.tmpl.New("Authors").Parse(tmplAuthors))
	p.tmpl = template.Must(p.tmpl.New("Contributors").Parse(tmplContributors))
	p.tmpl = template.Must(p.tmpl.New("Changelog").Parse(tmplChangelog))
	p.tmpl = template.Must(p.tmpl.New("Readme").Parse(tmplReadme))
	p.tmpl = template.Must(p.tmpl.New("Go").Parse(tmplGo))
	p.tmpl = template.Must(p.tmpl.New("Test").Parse(tmplTest))
	p.tmpl = template.Must(p.tmpl.New("Example").Parse(tmplExample))

	// == Ignore file
	if p.cfg.VCS == "hg" {
		tmplIgnore = hgIgnoreTop + tmplIgnore
	}
	p.tmpl = template.Must(p.tmpl.New("Ignore").Parse(tmplIgnore))
}
