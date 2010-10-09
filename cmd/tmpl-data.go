// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

/* Templates data. */

package main


// Copyright and licenses
const (
	tmplCopyright = `{{.section comment}}{{@}} {{.end}}Copyright {{year}}  The "{{project_name}}" Authors`

	tmplBSD = `
{{comment}}
{{comment}} Use of this source code is governed by the {{license}}
{{comment}} that can be found in the LICENSE file.
{{comment}}
{{comment}} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{{comment}} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{{comment}} for more details.
`

	tmplApache = `
{{comment}}
{{comment}} Licensed under the Apache License, Version 2.0 (the "License");
{{comment}} you may not use this file except in compliance with the License.
{{comment}} You may obtain a copy of the License at
{{comment}}
{{comment}}     http://www.apache.org/licenses/LICENSE-2.0
{{comment}}
{{comment}} Unless required by applicable law or agreed to in writing, software
{{comment}} distributed under the License is distributed on an "AS IS" BASIS,
{{comment}} WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
{{comment}} See the License for the specific language governing permissions and
{{comment}} limitations under the License.
`

	tmplGNU = `
{{comment}}
{{comment}} This program is free software: you can redistribute it and/or modify
{{comment}} it under the terms of the GNU {{.section Affero}}{{@}} {{.end}}{{.section Lesser}}{{@}} {{.end}}General Public License as published by
{{comment}} the Free Software Foundation, either version 3 of the License, or
{{comment}} (at your option) any later version.
{{comment}}
{{comment}} This program is distributed in the hope that it will be useful,
{{comment}} but WITHOUT ANY WARRANTY; without even the implied warranty of
{{comment}} MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
{{comment}} GNU {{.section Affero}}{{@}} {{.end}}{{.section Lesser}}{{@}} {{.end}}General Public License for more details.
{{comment}}
{{comment}} You should have received a copy of the GNU {{.section Affero}}{{@}} {{.end}}{{.section Lesser}}{{@}} {{.end}}General Public License
{{comment}} along with this program.  If not, see <http://www.gnu.org/licenses/>.
`

	tmplCC0 = `{{comment}} To the extent possible under law, Authors have waived all copyright and
{{comment}} related or neighboring rights to "{{project_name}}".
`
)

// For source code files
const (
	tmplTest = `package {{package_name}}

import (
	"testing"
)


func Test(t *testing.T) {

}

`

	tmplCmdMain = `package main

import (

)


func main() {

}

`

	tmplPkgMain = `package {{package_name}}
{{.section project_is_cgo}}

import "C"
{{.end}}

import (

)


func () {

}

`

	tmplCmdMakefile = `include $(GOROOT)/src/Make.inc

TARG={{package_name}}
GOFILES=\
	{{package_name}}.go\

include $(GOROOT)/src/Make.cmd

`

	tmplPkgMakefile = `include $(GOROOT)/src/Make.inc

TARG={{package_name}}
GOFILES=\
	main.go\

include $(GOROOT)/src/Make.pkg

`
)

// === Ignore file for VCS
var (
	tmplIgnore = `*~
_*

# Go
*.o
*.a
*.[568vq]
[568vq].out
main
`
)

const hgIgnoreTop = "syntax: glob\n"

