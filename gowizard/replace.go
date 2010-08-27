// Copyright 2010  The "gowizard" Authors
//
// Use of this source code is governed by the Simplified BSD License
// that can be found in the LICENSE file.
//
// This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied. See the License
// for more details.

package main

import (
	"bufio"
	"bytes"
	"os"

	"fmt"
)


func ReplaceHeader(fname, comment, packageName string) os.Error {
	var output bytes.Buffer

	commentBi := []byte(comment)
	newLine := []byte{'\n'}

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	// Create a buffer
	fileBuf := bufio.NewReader(file)

	// === Check if the first bytes are comment characters.
	// ===
	for i, _ := range commentBi {
		firstByte, err := fileBuf.ReadByte()
		if err != nil {
			return err
		}

		if firstByte != commentBi[i] {
			return errNoHeader
		}
	}

	// Back to the beginning
	for i := 0; i < len(commentBi); i++ {
		fileBuf.UnreadByte()
	}

	// === Fill the buffer
	input := make([]byte, fileBuf.Buffered())

	if _, err := fileBuf.Read(input); err != nil {
		return err
	}

	// === Read line to line
	// ===
	var endHeader bool
	var year string

	lines := bytes.Split(input, newLine, -1)
	packageBi := []byte("package ")
	copyrightBi := []byte("opyright ")
	count := 0 // In what line ends the header.

	for i, line := range lines {
		// Skip header.
		if !endHeader {
			if n := bytes.Index(line, copyrightBi); n != -1 {
				s := bytes.Split(line, copyrightBi, -1)
				s = bytes.Fields(s[1]) // All after of "Copyright"
				year = string(s[0])    // The first one, so the year.
			}

			if !bytes.HasPrefix(line, commentBi) {
				count = i
				endHeader = true
			}
		}

		// The package line is after of header.
		if endHeader {
			count++

			// When the line is found, then adds the new package name.
			if bytes.HasPrefix(line, packageBi) {
				_, err := output.Write(packageBi)
				_, err = output.Write([]byte(packageName))
				err = output.WriteByte('\n')

				if err != nil {
					return err
				}
				break
			}

			// Add the another lines.
			_, err := output.Write(line)
			err = output.WriteByte('\n')

			if err != nil {
				return err
			}
		}

	}

	if _, err := output.Write(bytes.Join(lines[count:], newLine)); err != nil {
		return err
	}
	fmt.Println(output.String(), year, len(year))

	//fmt.Println(lines[count:], string(packageName), len(packageName))


	return nil
}

