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
	"os"
	"path"
	"strconv"
	"strings"
	"template"
	"time"
)


// === Structure of a page for a source code file
const tmplCode = "{{tmplHeader}}\n{{content}}"

type code struct {
	tmplHeader string
	content    string
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
		reportExit(err)
	}

	t.Execute(data, _templateParser)

	return _templateParser.str
}

func parseFile(filename string, data interface{}) string {
	_templateParser := new(templateParser)

	t := template.New(nil)
	t.SetDelims("{{", "}}")

	if err := t.ParseFile(filename); err != nil {
		reportExit(err)
	}

	t.Execute(data, _templateParser)

	return _templateParser.str
}


// === Utility
// ===

// Renders template nesting both tmplHeader and content.
func renderNesting(destination, tmplHeader, template string,
tag map[string]string) {
	renderContent := parse(template, tag)
	render := parse(tmplCode, &code{tmplHeader, renderContent})

	ioutil.WriteFile(destination, []byte(render), PERM_FILE)
}


// Base to rendering single files.
func _renderFile(destination, template string, tag map[string]string) {
	render := parseFile(template, tag)
	ioutil.WriteFile(destination, []byte(render), PERM_FILE)
}

// Renders a single file.
func renderFile(destination, template string, tag map[string]string) {
	_renderFile(path.Join(destination, path.Base(template)), template, tag)
}

// Renders a single file, but uses a new name.
func renderNewFile(destination, template string, tag map[string]string) {
	_renderFile(destination, template, tag)
}


// === Render of header
// ===

// Base to render the headers of source code files according to the license.
// If `year` is nil then gets the actual year.
func _renderHeader(tag map[string]string, year string, renderCodeFile,
renderMakefile bool) (headerCodeFile, headerMakefile string) {
	licenseName := strings.Split(*fLicense, "-", -1)[0]

	if year == "" {
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
	}

	switch licenseName {
	case "apache":
		tmplHeader := tmplCopyright + tmplApache

		if renderCodeFile {
			tag["comment"] = CHAR_CODE_COMMENT
			headerCodeFile = parse(tmplHeader, tag)
		}
		if renderMakefile {
			tag["comment"] = CHAR_MAKE_COMMENT
			headerMakefile = parse(tmplHeader, tag)
		}
	case "bsd":
		tmplHeader := tmplCopyright + tmplBSD

		if renderCodeFile {
			tag["comment"] = CHAR_CODE_COMMENT
			headerCodeFile = parse(tmplHeader, tag)
		}
		if renderMakefile {
			tag["comment"] = CHAR_MAKE_COMMENT
			headerMakefile = parse(tmplHeader, tag)
		}
	case "cc0":
		if renderCodeFile {
			tag["comment"] = CHAR_CODE_COMMENT
			headerCodeFile = parse(tmplCC0, tag)
		}
		if renderMakefile {
			tag["comment"] = CHAR_MAKE_COMMENT
			headerMakefile = parse(tmplCC0, tag)
		}
	case "gpl", "lgpl", "agpl":
		tmplHeader := tmplCopyright + tmplGNU

		tag["Affero"] = ""
		tag["Lesser"] = ""

		if licenseName == "agpl" {
			tag["Affero"] = "Affero"
		} else if licenseName == "lgpl" {
			tag["Lesser"] = "Lesser"
		}

		if renderCodeFile {
			tag["comment"] = CHAR_CODE_COMMENT
			headerCodeFile = parse(tmplHeader, tag)
		}
		if renderMakefile {
			tag["comment"] = CHAR_MAKE_COMMENT
			headerMakefile = parse(tmplHeader, tag)
		}
	case "none":
		tmplHeader := tmplCopyright + "\n"

		if renderCodeFile {
			tag["comment"] = CHAR_CODE_COMMENT
			headerCodeFile = parse(tmplHeader, tag)
		}
		if renderMakefile {
			tag["comment"] = CHAR_MAKE_COMMENT
			headerMakefile = parse(tmplHeader, tag)
		}
	}

	// Tag to render the copyright in README.
	tag["comment"] = ""
	tag["copyright"] = parse(tmplCopyright, tag)

	// These tags are not used anymore.
	for _, t := range []string{"Affero", "comment", "year"} {
		tag[t] = "", false
	}

	if renderCodeFile && renderMakefile {
		return headerCodeFile, headerMakefile
	}

	if renderCodeFile {
		return headerCodeFile, ""
	}

	// if renderMakefile
	return headerMakefile, ""
}

func renderCodeHeader(tag map[string]string, year string) (
headerCodeFile, headerMakefile string) {
	return _renderHeader(tag, year, true, false)
}

func renderMakeHeader(tag map[string]string, year string) (
headerMakefile, headerCodeFile string) {
	return _renderHeader(tag, year, false, true)
}

func renderAllHeaders(tag map[string]string, year string) (
headerCodeFile, headerMakefile string) {
	return _renderHeader(tag, year, true, true)
}

