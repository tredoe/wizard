gowizard
========

Tired of adding the same files every time you create a new Go project?  
Don't know how to structure it?  
Don't know how to apply the license?

**gowizard** creates the base for new Go projects, adds the license header
to source code files, and creates an ignore file for the VCS given.

The ignore file has been configured to ignore files finished in "~" (used like
backups, in Unix), the ones started with "." (hidden files, in Unix), and the
ones started with "_" to have files that don't be committed. It also ignores
files got from compiling and linking.


## Installation

	go get github.com/kless/gowizard/gowizard

To only install the library, which could be used by a Go IDE:

	go get github.com/kless/gowizard


## Configuration

To don't repeat the same every time you create a project, you could use an user
configuration file in your home directory to have values by default.

	gowizard -i -cfg


## Operating instructions

#### Create project

By default, the program name (flag *-program*) is named as the project name but
in lower case, and removing the name "Go" of the prefix and suffix.

The way fastest and simple to create it, is using the interactive mode:

	gowizard -i

#### Add program

To add a program to the actual project:

	cd <project name>
	gowizard -i -add

#### Suggestion about naming

> Frequently, the name that you use for your package will include the name "Go"
as a prefix, suffix, or part of its acronym, and you may or may not want this
to be a part of the actual command or package name in a go source file.

http://code.google.com/p/go-wiki/wiki/PackagePublishing#Subdirectories

#### Suggestion about licenses

> The Apache License 2.0 is the best non-copyleft license that does what a
copyright license can to mitigate threats from software patents. It's a
well-established, mature license that users, developers, and distributors alike
are all comfortable with. You can tell it's important by the way that other free
software licenses work to cooperate with it: the drafting processes for GPLv3
and the Mozilla Public License 2.0 named compatibility with the Apache License
2.0 as a goal from day one. The Apache Software Foundation deserves a lot of
credit for pushing to do more to tackle software patents in a license, and
implementing an effective strategy in the Apache License.

http://www.fsf.org/blogs/licensing/new-license-recommendations-guide  
http://www.gnu.org/licenses/license-recommendations.html  
http://www.freebsd.org/doc/en/articles/bsdl-gpl/

#### Maintenance

Copyright notices only need the year the file was created, so don't add new
years.


## Copyright and licensing

*Copyright 2010  The "gowizard" Authors*. See file AUTHORS and CONTRIBUTORS.  
Unless otherwise noted, the source files are distributed under the
*Apache License, version 2.0* found in the LICENSE file.

