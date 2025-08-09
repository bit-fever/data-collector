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

import (
	"time"
)

//=============================================================================

type Common struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

//=============================================================================
//===
//=== Data entities
//===
//=============================================================================

type DPStatus int

const (
	DPStatusFetchingInventory DPStatus = 1
	DPStatusFetchingData      DPStatus = 2
	DPStatusReady             DPStatus = 0
)

//-----------------------------------------------------------------------------

type DataProduct struct {
	Id                   uint     `json:"id" gorm:"primaryKey"`
	Username             string   `json:"username"`
	ConnectionCode       string   `json:"connectionCode"`
	SystemCode           string   `json:"systemCode"`
	Symbol               string   `json:"symbol"`
	SupportsMultipleData bool     `json:"supportsMultipleData"`
	Connected            bool     `json:"connected"`
	Timezone             string   `json:"timezone"`
	Status               DPStatus `json:"status"`
}

//=============================================================================

const InstrumentStatusReady      =  0
const InstrumentStatusProcessing =  1
const InstrumentStatusError      = -1

type DataInstrument struct {
	Id               uint       `json:"id" gorm:"primaryKey"`
	DataProductId    uint       `json:"dataProductId"`
	Symbol           string     `json:"symbol"`
	Name             string     `json:"name"`
	ExpirationDate   *time.Time `json:"expirationDate,omitempty"`
	Continuous       bool       `json:"continuous"`
	Status           int8       `json:"status"`
	DataFrom         int        `json:"dataFrom"`
	DataTo           int        `json:"dataTo"`
}

//=============================================================================

type DataInstrumentFull struct {
	DataInstrument
	ProductSymbol  string `json:"productSymbol"`
	SystemCode     string `json:"systemCode"`
	ConnectionCode string `json:"connectionCode"`
}

//=============================================================================

const UploadJobStatusWaiting     = 0
const UploadJobStatusAdding      = 1
const UploadJobStatusAggregating = 2
const UploadJobStatusReady       = 3
const UploadJobStatusError       = -1

type UploadJob struct {
	Id               uint    `json:"id" gorm:"primaryKey"`
	DataInstrumentId uint    `json:"dataInstrumentId"`
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
	Id               uint   `json:"id" gorm:"primaryKey"`
	DataInstrumentId uint   `json:"dataInstrumentId"`
	Day              int    `json:"day"`
	Status           int    `json:"status"`
}

//=============================================================================
//===
//=== Broker entities
//===
//=============================================================================

type BrokerProduct struct {
	Id               uint    `json:"id" gorm:"primaryKey"`
	Username         string  `json:"username"`
	ConnectionCode   string  `json:"connectionCode"`
	Symbol           string  `json:"symbol"`
	Name             string  `json:"name"`
	PointValue       float32 `json:"pointValue"`
	CostPerOperation float32 `json:"costPerOperation"`
	CurrencyCode     string  `json:"currencyCode"`
}

//=============================================================================
//===
//=== Bias analysis
//===
//=============================================================================

type BiasAnalysis struct {
	Common
	Username          string  `json:"username"`
	DataInstrumentId  uint    `json:"dataInstrumentId"`
	BrokerProductId   uint    `json:"brokerProductId"`
	Name              string  `json:"name"`
	Notes             string  `json:"notes"`
}

//=============================================================================

type BiasAnalysisFull struct {
	BiasAnalysis
	DataSymbol     string  `json:"dataSymbol"`
	DataName       string  `json:"dataName"`
	BrokerSymbol   string  `json:"brokerSymbol"`
	BrokerName     string  `json:"brokerName"`
}

//=============================================================================

type BiasConfig struct {
	Id              uint    `json:"id" gorm:"primaryKey"`
	BiasAnalysisId  uint    `json:"biasAnalysisId"`
	StartDay        int16   `json:"startDay"`
	StartSlot       int16   `json:"startSlot"`
	EndDay          int16   `json:"endDay"`
	EndSlot         int16   `json:"endSlot"`
	Months          int16   `json:"months"`
	Excludes        string  `json:"excludes"`
	Operation       int8    `json:"operation"`
	GrossProfit     float64 `json:"grossProfit"`
	NetProfit       float64 `json:"netProfit"`
}

//=============================================================================
//===
//=== Table names
//===
//=============================================================================

func (DataProduct)    TableName() string { return "data_product"    }
func (DataInstrument) TableName() string { return "data_instrument" }
func (BrokerProduct)  TableName() string { return "broker_product"  }
func (LoadedPeriod)   TableName() string { return "loaded_period"   }
func (UploadJob)      TableName() string { return "upload_job"      }
func (BiasAnalysis)   TableName() string { return "bias_analysis"   }
func (BiasConfig)     TableName() string { return "bias_config"     }

//=============================================================================
