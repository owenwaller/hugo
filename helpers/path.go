// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Simple Public License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://opensource.org/licenses/Simple-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/viper"
)

var sanitizeRegexp = regexp.MustCompile("[^a-zA-Z0-9./_-]")

// Take a string with any characters and replace it so the string could be used in a path.
// MakePath creates a Unicode sanitized string, with the spaces replaced, whilst
// preserving the original casing of the string.
// E.g. Social Media -> Social-Media
func MakePath(s string) string {
	return UnicodeSanitize(strings.Replace(strings.TrimSpace(s), " ", "-", -1))
}

// MakePathToLower creates a Unicode santized string, with the spaces replaced,
// and transformed to lower case.
// E.g. Social Media -> social-media
func MakePathToLower(s string) string {
	return strings.ToLower(MakePath(s))
}

func MakeTitle(inpath string) string {
	return strings.Replace(strings.TrimSpace(inpath), "-", " ", -1)
}

// unused
//func Sanitize(s string) string {
//	return sanitizeRegexp.ReplaceAllString(s, "")
//}

func UnicodeSanitize(s string) string {
	source := []rune(s)
	target := make([]rune, 0, len(source))

	for _, r := range source {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '/' || r == '_' || r == '-' {
			target = append(target, r)
		}
	}

	return string(target)
}

func ReplaceExtension(path string, newExt string) string {
	f, _ := FileAndExt(path)
	return f + "." + newExt
}

// Check if Exists && is Directory
func DirExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func IsEmpty(path string) (bool, error) {
	if b, _ := Exists(path); !b {
		return false, fmt.Errorf("%q path does not exist", path)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		f, err := os.Open(path)
		// FIX: Resource leak - f.close() should be called here by defer or is missed
		// if the err != nil branch is taken.
		defer f.Close()
		if err != nil {
			return false, err
		}
		list, err := f.Readdir(-1)
		// f.Close() - see bug fix above
		return len(list) == 0, nil
	} else {
		return fi.Size() == 0, nil
	}
}

// Check if File / Directory Exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func AbsPathify(inPath string) string {
	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	return filepath.Clean(filepath.Join(viper.GetString("WorkingDir"), inPath))
}

func Filename(in string) (name string) {
	name, _ = FileAndExt(in)
	return
}

// FileAndExt returns the filename and any extension of a file path as
// two separate strings.
// If path, in, contains a directory name ending in a slash then
// both name and ext will be empty strings.
// If the path, in, is either the current directory, the parent
// directory or the root directory, or an empty string, then both
// name and ext will be empty strings.
// If the path, in, represents the path of a file without an extension
// then name will be the name of the file and ext will be an empty string.
// If the path, in, represents a filename with an extension then
// then name will be the filename minus any extension - including the dot
// and ext will contain the extension - minus the dot.
func FileAndExt(in string) (name string, ext string) {
	ext = path.Ext(in)
	base := path.Base(in) // path.Base strips any trailing slash!

	// No file name cases. These are defined as:
	// 1. any "in" path that ends in a os.PathSeparator i.e. "/" on linux
	// 2. any "base" consisting of just an os.PathSeparator
	// 3. any "base" consisting of just an empty string
	// 4. any "base" consisting of just the current directory i.e. "."
	// 5. any "base" consisting of just the parent directory i.e. ".."
	if (strings.LastIndex(in, string(os.PathSeparator)) == len(in)-1) || base == "" || base == "." || base == ".." || base == string(os.PathListSeparator) {
		name = "" // there is NO filename
	} else if ext != "" { // there was an Extension
		// return the filename minus the extension (and the ".")
		name = base[:strings.LastIndex(base, ".")]
	} else {
		// no extension case so just return base, which willi
		// be the filename
		name = base
	}
	return
}

func GuessSection(in string) string {
	x := strings.Split(in, "/")
	x = x[:len(x)-1]
	if len(x) == 0 {
		return ""
	}
	if x[0] == "content" {
		x = x[1:]
	}
	return path.Join(x...)
}

func PathPrep(ugly bool, in string) string {
	if ugly {
		return Uglify(in)
	} else {
		return PrettifyPath(in)
	}
}

// /section/name.html -> /section/name/index.html
// /section/name/  -> /section/name/index.html
// /section/name/index.html -> /section/name/index.html
func PrettifyPath(in string) string {
	if path.Ext(in) == "" {
		// /section/name/  -> /section/name/index.html
		if len(in) < 2 {
			return "/"
		}
		return path.Join(path.Clean(in), "index.html")
	} else {
		name, ext := FileAndExt(in)
		if name == "index" {
			// /section/name/index.html -> /section/name/index.html
			return path.Clean(in)
		} else {
			// /section/name.html -> /section/name/index.html
			return path.Join(path.Dir(in), name, "index"+ext)
		}
	}
}

func FindCWD() (string, error) {
	serverFile, err := filepath.Abs(os.Args[0])

	if err != nil {
		return "", fmt.Errorf("Can't get absolute path for executable: %v", err)
	}

	path := filepath.Dir(serverFile)
	realFile, err := filepath.EvalSymlinks(serverFile)

	if err != nil {
		if _, err = os.Stat(serverFile + ".exe"); err == nil {
			realFile = filepath.Clean(serverFile + ".exe")
		}
	}

	if err == nil && realFile != serverFile {
		path = filepath.Dir(realFile)
	}

	return path, nil
}

func SafeWriteToDisk(inpath string, r io.Reader) (err error) {
	dir, _ := filepath.Split(inpath)
	ospath := filepath.FromSlash(dir)

	if ospath != "" {
		err = os.MkdirAll(ospath, 0777) // rwx, rw, r
		if err != nil {
			return
		}
	}

	exists, err := Exists(inpath)
	if err != nil {
		return
	}
	if exists {
		return fmt.Errorf("%v already exists", inpath)
	}

	file, err := os.Create(inpath)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return
}

func WriteToDisk(inpath string, r io.Reader) (err error) {
	dir, _ := filepath.Split(inpath)
	ospath := filepath.FromSlash(dir)

	if ospath != "" {
		err = os.MkdirAll(ospath, 0777) // rwx, rw, r
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(inpath)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return
}
