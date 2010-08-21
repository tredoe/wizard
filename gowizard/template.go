// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

/* Template returns strings. */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"template"
)


// === Template and data to build source code files

// License headers
const (
	t_COPYRIGHT = `{.section comment}{@} {.end}Copyright {year}  The "{project_name}" Authors`

	t_BSD = `
{comment}
{comment} Use of this source code is governed by the {license}
{comment} that can be found in the LICENSE file.
{comment}
{comment} This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
{comment} OR CONDITIONS OF ANY KIND, either express or implied. See the License
{comment} for more details.
`

	t_APACHE = `
{comment}
{comment} Licensed under the Apache License, Version 2.0 (the "License");
{comment} you may not use this file except in compliance with the License.
{comment} You may obtain a copy of the License at
{comment}
{comment}     http://www.apache.org/licenses/LICENSE-2.0
{comment}
{comment} Unless required by applicable law or agreed to in writing, software
{comment} distributed under the License is distributed on an "AS IS" BASIS,
{comment} WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
{comment} See the License for the specific language governing permissions and
{comment} limitations under the License.
`

	t_GNU = `
{comment}
{comment} This program is free software: you can redistribute it and/or modify
{comment} it under the terms of the GNU {.section Affero}{@} {.end}General Public License as published by
{comment} the Free Software Foundation, either version 3 of the License, or
{comment} (at your option) any later version.
{comment}
{comment} This program is distributed in the hope that it will be useful,
{comment} but WITHOUT ANY WARRANTY; without even the implied warranty of
{comment} MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
{comment} GNU {.section Affero}{@} {.end}General Public License for more details.
{comment}
{comment} You should have received a copy of the GNU {.section Affero}{@} {.end}General Public License
{comment} along with this program.  If not, see <http://www.gnu.org/licenses/>.
`

	t_CC0 = `{comment} To the extent possible under law, Authors have waived all copyright and
{comment} related or neighboring rights to "{project_name}".
`
)

const t_PAGE = "{header}\n{content}"

type code struct {
	header  string
	content string
}


// === Template parser
// Based on http://go.hokapoka.com/go/embedding-or-nesting-go-templates/
// ===

type templateParser struct {
	str string
}

func (self *templateParser) Write(p []byte) (n int, err os.Error) {
	self.str += string(p)

	return len(p), nil
}


func parse(str string, data interface{}) string {
	_templateParser := new(templateParser)

	t := template.MustParse(str, nil)
	t.Execute(data, _templateParser)

	return _templateParser.str
}

func parseFile(filename string, data interface{}) string {
	_templateParser := new(templateParser)

	t := template.New(nil)
	t.SetDelims("{{", "}}")

	if err := t.ParseFile(filename); err != nil {
		log.Exit(err)
	}

	t.Execute(data, _templateParser)

	return _templateParser.str
}


// === Utility
// ===

/* Renders a source code file nesting both header and content. */
func renderCodeFile(header string, destination, template string, tag map[string]string) {
	renderContent := parseFile(template, tag)
	render := parse(t_PAGE, &code{header, renderContent})

	ioutil.WriteFile(
		path.Join(destination, path.Base(template)),
		[]byte(render),
		PERM_FILE,
	)
}

/* Base to rendering single files. */
func _renderFile(destination, template string, tag map[string]string) {
	render := parseFile(template, tag)
	ioutil.WriteFile(destination, []byte(render), PERM_FILE)
}

/* Renders a single file. */
func renderFile(destination, template string, tag map[string]string) {
	_renderFile(path.Join(destination, path.Base(template)), template, tag)
}

/* Renders a single file, but uses a new name. */
func renderNewFile(destination, template string, tag map[string]string) {
	_renderFile(destination, template, tag)
}

