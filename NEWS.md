###### Notice

*This file documents the changes in ***GoWizard*** versions that are
listed below.*

*Items should be added to this file as:*

	### YYYY-MM-DD  Release

	+ Additional changes.

	+ More changes.

* * *

### 2011-??-??  v1.0

+ Use a struct to handle the configuration instead of use global variables.
+ The command-line flags have been changed to lower case.
+ Updated Go-Inline's API.


### 2011-09-18  v0.9.9

+ Removed all stuff related to the file Metadata.

+ Removed the installation script.

+ The mode interactive takes values by default from the user configuration, if
any.

+ The paths use "path/filepath" to be compatible with Windows paths.

+ Updated Go API for templates.

+ Splitted in library and command, so the library could be used by a Go IDE.

+ Added flag "-config" to create an user configuration file.


### 2010-10-11  v0.9.8

+ The user configuration file is only added if it doesn't exist.

+ Changed to line comments.

+ At building a command project is added an install script.

+ Added GNU Lesser General Public License.

+ Removed package *log*.

+ The source file is named as the package name. The directory of source files is
named "cmd" when the project is a command application, and it is named as the
package name when it is a package.

+ Update fields related to the project name in file Metadata.


### 2010-09-15  v0.9.5

+ Added option *bazaar* for the version control.

+ At update a project, is replaced the URL in VCS configuration file.

+ Updated *goconfig* API.

+ In Metadata, the section *DEFAULT* has been renamed to *CORE* and the another
sections are in title case; it has been added option *vcs* in section *CORE*.


### 2010-09-05  v0.9

+ Initial release.
