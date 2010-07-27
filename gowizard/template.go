// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Template returns strings.

Based on http://go.hokapoka.com/go/embedding-or-nesting-go-templates/
*/

package main

import (
	"os"
	"template"
)


type parserTemplate struct {
	str string
}

func (self *parserTemplate) Write(p []byte) (n int, err os.Error) {
	self.str += string(p)

	return len(p), nil
}


func parse(str string, data interface{}) string {
	_parserTemplate := new(parserTemplate)

	t := template.MustParse(str, nil)
	t.Execute(data, _parserTemplate)

	return _parserTemplate.str
}

func parseFile(filename string, data interface{}) string {
	_parserTemplate := new(parserTemplate)

	t := template.MustParseFile(filename, nil)
	t.Execute(data, _parserTemplate)

	return _parserTemplate.str
}

