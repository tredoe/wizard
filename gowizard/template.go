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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"template"
	"time"
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


// === Header render
// ===

/* Renders the headers of source code files according to the license.
If `year` is nil then gets the actual year.
*/
func renderHeader(tag map[string]string, year string) map[string]string {
	const (
		COMMENT_CODE     = "//"
		COMMENT_MAKEFILE = "#"
	)

	var headerMakefile, headerCode string

	if year == "" {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
	}

	licenseName := strings.Split(*fLicense, "-", -1)[0]

	switch licenseName {
	case "apache":
		header := fmt.Sprint(tmplCopyright, tmplApache)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "bsd":
		header := fmt.Sprint(tmplCopyright, tmplBSD)

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "cc0":
		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(tmplCC0, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(tmplCC0, tag)
	case "gpl", "agpl":
		header := fmt.Sprint(tmplCopyright, tmplGNU)
		if licenseName == "agpl" {
			tag["Affero"] = "Affero"
		} else {
			tag["Affero"] = ""
		}

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	case "none":
		header := fmt.Sprint(tmplCopyright, "\n")

		tag["comment"] = COMMENT_MAKEFILE
		headerMakefile = parse(header, tag)

		tag["comment"] = COMMENT_CODE
		headerCode = parse(header, tag)
	}

	// Tag to render the copyright in README.
	tag["comment"] = ""
	tag["copyright"] = parse(tmplCopyright, tag)

	// These tags are not used anymore.
	for _, t := range []string{"Affero", "comment", "year"} {
		tag[t] = "", false
	}

	return map[string]string{
		"makefile": headerMakefile,
		"code":     headerCode,
	}
}

