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

import (
	"github.com/bit-fever/core"
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/datatype"
	"github.com/bit-fever/core/req"
	"strconv"
)

//=============================================================================

var platform *core.Platform

//=============================================================================
//===
//=== Init
//===
//=============================================================================

func Init(p *core.Platform) {
	platform = p
}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func GetInstruments(username string, connectionCode string, root string) ([]Instrument, error) {
	var res InstrumentResponse

	token,err := auth.Token()
	if err != nil {
		return nil, err
	}

	client :=req.GetClient("bf")
	url := platform.System +"/v1/connections/"+connectionCode+"/roots/"+ root +"/instruments"
	err = req.DoGetOnBehalfOf(client, url, &res, token, username)

	if err != nil {
		return nil,err
	}

	return res.Result, nil
}

//=============================================================================

func GetPriceBars(username string, connectionCode string, symbol string, date datatype.IntDate) (*PriceBars, error) {
	var res PriceBars

	token,err := auth.Token()
	if err != nil {
		return nil, err
	}

	client :=req.GetClient("bf")
	url := platform.System +"/v1/connections/"+connectionCode+"/instruments/"+ symbol +"/bars?date="+strconv.Itoa(int(date))
	err = req.DoGetOnBehalfOf(client, url, &res, token, username)

	if err != nil {
		return nil,err
	}

	return &res, nil
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================
