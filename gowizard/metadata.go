// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Based on Metadata for Python Software Packages. */

package main

import (
	"io/ioutil"
	"json"
	"log"
	"path"
)


const _FILE_NAME = "Metadata"


/* v1.1 http://www.python.org/dev/peps/pep-0314/

The next fields have not been taken:

	Supported-Platform
	Requires
	Provides
	Obsoletes

The field 'Name' has been substituted by 'Project-name' and 'Package-name'.
The field 'License' needs a value from the map 'license'.

For 'Classifier' see on http://pypi.python.org/pypi?%3Aaction=list_classifiers
*/
type metadata struct {
	MetadataVersion string "Metadata-Version" // Version of the file format
	ProjectName     string "Project-name"
	PackageName     string "Package-name"
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
}

func NewMetadata(ProjectName, PackageName, Version, Summary, DownloadURL,
Author, AuthorEmail, License, Platform, Description, Keywords, HomePage string,
Classifier []string) *metadata {
	metadata := new(metadata)

	metadata.MetadataVersion = "1.1"
	metadata.ProjectName = ProjectName
	metadata.PackageName = PackageName
	metadata.Version = Version
	metadata.Summary = Summary
	metadata.DownloadURL = DownloadURL
	metadata.Author = Author
	metadata.AuthorEmail = AuthorEmail
	metadata.License = License

	metadata.Platform = Platform
	metadata.Description = Description
	metadata.Keywords = Keywords
	metadata.HomePage = HomePage
	metadata.Classifier = Classifier

	return metadata
}

/* Serializes to its equivalent JSON representation and write it to file
`_FILE_NAME` in directory `dir`.
*/
func (self *metadata) writeJSON(dir string) {
	filePath := path.Join(dir, _FILE_NAME)

	bytesOutput, err := json.MarshalIndent(self, " ", "   ")
	if err != nil {
		log.Exit(err)
	}

	if err := ioutil.WriteFile(filePath, bytesOutput, PERM_FILE); err != nil {
		log.Exit(err)
	}
}


func ReadMetadata() {

}

