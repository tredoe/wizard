/*
Command gowizard creates the base for new Go projects, adds the license header
to source code files, and creates a file ignore for the VCS given.

The file ignore has been configured to ignore files finished in "~" (used like
backups in Unix), the ones started with "." (hidden files in Unix), and the
ones started with "_" to have files that don't be committed. It also ignores
files got from compiling and linking.


Configuration

To don't repeat the same every time you create a project, you could use an user
configuration file in your home directory to have values by default.

	gowizard -i -cfg

Create project

By default, the program name (flag *-program*) is named as the project name but
in lower case, and removing the name "Go" of the prefix and suffix.

The way fastest and simple to create it, is using the interactive mode:

	gowizard -i

Add program

To add a program to the actual project:

	cd <project name>
	gowizard -i -add


Licenses

The BSD-like licenses have been excluded because they can not mitigate threats
from software patents and LGPL because it has not benefits using into a language
of static linking.

My suggestion is to use MPL 2.0 because it allows covered source code to be mixed
with other files under a different, even proprietary license. However, code
files licensed under the MPL must remain under the MPL and freely available in
source form.

GPL/AGPL 3.0

Proprietary software linking: Not allowed.
Distribution with code under another license: Not allowed with software whose license is not GNU GPL compatible.
Redistributing of the code with changes: Only under GNU GPL/AGPL.

Apache Public 2.0

Proprietary software linking: Allowed.
Distribution with code under another license: Allowed.
Redistributing of the code with changes: Allowed.
Compatible with GNU GPL: Yes.

MPL 2.0

Proprietary software linking: Allowed.
Distribution with code under another license: Allowed.
Redistributing of the code with changes: Only under MPL.
Compatible with GNU GPL: Yes


Maintenance

Copyright notices only need the year the file was created, so don't add new years.
*/
package main
