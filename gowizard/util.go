// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)


/* Copy a file from the data directory to the project. */
func copy(destinationFile, sourceFile string) {
	src, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Exit(err)
	}

	err = ioutil.WriteFile(destinationFile, src, _PERM_FILE)
	if err != nil {
		log.Exit(err)
	}
}

/* Reads the standard input until Return is pressed. */
func read() string {
	stdin := bufio.NewReader(os.Stdin)

	input, err := stdin.ReadString('\n')
	if err != nil {
		log.Exit(err)
	}

	return strings.TrimRight(input, "\n")
}

