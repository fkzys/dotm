/*
 Copyright (C) 2026 fkzys <fkzys@proton.me>
                    https://github.com/fkzys/dotm

 SPDX-License-Identifier: AGPL-3.0-or-later

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import "fmt"

// Set at build time via: -ldflags "-X main.version=..."
var version = "dev"

func cmdVersion() {
	fmt.Println("dotm " + version)
}
