// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

/* Based on Metadata for Python Software Packages. */

package main

import ()


/* v1.1 http://www.python.org/dev/peps/pep-0314/

The next fields have not been taken:

	Supported-Platform
	License
	Requires
	Provides
	Obsoletes

And the field 'Name' has been substituted by 'Project-name' and 'Package-name'.

*/
type metadata_1_1 struct {
	// Version of the file format.
	MetadataVersion string "Metadata-Version" // 1.1

	// The name of the project.
	ProjectName string "Project-name"

	// The name of the package.
	PackageName string "Package-name"

	// A string containing the package's version number.
	Version string

	// A one-line summary of what the package does.
	Summary string

	// A string containing the URL from which this version of the package can be
	// downloaded.
	DownloadURL string "Download-URL"

	// A string containing the author's name at a minimum.
	Author string

	// A string containing the author's e-mail address.
	AuthorEmail string "Author-email"

	// === Optional
	// ===

	// A comma-separated list of platform specifications, summarizing the
	// operating systems supported by the package which are not listed in the
	// "Operating System" Trove classifiers.
	Platform []string

	// A longer description of the package that can run to several paragraphs.
	Description string

	// A list of additional keywords to be used to assist searching for the
	// package in a larger catalog.
	Keywords string

	// A string containing the URL for the package's home page.
	HomePage string "Home-page"

	// Each entry is a string giving a single classification value for the
	// package.
	Classifier []string
}

