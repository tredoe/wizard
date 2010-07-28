// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Template returns strings.

Based on http://go.hokapoka.com/go/embedding-or-nesting-go-templates/
*/

package main

import (
	"io/ioutil"
	"os"
	"template"
)


// === Template and data to build source code files

// License headers
const (
	t_HEADER     = `// Copyright {year}, The '{project}' Authors.  All rights reserved.
// Use of this source code is governed by the {license} License
// that can be found in the LICENSE file.
`
	t_HEADER_CC0 = `// To the extent possible under law, Authors have waived all copyright and
// related or neighboring rights to '{project}'.
`
)

const t_PAGE = "{header}\n{content}"

type page struct {
	header  string
	content string
}


// === Template parser
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

