// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

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
	t_LICENSE     = `// Copyright {year}, The "{project_name}" Authors.  All rights reserved.
// Use of this source code is governed by the {license} License
// that can be found in the LICENSE file.
`
	t_LICENSE_CC0 = `// To the extent possible under law, Authors have waived all copyright and
// related or neighboring rights to "{project_name}".
`
)

const t_PAGE = "{license}\n{content}"

type code struct {
	license string
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

/* Renders a source code file nesting both license and content. */
func renderCodeFile(license *string, destination, template string, tag map[string]string) {
	renderContent := parseFile(template, tag)
	render := parse(t_PAGE, &code{*license, renderContent})

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

