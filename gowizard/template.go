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
	t_LICENSE     = `// Copyright {year}, The '{project}' Authors.  All rights reserved.
// Use of this source code is governed by the {license} License
// that can be found in the LICENSE file.
`
	t_LICENSE_CC0 = `// To the extent possible under law, Authors have waived all copyright and
// related or neighboring rights to '{project}'.
`
)

const t_PAGE = "{license}\n{content}"

type code struct {
	license string
	content string
}


// === Utility
// ===

/* Copy a file from the data directory to the project. */
func copy(destinationFile, sourceFile string) {
	src, err := ioutil.ReadFile(dataDir + sourceFile)
	if err != nil {
		log.Exit(err)
	}

	err = ioutil.WriteFile(*fProjectName+destinationFile, src, _PERM_FILE)
	if err != nil {
		log.Exit(err)
	}
}

/* Creates a source code file nesting both license and content. */
func renderCodeFile(license *string, contentTemplate string, tag map[string]string) {
	contentRender := parseFile(dataDir+contentTemplate, tag)
	render := parse(t_PAGE, &code{*license, contentRender})

	ioutil.WriteFile(
		path.Join(*fProjectName, *fPackageName, path.Base(contentTemplate)),
		[]byte(render),
		_PERM_FILE,
	)
}

/* Creates single files. */
func renderFile(contentTemplate string, tag map[string]string) {
	render := parseFile(dataDir+contentTemplate, tag)

	ioutil.WriteFile(
		path.Join(*fProjectName, path.Base(contentTemplate)),
		[]byte(render),
		_PERM_FILE,
	)
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

	// === Gets the content of filename
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic("gowizard.parseFile error: " + err.String())
	}

	t := template.New(nil)
	t.SetDelims("{{", "}}")

	if err := t.Parse(string(b)); err != nil {
		panic(err)
	}

	t.Execute(data, _templateParser)

	return _templateParser.str
}

