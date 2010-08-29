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
)


/* Base to replace header and package name. */
func _replaceSourceFile(fname string, isCodeFile bool, comment, packageName []byte,
tag map[string]string, update map[string]bool) (newFile []byte, err os.Error) {
	var output bytes.Buffer

	copyrightBi := []byte("opyright ")
	newLineBi := []byte{'\n'}

	// Lines where replace the package name.
	pkgInCodeBi := []byte("package ")
	pkgInMakefileBi := []byte("TARG=")

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Create a buffer
	fileBuf := bufio.NewReader(file)

	// === Check if the first bytes are comment characters.
	for i, _ := range comment {
		firstByte, err := fileBuf.ReadByte()
		if err != nil {
			return nil, err
		}

		if firstByte != comment[i] {
			return nil, errNoHeader
		}
	}

	// Backs to the beginning
	for i := 0; i < len(comment); i++ {
		fileBuf.UnreadByte()
	}

	// === Fill the buffer
	input := make([]byte, fileBuf.Buffered())

	if _, err := fileBuf.Read(input); err != nil {
		return nil, err
	}

	// === Read line to line
	var endHeader, skipHeader bool

	lines := bytes.Split(input, newLineBi, -1)
	count := 0 // In what line ends the header.

	if !update["ProjectName"] && !update["License"] {
		skipHeader = true
	}

	for i, line := range lines {

		// The header.
		if !skipHeader && !endHeader {
			var header, year string

			// Search the year.
			if n := bytes.Index(line, copyrightBi); n != -1 {
				s := bytes.Split(line, copyrightBi, -1)
				s = bytes.Fields(s[1]) // All after of "Copyright"
				year = string(s[0])    // The first one, so the year.
			}

			// End of header.
			if !bytes.HasPrefix(line, comment) {
				endHeader = true
				count = i

				// Insert the new header using the year that it just be got.
				if isCodeFile {
					header, _ = renderHeaderCode(tag, year)
				} else {
					header, _ = renderHeaderMakefile(tag, year)
				}

				if _, err := output.Write([]byte(header)); err != nil {
					return nil, err
				}
			}
		}

		// The package line is after of header.
		if skipHeader || endHeader {

			if isCodeFile {
				// When the line is found, then adds the new package name.
				if bytes.HasPrefix(line, pkgInCodeBi) {

					if !bytes.HasSuffix(line, packageName) {
						_, err := output.Write(pkgInCodeBi)
						_, err = output.Write(packageName)
						err = output.WriteByte('\n')

						if err != nil {
							return nil, err
						}
					}

					count++
					break
				}
			// Makefile
			} else {
				if bytes.HasPrefix(line, pkgInMakefileBi) {
					// Simple argument without full path to install via goinstall.
					if bytes.IndexByte(line, '/') != -1 {
						// Add character of new line for that the package name
						// can be matched correctly.
						line = bytes.Add(line, newLineBi)
						old := []byte(cfg.PackageName + "\n")

						newLine := bytes.Replace(line, old, packageName, 1)

						_, err := output.Write(newLine)
						err = output.WriteByte('\n')

						if err != nil {
							return nil, err
						}
					} else {
						_, err := output.Write(pkgInMakefileBi)
						_, err = output.Write(packageName)
						err = output.WriteByte('\n')

						if err != nil {
							return nil, err
						}
					}

					count++
					break
				}
			}

			// Add the another lines.
			_, err := output.Write(line)
			err = output.WriteByte('\n')

			if err != nil {
				return nil, err
			}
			count++
		}
	}

	if _, err := output.Write(bytes.Join(lines[count:], newLineBi)); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func replaceCode(fname string, packageName []byte,
tag map[string]string, update map[string]bool) (newFile []byte, err os.Error) {
	return _replaceSourceFile(fname, true, commentCodeBi, packageName,
		tag, update)
}

func replaceMakefile(fname string, packageName []byte,
tag map[string]string, update map[string]bool) (newFile []byte, err os.Error) {
	return _replaceSourceFile(fname, false, commentMakefileBi, packageName,
		tag, update)
}

