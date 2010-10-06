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


// Replaces the project name on file `fname`.
func replaceTextFile(fname string, projectName []byte, cfg *Metadata,
tag map[string]string, update map[string]bool) os.Error {
	var isReadme bool
	var oldLicense, newLicense []byte
	var output bytes.Buffer

	// Text to search
	endOfNotice := []byte("* * *")

	// === Regular expressions
	// Header under the project name.
	reHeader := regexp.MustCompile(fmt.Sprintf("^%c+\n", CHAR_HEADER))

	reFirstOldName := regexp.MustCompile(fmt.Sprintf("^%s\n", cfg.ProjectName))
	reLineOldName := regexp.MustCompile(
		fmt.Sprintf("[\"*'/, .]%s[\"*'/, .]", cfg.ProjectName))
	reOldName := regexp.MustCompile(cfg.ProjectName)

	if strings.HasPrefix(fname, README) {
		isReadme = true
	}

	if isReadme && update["License"] {
		oldLicense = []byte(listLicense[cfg.License])
		newLicense = []byte(tag["license"])
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
		if !isReadme && bytes.Index(line, endOfNotice) != -1 {
			if _, err := output.Write(line); err != nil {
				return err
			}
			break
		}

		if update["ProjectName"] {
			if isFirstLine {

				if reFirstOldName.Match(line) {
					newLine := reFirstOldName.ReplaceAll(line, projectName)
					_, err := output.Write(newLine)
					err = output.WriteByte('\n')

					if err != nil {
						return err
					}

					// === Get the second line to change the header
					line, err := rw.ReadSlice('\n')
					if err != nil {
						return err
					}

					if reHeader.Match(line) {
						newHeader := header(string(projectName))
						_, err := output.Write([]byte(newHeader))
						err = output.WriteByte('\n')

						if err != nil {
							return err
						}
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
				newLine := reOldName.ReplaceAll(line, projectName)
				if _, err := output.Write(newLine); err != nil {
					return err
				}
				continue
			}
		}

		if isReadme && update["License"] && bytes.Index(line, oldLicense) != -1 {
			newLine := bytes.Replace(line, oldLicense, newLicense, 1)

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

	if err := rewrite(file, rw, &output); err != nil {
		return err
	}

	return nil
}

// Base to replace both header and package name.
func _replaceSourceFile(fname string, isCodeFile bool, comment, packageName []byte,
cfg *Metadata, tag map[string]string, update map[string]bool) os.Error {
	var output bytes.Buffer

	// Text to search
	var (
		copyright     = []byte("opyright ")
		LF            = []byte{'\n'}
		pkgInCode     = []byte("package ")
		pkgInMakefile = []byte("TARG=")
	)

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, PERM_FILE)
	if err != nil {
		return err
	}

	defer file.Close()

	// Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// === Check if the first bytes are comment characters.
	fileComment, err := rw.Peek(len(comment))
	if err != nil {
		return err
	}

	if !bytes.Equal(comment, fileComment) {
		return errNoHeader
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
			if bytes.Index(line, copyright) != -1 {
				s := bytes.Split(line, copyright, -1)
				s = bytes.Fields(s[1]) // All after of "Copyright"
				year = string(s[0])    // The first one, so the year.
			}

			// End of header.
			if bytes.Equal(line, LF) {
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
			if !update["PackageInCode"] {
				break
			}

			if isCodeFile {
				// When the line is found, then adds the new package name.
				if bytes.HasPrefix(line, pkgInCode) {

					if !bytes.HasSuffix(line, packageName) {
						_, err := output.Write(pkgInCode)
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
				if bytes.HasPrefix(line, pkgInMakefile) {
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
						_, err := output.Write(pkgInMakefile)
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

	if err := rewrite(file, rw, &output); err != nil {
		return err
	}

	return nil
}

func replaceGoFile(fname string, packageName []byte, cfg *Metadata,
tag map[string]string, update map[string]bool) os.Error {
	return _replaceSourceFile(fname, true, []byte(CHAR_CODE_COMMENT),
		packageName, cfg, tag, update)
}

func replaceMakefile(fname string, packageName []byte, cfg *Metadata,
tag map[string]string, update map[string]bool) os.Error {
	return _replaceSourceFile(fname, false, []byte(CHAR_MAKE_COMMENT),
		packageName, cfg, tag, update)
}

// Replaces the project name from URL configured in the Version Control System.
func replaceVCS_URL(fname, oldProjectName, newProjectName, vcs string) os.Error {
	var isHeader bool
	var output bytes.Buffer
	var header, option_1, option_2 []byte

	// === Text to search
	oldProjec := []byte("/" + oldProjectName)
	newProjec := []byte("/" + newProjectName)

	if vcs == "git" {
		header = []byte("[remote \"origin\"]\n")
		option_1 = []byte("\turl =")
		oldProjec = []byte(oldProjectName + ".git")
		newProjec = []byte(newProjectName + ".git")
	} else if vcs == "hg" {
		header = []byte("[paths]\n")
		option_1 = []byte("default-push =")
		option_2 = []byte("default =")
	} else if vcs == "bzr" {
		option_1 = []byte("push_location =")
		option_2 = []byte("parent_location =")
	}

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, PERM_FILE)
	if err != nil {
		return err
	}

	defer file.Close()

	// Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	if len(option_2) == 0 {
		// Read line to line
		for {
			line, err := rw.ReadSlice('\n')
			if err == os.EOF {
				break
			}

			// Header found
			if !isHeader && bytes.Equal(line, header) {
				isHeader = true
			}

			// Line found
			if isHeader && bytes.HasPrefix(line, option_1) {
				newLine := bytes.Replace(line, oldProjec, newProjec, 1)

				if _, err := output.Write(newLine); err != nil {
					return err
				}
				break
			}

			// Add the another lines.
			if _, err := output.Write(line); err != nil {
				return err
			}
		}
	// Could have two options
	} else {
		var isOption_1, isLine bool

		isRound_1 := true

		// In the first, is searched `option_1` else `option_2`.
		for {
			line, err := rw.ReadSlice('\n')
			if err == os.EOF {
				if isRound_1 {
					isRound_1, isHeader, isLine = false, false, false

					// === Reload the file again.
					if _, err := file.Seek(0, 0); err != nil {
						return err
					}

					rw.Reader = bufio.NewReader(file)
					line, _ = rw.ReadSlice('\n')
				// Round 2
				} else {
					break
				}
			}

			// Header found
			if !isHeader && (len(header) == 0 || bytes.Equal(line, header)) {
				isHeader = true
			}

			if isRound_1 && isHeader && bytes.HasPrefix(line, option_1) {
				isOption_1 = true
			}

			// Round 2
			if !isRound_1 {
				if isHeader {
					if isOption_1 {
						// Push line found
						if bytes.HasPrefix(line, option_1) {
							isLine = true
						}
					} else {
						// Line found
						if bytes.HasPrefix(line, option_2) {
							isLine = true
						}
					}
				}

				if isLine {
					newLine := bytes.Replace(line, oldProjec, newProjec, 1)

					if _, err := output.Write(newLine); err != nil {
						return err
					}
					break
				}

				// Add the another lines.
				if _, err := output.Write(line); err != nil {
					return err
				}
			}
		}
	}

	if err := rewrite(file, rw, &output); err != nil {
		return err
	}

	return nil
}


// === Utility
// ===

// Gets the remaining of file buffer to add it to the output buffer. Finally
// it is saved in the original file.
func rewrite(file *os.File, rw *bufio.ReadWriter, output *bytes.Buffer) os.Error {
	// === Get the remaining of the buffer.
	end := make([]byte, rw.Reader.Buffered())
	if _, err := rw.Read(end); err != nil {
		return err
	}

	if _, err := output.Write(end); err != nil {
		return err
	}

	// === Write changes to file

	// Set the new size of file.
	if err := file.Truncate(int64(len(output.Bytes()))); err != nil {
		return err
	}

	// Offset at the beggining of file.
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Write buffer to file.
	if _, err := rw.Write(output.Bytes()); err != nil {
		return err
	}
	if err := rw.Writer.Flush(); err != nil {
		return err
	}

	return nil
}

