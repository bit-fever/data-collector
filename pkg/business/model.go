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

import (
	"github.com/bit-fever/data-collector/pkg/ds"
	"time"
)

//=============================================================================
//===
//=== Upload spec & response
//===
//=============================================================================

type DatafileUploadSpec struct {
	Symbol        string `json:"symbol"       binding:"required"`
	Name          string `json:"name"         binding:"required"`
	Continuous    bool   `json:"continuous"   binding:"required"`
	FileTimezone  string `json:"fileTimezone" binding:"required"`
	Parser        string `json:"parser"       binding:"required"`
}

//=============================================================================

type DatafileUploadResponse struct {
	Duration int   `json:"duration"`
	Bytes    int64 `json:"bytes"`
}

//=============================================================================
//=== Get data response
//=============================================================================

type InstrumentDataSpec struct {
	Id        uint
	From      string
	To        string
	Timezone  string
	Reduction string
	Config    *ds.DataConfig
}

//=============================================================================

type InstrumentDataParams struct {
	Location  *time.Location
	From       time.Time
	To         time.Time
	Reduction  int
}

//=============================================================================

type InstrumentDataResponse struct {
	Id          uint            `json:"id"`
	Symbol      string          `json:"symbol"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Timeframe   string          `json:"timeframe"`
	Timezone    string          `json:"timezone"`
	Reduction   int             `json:"reduction,omitempty"`
	Reduced     bool            `json:"reduced"`
	Records     int             `json:"records"`
	DataPoints  []*ds.DataPoint `json:"dataPoints"`
}

//=============================================================================
