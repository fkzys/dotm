/*
 Copyright (C) 2026 fkzys <fkzys@proton.me>
                    https://github.com/fkzys/dotm

 SPDX-License-Identifier: AGPL-3.0-or-later

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package tmpl

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// RenderFile reads a template file from disk and renders it with data.
func RenderFile(path string, data map[string]any) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Render(string(content), path, data)
}

// Render parses and executes a Go template string.
// name is used in error messages to identify which template failed.
func Render(content string, name string, data map[string]any) ([]byte, error) {
	tmpl, err := template.New(name).
		Funcs(FuncMap()).
		Option("missingkey=error").
		Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}
