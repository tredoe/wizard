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
*/
package main
