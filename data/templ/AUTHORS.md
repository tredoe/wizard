###### Notice

*This is the official list of **{{.ProjectName}}** authors for copyright purposes.*

*This file is distinct from the CONTRIBUTORS file. See the latter for an
explanation.*

*Names should be added to this file as:*

	`Organization` or `Name <email address>`

*Please keep the list sorted.*

* * *

{{if .IsForOrg}}{{.OrgName}}{{else}}{{.Author}}{{with .Email}} <{{.}}>{{end}}{{end}}

