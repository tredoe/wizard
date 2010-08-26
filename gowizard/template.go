// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"template"
)


// === Structure of a page for a source code file
const tmplCode = "{{header}}\n{{content}}"

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

	t := template.New(nil)
	t.SetDelims("{{", "}}")

	if err := t.Parse(str); err != nil {
		log.Exit(err)
	}

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

/* Renders template nesting both header and content. */
func renderCode(destination, template, header string, tag map[string]string) {
	renderContent := parse(template, tag)
	render := parse(tmplCode, &code{header, renderContent})

	ioutil.WriteFile(destination, []byte(render), PERM_FILE)
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

