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
)


/* Replace */
func replaceReadme(fname, old string, new []byte) (err os.Error) {
	var output bytes.Buffer

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	// === Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// === Read line to line
	for {
		line, err := rw.ReadSlice('\n')
		if err != nil {
			if err == os.EOF {
				break
			} else {
				return err
			}
		}

		println(line)

	}

	// === Write changes to file
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := rw.Write(output.Bytes()); err != nil {
		return err
	}
	rw.Writer.Flush()

	println("File changed", fname)

	return nil
}

/* Replaces the project name on file `fname`. */
func replaceProjectName(fname, old string, new []byte) (err os.Error) {
	var output bytes.Buffer

	bEndOfNotice := []byte("* * *")
	reFullOld := regexp.MustCompile(fmt.Sprint(`["*', .]`, old, `["*', .]`))
	reOld := regexp.MustCompile(old)

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	// === Buffered I/O
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))

	// === Read line to line
	for {
		line, err := rw.ReadSlice('\n')
		if err != nil {
			if err == os.EOF {
				break
			} else {
				return err
			}
		}

		if s := bytes.Index(line, bEndOfNotice); s != -1 {
			break
		}

		// Search the old name of the project name.
		if reFullOld.Match(line)  {
			newLine := reOld.ReplaceAll(line, new)
			if _, err := output.Write(newLine); err != nil {
				return err
			}
		} else {
			if _, err := output.Write(line); err != nil {
				return err
			}
		}
	}
/*
	// === Get the remaining of the buffer.
	end := make([]byte, rw.Reader.Buffered())
	if _, err := rw.Read(end); err != nil {
		return err
	}

	if _, err = output.Write(end); err != nil {
		return err
	}
*/
	// === Write changes to file
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := rw.Write(output.Bytes()); err != nil {
		return err
	}
	rw.Writer.Flush()

	println("File changed", fname)

	return nil
}

/* Base to replace header and package name. */
func _replaceSourceFile(fname string, isCodeFile bool, comment, packageName []byte,
tag map[string]string, update map[string]bool) (err os.Error) {
	var output bytes.Buffer

	bCopyright := []byte("opyright ")

	// Lines where replace the package name.
	bPkgInCode := []byte("package ")
	bPkgInMakefile := []byte("TARG=")

	// === Read file
	file, err := os.Open(fname, os.O_RDWR, 0644)
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
		if err != nil {
			if err == os.EOF {
				break
			} else {
				return err
			}
		}

		// The header.
		if !skipHeader && !endHeader {
			var header, year string

			// Search the year.
			if n := bytes.Index(line, bCopyright); n != -1 {
				s := bytes.Split(line, bCopyright, -1)
				s = bytes.Fields(s[1]) // All after of "Copyright"
				year = string(s[0])    // The first one, so the year.
			}

			// End of header.
			if !bytes.HasPrefix(line, comment) {
				endHeader = true

				// Insert the new header using the year that it just be got.
				if isCodeFile {
					header, _ = renderHeaderCode(tag, year)
				} else {
					header, _ = renderHeaderMakefile(tag, year)
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
/*
	// === Get the remaining of the buffer.
	end := make([]byte, rw.Reader.Buffered())
	if _, err := rw.Read(end); err != nil {
		return err
	}

	if _, err = output.Write(end); err != nil {
		return err
	}
*/
	// === Write changes to file
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := rw.Write(output.Bytes()); err != nil {
		return err
	}
	rw.Writer.Flush()

	println("File changed", fname)

	return nil
}

func replaceCode(fname string, packageName []byte,
tag map[string]string, update map[string]bool) (err os.Error) {
	return _replaceSourceFile(fname, true, bCommentCode, packageName,
		tag, update)
}

func replaceMakefile(fname string, packageName []byte,
tag map[string]string, update map[string]bool) (err os.Error) {
	return _replaceSourceFile(fname, false, bCommentMakefile, packageName,
		tag, update)
}

