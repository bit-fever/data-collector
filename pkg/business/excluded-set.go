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

package business

import "strconv"

//=============================================================================
//===
//=== ExcludedPeriod
//===
//=============================================================================

type ExcludedPeriod struct {
	Year  int16
	Month int16
}

//=============================================================================

func NewExcludedPeriod(value string) (*ExcludedPeriod, error) {
	if len(value) == 4 {
		y, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		return &ExcludedPeriod{
			Year : int16(y),
			Month: 0,
		}, nil
	}

	//--- The period is YYYY-[M]M

	y, err1 := strconv.Atoi(value[0:4])
	m, err2 := strconv.Atoi(value[5:])

	if err1 != nil {
		return nil, err1
	}

	if err2 != nil {
		return nil, err2
	}

	return &ExcludedPeriod{
		Year : int16(y),
		Month: int16(m),
	}, nil
}

//=============================================================================

func (ep * ExcludedPeriod) ShouldBeExcluded(month, year int16) bool {
	if year == ep.Year {
		return (ep.Month == 0) || (ep.Month == month)
	}

	return false
}

//=============================================================================
//===
//=== ExcludedPeriod
//===
//=============================================================================

type ExcludedSet struct {
	periods []*ExcludedPeriod
}

//=============================================================================

func NewExcludedSet(items []string) (*ExcludedSet, error) {
	var periods []*ExcludedPeriod

	for _, item := range items {
		ep, err := NewExcludedPeriod(item)

		if err != nil {
			return nil, err
		}

		periods = append(periods, ep)
	}

	return &ExcludedSet{
		periods: periods,
	}, nil
}

//=============================================================================

func (es *ExcludedSet) ShouldBeExcluded(month, year int16) bool {
	for _, ep := range es.periods {
		if ep.ShouldBeExcluded(month, year) {
			return true
		}
	}

	return false
}

//=============================================================================
