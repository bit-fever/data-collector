//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

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

package platform

import "time"

//=============================================================================

type InstrumentResponse struct {
	Offset   int           `json:"offset"`
	Limit    int           `json:"limit"`
	OverFlow bool          `json:"overFlow"`
	Result   []Instrument  `json:"result"`
}

//=============================================================================

type Instrument struct {
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Exchange        string     `json:"exchange"`
	Country         string     `json:"country"`
	Root            string     `json:"root"`
	ExpirationDate  *time.Time `json:"expirationDate"`
	PointValue      int        `json:"pointValue"`
	MinMove         float64    `json:"minMove"`
	Continuous      bool       `json:"continuous"`
}

//=============================================================================

type PriceBars struct {
	Symbol string      `json:"symbol"`
	Date   int         `json:"date"`
	Bars   []*PriceBar `json:"bars"`
	NoData bool        `json:"noData"`
}

//=============================================================================

type PriceBar struct {
	TimeStamp    time.Time
	High         float64
	Low          float64
	Open         float64
	Close        float64
	UpVolume     int
	DownVolume   int
	UpTicks      int
	DownTicks    int
	OpenInterest int
}

//=============================================================================
