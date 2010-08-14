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

// License licenses
const (
	t_LICENSE     = `// Copyright {year}, The '{project_name}' Authors.  All rights reserved.
// Use of this source code is governed by the {license} License
// that can be found in the LICENSE file.
`
	t_LICENSE_CC0 = `// To the extent possible under law, Authors have waived all copyright and
// related or neighboring rights to '{project_name}'.
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

/* Creates a source code file nesting both license and content. */
func renderCodeFile(license *string, contentTemplate string, tag map[string]string) {
	contentRender := parseFile(contentTemplate, tag)
	render := parse(t_PAGE, &code{*license, contentRender})

	ioutil.WriteFile(
		path.Join(cfg.ProjectName, cfg.ApplicationName, path.Base(contentTemplate)),
		[]byte(render),
		PERM_FILE,
	)
}

/* Creates single files. */
func renderFile(contentTemplate string, tag map[string]string) {
	render := parseFile(contentTemplate, tag)

	ioutil.WriteFile(
		path.Join(cfg.ProjectName, path.Base(contentTemplate)),
		[]byte(render),
		PERM_FILE,
	)
}

