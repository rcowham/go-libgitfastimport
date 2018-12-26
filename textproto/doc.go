// Copyright (C) 2017-2018  Luke Shumaker <lukeshu@lukeshu.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package textproto implements low-level details of the fast-import
// format.
//
// This package deals with parsing and marshalling idiosyncratic
// datatypes used by the format (Ident tuples, 18-bit Mode numbers,
// oddly-quoted Path strings), and abstracting over special-case
// commands that break the "line-based" nature of the format (the
// "data" command, responses to the "cat-blob" command).
package textproto
