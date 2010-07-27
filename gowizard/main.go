// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the 2-clause BSD License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"template"
	"time"
)


var license = map[string]string{
	"apache": "Apache (version 2.0)",
	"bsd-2":  "Simplified BSD",
	//"bsd-3":  "New BSD",
	"cc0":    "Creative Commons CC0 1.0 Universal",
}

// Headers for source code files
var (
	header    = `// Copyright {year}, The '{project}' Authors.  All rights reserved.
// Use of this source code is governed by the {license} License
// that can be found in the LICENSE file.
`
	headerCC0 = `// To the extent possible under law, Authors have waived all copyright and
// related or neighboring rights to '{project}'.
`
)

// Flags for the command line
var (
	fDebug   = flag.Bool("d", false, "debug mode")
	fList    = flag.Bool("l", false, "show the list of licenses for the flag `license`")
	fProject = flag.String("project", "", "the name of the project (e.g. 'goweb-foo')")
	fPkg     = flag.String("pkg", "", "the name of the package (e.g. 'foo')")
	fLicense = flag.String("license", "bsd-2", "the kind of license")
)


func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gowizard -project [-license]\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func checkFlags() {
	reGo := regexp.MustCompile(`^go`) // To remove it from the project name

	flag.Usage = usage
	flag.Parse()

	if *fList {
		fmt.Printf("Licenses\n\n")
		for k, v := range license {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	*fLicense = strings.ToLower(*fLicense)
	if _, present := license[*fLicense]; !present {
		log.Exitf("license unavailable %s", *fLicense)
	}

	if *fProject == "" {
		usage()
	}

	if *fPkg == "" {
		// The package name is created:
		// getting the last string after of the dash ('-'), if any,
		// and removing 'go'.
		pkg := strings.Split(*fProject, "-", -1)
		*fPkg = reGo.ReplaceAllString(strings.ToLower(pkg[len(pkg)-1]), "")
	} else {
		*fPkg = strings.ToLower(*fPkg)
	}

	return
}


// Main program execution
func main() {
	var t *template.Template

	checkFlags()

	tag := map[string]string{
		"license": license[*fLicense],
		"pkg":     *fPkg,
		"project": *fProject,
	}

	if *fDebug {
		fmt.Printf("Debug\n\n")
		for k, v := range tag {
			fmt.Printf("  %s: %s\n", k, v)
		}
		os.Exit(0)
	}

	if *fLicense == "cc0" {
		t = template.MustParse(headerCC0, nil)
	} else {
		t = template.MustParse(header, nil)
		tag["year"] = strconv.Itoa64(time.LocalTime().Year)
	}
	t.Execute(tag, os.Stdout)
	//fmt.Fprint(os.Stdout, )

	//t = template.MustParseFile("tmpl-setup.txt", nil)
	//t.Execute(data, c)

}

