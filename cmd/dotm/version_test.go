/*
 Copyright (C) 2026 fkzys <fkzys@proton.me>
                    https://github.com/fkzys/dotm

 SPDX-License-Identifier: AGPL-3.0-or-later

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestCmdVersion(t *testing.T) {
	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdVersion()

	if err := w.Close(); err != nil {
		t.Fatalf("w.Close: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	output := buf.String()

	if output != "dotm dev\n" {
		t.Errorf("output = %q, want %q", output, "dotm dev\n")
	}
}
