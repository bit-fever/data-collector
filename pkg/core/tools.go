//=============================================================================
/*
Copyright © 2024 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package core

import (
	"strings"
)

//=============================================================================
//===
//=== BiasConfig encoding/decoding
//===
//=============================================================================

func EncodeMonths(months []bool) int16 {
	var value int16

	if months != nil && len(months) == 12 {
		for _, month := range months {
			value <<= 1
			if month {
				value |= 1
			}
		}
	}

	return value
}

//=============================================================================

func EncodeExcludes(list []string) string {
	var sb strings.Builder

	if list != nil {
		for i, exc := range list {
			if i != 0 {
				sb.WriteString("|")
			}

			sb.WriteString(exc)
		}
	}

	return sb.String()
}

//=============================================================================

func DecodeMonths(value int16) []bool {
	var list []bool
	var bit int16 = 1<<11

	for i:=0; i<12; i++ {
		month := (value & bit) != 0
		list = append(list, month)
		bit >>=1
	}

	return list
}

//=============================================================================

func DecodeExcludes(value string) []string {
	if len(value) == 0 {
		return []string{}
	}

	return strings.Split(value, "|")
}

//=============================================================================

func Trunc2d(value float64) float64 {
	return float64(int(value * 100)) / 100
}

//=============================================================================
