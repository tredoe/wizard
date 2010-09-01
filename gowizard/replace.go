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
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Text to search.
var (
	bCommentCode   = []byte(CHAR_COMMENT_CODE)
	bCommentMake   = []byte(CHAR_COMMENT_MAKE)
	bCopyright     = []byte("opyright ")
	bEndOfNotice   = []byte("* * *")
	bPkgInCode     = []byte("package ")
	bPkgInMakefile = []byte("TARG=")
)

// Regular expressions
var reHeader = regexp.MustCompile(fmt.Sprintf("^%s+\n", CHAR_HEADER))


/* Replaces the project name on file `fname`. */
func replaceTextFile(fname, old string, new []byte, tag map[string]string, update map[string]bool) (err os.Error) {
	var bOldLicense, bNewLicense []byte
	var isReadme bool
	var output bytes.Buffer

	reFirstOldName := regexp.MustCompile(fmt.Sprintf("^%s\n", old))
	reLineOldName := regexp.MustCompile(fmt.Sprintf("[\"*'/, .]%s[\"*'/, .]", old))
	reOldName := regexp.MustCompile(old)

	if strings.HasPrefix(fname, README) {
		isReadme = true
	}

	if isReadme && update["License"] {
		bOldLicense = []byte(listLicense[cfg.License])
		bNewLicense = []byte(tag["license"])
	}

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, PERM_FILE)
	if err != nil {
		return err
	}

	defer file.Close()

	// === Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// === Read line to line
	isFirstLine := true

	for {
		line, err := rw.ReadSlice('\n')
		if err == os.EOF {
			break
		}

		// Write the line of the separator and exits of loop.
		if !isReadme && bytes.Index(line, bEndOfNotice) != -1 {
			if _, err := output.Write(line); err != nil {
				return err
			}
			break
		}

		if update["ProjectName"] {
			if isFirstLine {

				if reFirstOldName.Match(line) {
					newLine := reFirstOldName.ReplaceAll(line, new)
					if _, err := output.Write(newLine); err != nil {
						return err
					}

					// === Get the second line to change the header
					line2, err := rw.ReadSlice('\n')
					if err != nil {
						return err
					}

					if reHeader.Match(line2) {
					println("HEADER")
						newHeader := header(string(new))
						_, err := output.Write([]byte(newHeader))
						err = output.WriteByte('\n')

						if err != nil {
							return err
						}
						println("header changed")
					}
				} else {
					if _, err := output.Write(line); err != nil {
						return err
					}
				}

				isFirstLine = false
				continue
			}

			// === Not first line.

			// Search the old name of the project name.
			if reLineOldName.Match(line) {
				newLine := reOldName.ReplaceAll(line, new)
				if _, err := output.Write(newLine); err != nil {
					return err
				}
				continue
			}
		}

		if isReadme && update["License"] && bytes.Index(line, bOldLicense) != -1 {
			newLine := bytes.Replace(line, bOldLicense, bNewLicense, 1)

			if _, err := output.Write(newLine); err != nil {
				return err
			}
			continue
		}

		// Add lines that have not matched.
		if _, err := output.Write(line); err != nil {
			return err
		}
	}

	// === Get the remaining of the buffer.
	end := make([]byte, rw.Reader.Buffered())
	if _, err := rw.Read(end); err != nil {
		return err
	}

	if _, err = output.Write(end); err != nil {
		return err
	}

	// === Write changes to file

	// Set the new size of file.
	if err = file.Truncate(int64(len(output.Bytes()))); err != nil {
		return err
	}

	// Offset at the beggining of file.
	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	// Write buffer to file.
	if _, err = rw.Write(output.Bytes()); err != nil {
		return err
	}
	if err = rw.Writer.Flush(); err != nil {
		return err
	}

	return nil
}

/* Base to replace header and package name. */
func _replaceSourceFile(fname string, isCodeFile bool, comment, packageName []byte, tag map[string]string, update map[string]bool) (err os.Error) {
	var output bytes.Buffer

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, PERM_FILE)
	if err != nil {
		return err
	}

	defer file.Close()

	// === Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// === Check if the first bytes are comment characters.
	for i, _ := range comment {
		firstByte, err := rw.ReadByte()
		if err != nil {
			return err
		}

		if firstByte != comment[i] {
			return errNoHeader
		}
	}

	// Backs to the beginning
	for i := 0; i < len(comment); i++ {
		rw.UnreadByte()
	}

	// === Read line to line
	var endHeader, skipHeader bool

	if !update["ProjectName"] && !update["License"] {
		skipHeader = true
	}

	for {
		line, err := rw.ReadSlice('\n')
		if err == os.EOF {
			break
		}

		// The header.
		if !skipHeader && !endHeader {
			var header, year string

			// Search the year.
			if bytes.Index(line, bCopyright) != -1 {
				s := bytes.Split(line, bCopyright, -1)
				s = bytes.Fields(s[1]) // All after of "Copyright"
				year = string(s[0])    // The first one, so the year.
			}

			// End of header.
			if !bytes.HasPrefix(line, comment) {
				endHeader = true

				// Insert the new header using the year that it just be got.
				if isCodeFile {
					header, _ = renderCodeHeader(tag, year)
				} else {
					header, _ = renderMakeHeader(tag, year)
				}

				if _, err := output.Write([]byte(header)); err != nil {
					return err
				}
			}
		}

		// The package line is after of header.
		if skipHeader || endHeader {

			if isCodeFile {
				// When the line is found, then adds the new package name.
				if bytes.HasPrefix(line, bPkgInCode) {

					if !bytes.HasSuffix(line, packageName) {
						_, err := output.Write(bPkgInCode)
						_, err = output.Write(packageName)
						err = output.WriteByte('\n')

						if err != nil {
							return err
						}
					}

					break
				}
				// Makefile
			} else {
				if bytes.HasPrefix(line, bPkgInMakefile) {
					// Simple argument without full path to install via goinstall.
					if bytes.IndexByte(line, '/') != -1 {
						// Add character of new line for that the package name
						// can be matched correctly.
						old := []byte(cfg.PackageName + "\n")
						newLine := bytes.Replace(line, old, packageName, 1)

						_, err := output.Write(newLine)
						err = output.WriteByte('\n')

						if err != nil {
							return err
						}
					} else {
						_, err := output.Write(bPkgInMakefile)
						_, err = output.Write(packageName)
						err = output.WriteByte('\n')

						if err != nil {
							return err
						}
					}

					break
				}
			}

			// Add the another lines.
			if _, err := output.Write(line); err != nil {
				return err
			}
		}
	}

	// === Get the remaining of the buffer.
	end := make([]byte, rw.Reader.Buffered())
	if _, err := rw.Read(end); err != nil {
		return err
	}

	if _, err = output.Write(end); err != nil {
		return err
	}

	// === Write changes to file

	// Set the new size of file.
	if err = file.Truncate(int64(len(output.Bytes()))); err != nil {
		return err
	}

	// Offset at the beggining of file.
	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	// Write buffer to file.
	if _, err = rw.Write(output.Bytes()); err != nil {
		return err
	}
	if err = rw.Writer.Flush(); err != nil {
		return err
	}

	return nil
}

func replaceGoFile(fname string, packageName []byte, tag map[string]string, update map[string]bool) (err os.Error) {
	return _replaceSourceFile(fname, true, bCommentCode, packageName,
		tag, update)
}

func replaceMakefile(fname string, packageName []byte, tag map[string]string, update map[string]bool) (err os.Error) {
	return _replaceSourceFile(fname, false, bCommentMake, packageName,
		tag, update)
}

