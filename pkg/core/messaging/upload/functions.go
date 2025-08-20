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

package upload

import (
	"errors"
	"strconv"
	"time"

	"github.com/bit-fever/core/datatype"
)

//=============================================================================

const Date  = "Date"
const Time  = "Time"
const Open  = "Open"
const High  = "High"
const Low   = "Low"
const Close = "Close"
const Up    = "Up"
const Down  = "Down"

//=============================================================================
//===
//=== Common functions
//===
//=============================================================================

func parseInt(value string, name string) (int, error) {
	res,err := strconv.Atoi(value)

	if err != nil {
		return 0, errors.New("Field '"+name+"' is not a valid integer")
	}

	return res, nil
}

//=============================================================================

func parseFloat(value string, name string) (float64, error) {
	res,err := strconv.ParseFloat(value, 64)

	if err != nil {
		return 0, errors.New("Field '"+name+"' is not a valid float")
	}

	return res, nil
}

//=============================================================================

func parseTimestamp(date string, hhmm string, loc *time.Location) (time.Time, error) {
	year, mon, day, err := parseDate(date)

	if err == nil {
		hh, mm, err := parseTime(hhmm)

		if err == nil {
			return time.Date(year, time.Month(mon), day, hh, mm, 0, 0, loc), nil
		}

		return time.Now(), err
	}

	return time.Now(), err
}

//=============================================================================

func parseDate(date string) (int, int, int, error) {
	if len(date) != 10 || date[2] != '/' || date[5] != '/' {
		return 0,0,0, errors.New("Field '"+ Date +"' has an invalid format")
	}

	sMon := date[0:2]
	sDay := date[3:5]
	sYear:= date[6:]

	if mon,err := strconv.Atoi(sMon); err == nil {
		if day,err := strconv.Atoi(sDay); err == nil {
			if year,err := strconv.Atoi(sYear); err == nil {
				return year, mon, day, nil
			}
		}
	}

	return 0,0,0, errors.New("Field '"+ Date +"' has an invalid format")
}

//=============================================================================

func parseTime(hhmm string) (int, int, error) {
	if len(hhmm) != 5 || hhmm[2] != ':' {
		return 0, 0, errors.New("Field '"+ Time +"' has an invalid format")
	}

	sHH := hhmm[0:2]
	sMM := hhmm[3:]

	if hh,err := strconv.Atoi(sHH); err == nil {
		if mm,err := strconv.Atoi(sMM); err == nil {
			return hh, mm, nil
		}
	}

	return 0,0, errors.New("Field '"+ Time +"' has an invalid format")
}

//=============================================================================

func updateDataRange(t time.Time, r *DataRange) {
	y, m, d := t.Date()
	date := datatype.IntDate(y*10000 + int(m)*100 + d)

	//--- Handle from day

	if r.FromDay.IsNil() || r.FromDay > date {
		r.FromDay = date
	}

	//--- Handle to day

	if r.ToDay.IsNil() || r.ToDay < date {
		r.ToDay = date
	}
}

//=============================================================================
