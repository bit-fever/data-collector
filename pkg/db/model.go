//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

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

package db

import "time"

//=============================================================================
//===
//=== Entities
//===
//=============================================================================

type Common struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

//=============================================================================

type Product struct {
	Id                   uint    `json:"id" gorm:"primaryKey"`
	SourceId             uint    `json:"sourceId"`
	Symbol               string  `json:"symbol"`
	Username             string  `json:"username"`
	SystemCode           string  `json:"systemCode"`
	ConnectionCode       string  `json:"connectionCode"`
	SupportsMultipleData bool    `json:"supportsMultipleData"`
	Timezone             string  `json:"timezone"`
}

//=============================================================================

const InstrumentStatusReady      =  0
const InstrumentStatusProcessing =  1
const InstrumentStatusError      = -1

type Instrument struct {
	Id               uint    `json:"id" gorm:"primaryKey"`
	ProductId        uint    `json:"productId"`
	Symbol           string  `json:"symbol"`
	Name             string  `json:"name"`
	ExpirationDate   int     `json:"expirationDate,omitempty"`
	IsContinuous     bool    `json:"isContinuous"`
	Status           int8    `json:"status"`
	DataFrom         int     `json:"dataFrom"`
	DataTo           int     `json:"dataTo"`
}

//=============================================================================

const UploadJobStatusWaiting     = 0
const UploadJobStatusAdding      = 1
const UploadJobStatusAggregating = 2
const UploadJobStatusReady       = 3
const UploadJobStatusError       = -1

type UploadJob struct {
	Id               uint    `json:"id" gorm:"primaryKey"`
	InstrumentId     uint    `json:"instrumentId"`
	Status           int8    `json:"status"`
	Filename         string  `json:"filename"`
	Error            string  `json:"error"`
	Progress         int8    `json:"progress"`
	Records          int     `json:"records"`
	Bytes            int64   `json:"bytes"`
	Timezone         string  `json:"timezone"`
	Parser           string  `json:"parser"`
}

//=============================================================================

type LoadedPeriod struct {
	Id           uint   `json:"id" gorm:"primaryKey"`
	InstrumentId uint   `json:"instrumentId"`
	Day          int    `json:"day"`
	Status       int    `json:"status"`
}

//=============================================================================
//===
//=== Table names
//===
//=============================================================================

func (Product)      TableName() string { return "product"       }
func (Instrument)   TableName() string { return "instrument"    }
func (LoadedPeriod) TableName() string { return "loaded_period" }
func (UploadJob)    TableName() string { return "upload_job"    }

//=============================================================================
