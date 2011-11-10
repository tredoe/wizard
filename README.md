GoWizard
========

Tired of adding the same files every time you create a new Go project?  
Don't know how to structure it?  
Don't know how to apply the license?

**GoWizard** creates the base for new Go projects, adds the license header
to source code files, and creates an ignore file for the VCS given.

The ignore file has been configured to ignore files finished in "~" (used like
backups, in Unix), the ones started with "." (hidden files, in Unix), and the
ones started with "_" to have files that don't be committed. It also ignores
files got from compiling and linking.


## Installation

	goinstall github.com/kless/GoWizard/gowizard

To only install the library, which could be used by a Go IDE:

	goinstall github.com/kless/GoWizard/wizard


## Configuration

If you want not repeat the same input every time you create a project, then you
could use an user configuration file in your home directory. It allows set the
flags *-author*, *-email*, *-license*, and *-vcs*, with a value by default.

To create it, you have to use the flag *-config*, besides of the required flags
at creating a project.


## Operating instructions

### Create

The required flags to build a new project are *-project-type*, *-project-name*,
*-license* ,*-author*, *-email* and *-vcs*.

By default, the package name (*-package-name*) is named as the project name but
in lower case, and removing the name "Go" of the prefix and suffix.

The flag *-vcs* indicates the version control system to use, which allows to add
the appropriate ignore file.

The way fastest and simple to use it, is create a configuration file and then
use the interactive mode:

	gowizard -config -author="John Foo" -email="e@mail.com" -license=bsd-2 -vcs=git
	gowizard -i

### Suggestions about naming

> Frequently, the name that you use for your package will include the name "Go"
as a prefix, suffix, or part of its acronym, and you may or may not want this
to be a part of the actual command or package name in a go source file.

http://code.google.com/p/go-wiki/wiki/PackagePublishing#Subdirectories

### Maintenance

Copyright notices only need the year the file was created, so don't add new
years.


## Licensing

***Copyright 2010  The "GoWizard" Authors***  
See file AUTHORS and CONTRIBUTORS.

Licensed under **BSD-2 Clause license**.  
See file LICENSE.


* * *
*Generated by* **myself**

