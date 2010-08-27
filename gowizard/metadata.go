// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

/* Based on Metadata for Python Software Packages.

Description of fields that are not set via 'flag':

* Version: A string containing the package's version number.

* Summary: A one-line summary of what the package does.

* Download-URL: A string containing the URL from which this version of the
	package can be downloaded.

* Platform: A comma-separated list of platform specifications, summarizing
	the operating systems supported by the package which are not listed
	in the "Operating System" Trove classifiers.

* Description: A longer description of the package that can run to several
	paragraphs.

* Keywords: A list of additional keywords to be used to assist searching for
	the package in a larger catalog.

* Home-page: A string containing the URL for the package's home page.

* Classifier: Each entry is a string giving a single classification value
	for the package.

*/

package main

import (
	"os"
	"path"
	"reflect"

	"github.com/kless/goconfig/config"
)


const _FILE_NAME = "Metadata"

// Available project types
var listProject = map[string]string{
	"tool": "Development tool",
	"app":  "Program",
	"cgo":  "Package that calls C code",
	"lib":  "Library",
}

// Available licenses
var listLicense = map[string]string{
	"apache-2": "Apache License, version 2.0",
	"bsd-2":    "Simplified BSD License",
	"bsd-3":    "New BSD License",
	"cc0-1":    "Creative Commons CC0, version 1.0 Universal",
	"gpl-3":    "GNU General Public License, version 3 or later",
	"agpl-3":   "GNU Affero General Public License, version 3 or later",
	"none":     "Proprietary License",
}


// === Errors
type MetadataFieldError string

func (self MetadataFieldError) String() string {
	return "metadata: section default has not field '" + string(self) + "'"
}


/* v1.1 http://www.python.org/dev/peps/pep-0314/

The next fields have not been taken:

	Supported-Platform
	Requires
	Provides
	Obsoletes

Neither the next ones because they are only useful on Python since they are
added to pages on packages index:

	Description
	Classifier

The field 'Name' has been substituted by 'Project-name' and 'Package-name'.
The field 'License' needs a value from the map 'license'.

It has been added 'Project-type'.

For 'Classifier' see on http://pypi.python.org/pypi?%3Aaction=list_classifiers
*/
type metadata struct {
	MetadataVersion string "Metadata-Version" // Version of the file format
	ProjectType     string "Project-type"
	ProjectName     string "Project-name"
	PackageName     string "Package-name"
	Version         string
	Summary         string
	DownloadURL     string "Download-URL"
	Author          string
	AuthorEmail     string "Author-email"
	License         string

	// === Optional
	Platform string
	//Description string
	Keywords string
	HomePage string "Home-page"
	//Classifier  []string

	// Config file
	file *config.File
}

/* Creates a new metadata with the basic fields to build the project. */
func NewMetadata(ProjectType, ProjectName, PackageName, License, Author,
AuthorEmail string) *metadata {
	metadata := new(metadata)
	metadata.file = config.NewFile()

	metadata.MetadataVersion = "1.1"
	metadata.ProjectType = ProjectType
	metadata.ProjectName = ProjectName
	metadata.PackageName = PackageName
	metadata.License = License
	metadata.Author = Author
	metadata.AuthorEmail = AuthorEmail

	return metadata
}

/* Reads metadata file. */
func ReadMetadata() (*metadata, os.Error) {
	file, err := config.ReadFile(_FILE_NAME)
	if err != nil {
		return nil, err
	}

	metadata := new(metadata)
	metadata.file = file

	// === Section 'default' has several required fields.
	if s, err := file.String("default", "project-type"); err == nil {
		metadata.ProjectType = s
	} else {
		return nil, MetadataFieldError("project-type")
	}
	if s, err := file.String("default", "project-name"); err == nil {
		metadata.ProjectName = s
	} else {
		return nil, MetadataFieldError("project-name")
	}
	if s, err := file.String("default", "package-name"); err == nil {
		metadata.PackageName = s
	} else {
		return nil, MetadataFieldError("package-name")
	}
	if s, err := file.String("default", "license"); err == nil {
		metadata.License = s
	} else {
		return nil, MetadataFieldError("license")
	}

	if s, err := file.String("base", "author"); err == nil {
		metadata.Author = s
	}
	if s, err := file.String("base", "author-email"); err == nil {
		metadata.AuthorEmail = s
	}
	if s, err := file.String("base", "version"); err == nil {
		metadata.Version = s
	}
	if s, err := file.String("base", "summary"); err == nil {
		metadata.Summary = s
	}
	if s, err := file.String("base", "download-url"); err == nil {
		metadata.DownloadURL = s
	}

	if s, err := file.String("optional", "platform"); err == nil {
		metadata.Platform = s
	}
	if s, err := file.String("optional", "keywords"); err == nil {
		metadata.Keywords = s
	}
	if s, err := file.String("optional", "home-page"); err == nil {
		metadata.HomePage = s
	}

	return metadata, nil
}

/* Serializes to INI format and write it to file `_FILE_NAME` in directory `dir`.
 */
func (self *metadata) WriteINI(dir string) os.Error {
	header := "Generated by gowizard"
	reflectMetadata := self.getStruct()

	default_ := []string{
		"MetadataVersion",
		"ProjectType",
		"ProjectName",
		"PackageName",
		"License",
	}

	base := []string{
		"Version",
		"Summary",
		"DownloadURL",
		"Author",
		"AuthorEmail",
	}

	optional := []string{
		"Platform",
		//"Description",
		"Keywords",
		"HomePage",
		//"Classifier",
	}

	for i := 0; i < len(default_); i++ {
		name, value := reflectMetadata.name_value(default_[i])
		self.file.AddOption("", name, value)
	}

	for i := 0; i < len(base); i++ {
		name, value := reflectMetadata.name_value(base[i])
		self.file.AddOption("base", name, value)
	}

	for i := 0; i < len(optional); i++ {
		name, value := reflectMetadata.name_value(optional[i])
		self.file.AddOption("optional", name, value)
	}

	filePath := path.Join(dir, _FILE_NAME)
	if err := self.file.WriteFile(filePath, PERM_FILE, header); err != nil {
		return err
	}

	return nil
}


// === Reflection
// ===

// To handle the reflection of a struct
type reflectStruct struct {
	strType  *reflect.StructType
	strValue *reflect.StructValue
}

/* Gets structs that represent the type 'metadata'. */
func (self *metadata) getStruct() *reflectStruct {
	v := reflect.NewValue(self).(*reflect.PtrValue)

	strType := v.Elem().Type().(*reflect.StructType)
	strValue := v.Elem().(*reflect.StructValue)

	return &reflectStruct{strType, strValue}
}

/* Gets tag or field name and its value, given the field name. */
func (self *reflectStruct) name_value(fieldName string) (name, value string) {
	field, _ := self.strType.FieldByName(fieldName)
	value_ := self.strValue.FieldByName(fieldName)

	value = value_.(*reflect.StringValue).Get()

	if tag := field.Tag; tag != "" {
		name = tag
	} else {
		name = field.Name
	}

	return
}

