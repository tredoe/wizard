// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Based on Metadata for Python Software Packages. */

package main

import (
	"flag"
	"regexp"
	"strings"
)


// Metadata flags
var (
	fProjectName = flag.String("Project-name", "",
		"The name of the project.")

	fPackageName = flag.String("Package-name", "",
		"The name of the package.")

	fVersion = flag.String("Version", "",
		"A string containing the package's version number.")

	fSummary = flag.String("Summary", "",
		"A one-line summary of what the package does.")

	fDownloadURL = flag.String("Download-URL", "",
		"A string containing the URL from which this version of the\n"+
			"\tpackage can be downloaded.")

	fAuthor = flag.String("Author", "",
		"A string containing the author's name at a minimum.")

	fAuthorEmail = flag.String("Author-email", "",
		"A string containing the author's e-mail address.")

	fPlatform = flag.String("Platform", "",
		"A comma-separated list of platform specifications, summarizing\n"+
			"\tthe operating systems supported by the package which are not listed\n"+
			"\tin the \"Operating System\" Trove classifiers.")

	fDescription = flag.String("Description", "",
		"A longer description of the package that can run to several\n"+
			"\tparagraphs.")

	fKeywords = flag.String("Keywords", "",
		"A list of additional keywords to be used to assist searching for\n"+
			"\tthe package in a larger catalog.")

	fHomePage = flag.String("Home-page", "",
		"A string containing the URL for the package's home page.")

	fClassifier = flag.String("Classifier", "",
		"Each entry is a string giving a single classification value\n"+
			"\tfor the package.")
)

/* v1.1 http://www.python.org/dev/peps/pep-0314/

The next fields have not been taken:

	Supported-Platform
	License
	Requires
	Provides
	Obsoletes

The field 'Name' has been substituted by 'Project-name' and 'Package-name'.

*/
type metadata_1_1 struct {
	MetadataVersion string "Metadata-Version" // Version of the file format
	ProjectName     string "Project-name"
	PackageName     string "Package-name"
	Version         string
	Summary         string
	DownloadURL     string "Download-URL"
	Author          string
	AuthorEmail     string "Author-email"

	// === Optional

	Platform    []string
	Description string
	Keywords    string
	HomePage    string "Home-page"
	Classifier  []string
}

func newMetadata_1_1(ProjectName, PackageName, Version, Summary, DownloadURL,
Author, AuthorEmail string) *metadata_1_1 {
	reGo := regexp.MustCompile(`^go`) // To remove it from the project name
	metadata := new(metadata_1_1)

	metadata.MetadataVersion = "1.1"
	metadata.Version = Version
	metadata.Summary = Summary
	metadata.DownloadURL = DownloadURL
	metadata.Author = Author
	metadata.AuthorEmail = AuthorEmail

	metadata.ProjectName = strings.TrimSpace(ProjectName)

	// The package name is created:
	// getting the last string after of the dash ('-'), if any,
	// and removing 'go'. Finally, it's lower cased.
	if PackageName == "" {
		packageName := strings.Split(metadata.ProjectName, "-", -1)
		metadata.PackageName = reGo.ReplaceAllString(
			strings.ToLower(packageName[len(packageName)-1]), "")
	} else {
		metadata.PackageName = strings.ToLower(
			strings.TrimSpace(metadata.PackageName))
	}

	return metadata
}

