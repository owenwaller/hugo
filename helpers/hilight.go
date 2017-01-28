// Copyright © 2013-14 Steve Francia <spf@spf13.com>.
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

// Rework things to use the highlight tool instead of Pygments
// See:
// http://www.andre-simon.de/doku/highlight/en/highlight.php

package helpers

import (
	"bytes"
	"os/exec"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

const highlightBin = "highlight"

// HasHighlight checks to see if highlight is installed and available
// on the system.
func HasHighlights() bool {
	if _, err := exec.LookPath(highlightBin); err != nil {
		return false
	}
	return true
}

// Hilight takes some code and returns highlighted code.
func Hilight(code string, lexer string, style string, lineNumbers string) string {

	if !HasHighlights() {
		jww.WARN.Println("Highlighting requires highlight to be installed and in the path")
		return code
	}

	var out bytes.Buffer
	var stderr bytes.Buffer

	/*
		style := viper.GetString("PygmentsStyle")

		noclasses := "true"
		if viper.GetBool("PygmentsUseClasses") {
			noclasses = "false"
		}

		cmd := exec.Command(pygmentsBin, "-l"+lexer, "-fhtml", "-O",
			fmt.Sprintf("style=%s,noclasses=%s,encoding=utf8", style, noclasses))
	*/
	// For some reason a blank line seems to be inserted into the code
	// BEFORE this point, but it is unclear where. This cause an problem when
	// you turn on line numbers because the initial blank line is counted.
	// To avoicd this we shop all leading and training spaces
	code = strings.TrimSpace(code)
	lexer = "--syntax=" + lexer
	style = "--style=" + style
	if lineNumbers == "y" || lineNumbers == "Y" {
		lineNumbers = "-l"
	} else {
		lineNumbers = ""
	}
	cmd := exec.Command(highlightBin, "--enclose-pre", "-O xhtml", lineNumbers, "-K=14", "--fragment", "--include-style", "--inline-css", lexer, style)
	cmd.Stdin = strings.NewReader(code)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		jww.ERROR.Printf("lexer: %v\ncode:\n%s\n%s", lexer, code, stderr.String())
		return code
	}

	return out.String()
}
