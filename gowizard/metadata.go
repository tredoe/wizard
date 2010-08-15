// Copyright 2010, The "gowizard" Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Based on Metadata for Python Software Packages. */

package main

import (
	"log"
	"path"
	"reflect"

	conf "goconf.googlecode.com/hg"
)


const _FILE_NAME = "Metadata"

// Available application types
var listApp = map[string]string{
	"cmd":    "command line",
	"pkg":    "package",
	"web.go": "web environment",
}

// Available licenses
var listLicense = map[string]string{
	"apache-2": "Apache License (version 2.0)",
	"bsd-2":    "Simplified BSD License",
	"bsd-3":    "New BSD License",
	"cc0-1":    "Creative Commons CC0 1.0 Universal",
	"gpl-3":    "GNU General Public License",
	"agpl-3":   "GNU Affero General Public License",
	"none":     "Proprietary License",
}


/* v1.1 http://www.python.org/dev/peps/pep-0314/

The next fields have not been taken:

	Supported-Platform
	Requires
	Provides
	Obsoletes

The field 'Name' has been substituted by 'Project-name' and 'Application-name'.
The field 'License' needs a value from the map 'license'.

It has been added 'Application-type'.

For 'Classifier' see on http://pypi.python.org/pypi?%3Aaction=list_classifiers
*/
type metadata struct {
	MetadataVersion string "Metadata-Version" // Version of the file format
	ProjectName     string "Project-name"
	ApplicationName string "Application-name"
	ApplicationType string "Application-type"
	Version         string
	Summary         string
	DownloadURL     string "Download-URL"
	Author          string
	AuthorEmail     string "Author-email"
	License         string

	// === Optional
	Platform    string
	Description string
	Keywords    string
	HomePage    string "Home-page"
	Classifier  []string

	// Config file
	file *conf.ConfigFile
}

func NewMetadata(ProjectName, ApplicationName, ApplicationType, Author,
AuthorEmail, License string, file *conf.ConfigFile) *metadata {
	metadata := new(metadata)

	metadata.MetadataVersion = "1.1"
	metadata.ProjectName = ProjectName
	metadata.ApplicationName = ApplicationName
	metadata.ApplicationType = ApplicationType
	metadata.Author = Author
	metadata.AuthorEmail = AuthorEmail
	metadata.License = License

	metadata.file = file

	return metadata
}

/* Serializes to INI format and write it to file `_FILE_NAME` in directory `dir`.
*/
func (self *metadata) WriteINI(dir string) {
	name := getTag(self)

	value := []string{
		self.MetadataVersion,
		self.ProjectName,
		self.ApplicationName,
		self.ApplicationType,
		self.Version,
		self.Summary,
		self.DownloadURL,
		self.Author,
		self.AuthorEmail,
		self.License,
	}

	optional := []string{
		self.Platform,
		self.Description,
		self.Keywords,
		self.HomePage,
		//self.Classifier,
	}

	for i := 0; i < len(value); i++ {
		self.file.AddOption(conf.DefaultSection, name[i], value[i])
	}

	for i := 0; i < len(optional); i++ {
		self.file.AddOption("optional", name[i+len(value)], optional[i])
	}

	filePath := path.Join(dir, _FILE_NAME)
	if err := self.file.WriteConfigFile(filePath, PERM_FILE, "Created by gowizard"); err != nil {
		log.Exit(err)
	}
}

func ReadMetadata() {

}


// === Utility
// ===

/* Gets the tags of a struct, if any. */
func getTag(i interface{}) (name []string) {
	switch v := reflect.NewValue(i).(type) {
	case *reflect.PtrValue:
		t := v.Elem().Type().(*reflect.StructType)
		num := t.NumField()
		name = make([]string, num)

		for i := 0; i < num; i++ {
			field := t.Field(i)

			if f := field.Tag; f != "" {
				name[i] = f
			} else {
				name[i] = field.Name
			}
		}
	}

	return name
}

