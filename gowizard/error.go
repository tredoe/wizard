// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

/* Errors introduced by this package. */

package main

import (
	"os"
)


type error struct {
	os.ErrorString
}

var (
	errNoHeader = &error{"gowizard: no header with copyright"}
)


type goFileError string

func (self goFileError) String() string {
	return "gowizard: no Go source files in '" + string(self) + "'"
}

