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

	"github.com/bit-fever/core/datatype"
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

type DPRollTrigger string

const (
	DPRollTriggerSD4  = "sd4"
	DPRollTriggerSD6  = "sd6"
	DPRollTriggerSD30 = "sd30"
)

//-----------------------------------------------------------------------------

type DataProduct struct {
	Id                   uint          `json:"id" gorm:"primaryKey"`
	Username             string        `json:"username"`
	ConnectionCode       string        `json:"connectionCode"`
	SystemCode           string        `json:"systemCode"`
	Symbol               string        `json:"symbol"`
	SupportsMultipleData bool          `json:"supportsMultipleData"`
	Connected            bool          `json:"connected"`
	Timezone             string        `json:"timezone"`
	Status               DPStatus      `json:"status"`
	Months               string        `json:"months"`
	RolloverTrigger      DPRollTrigger `json:"rollTrigger"`
}

//=============================================================================

type DataInstrument struct {
	Id               uint      `json:"id" gorm:"primaryKey"`
	DataProductId    uint      `json:"dataProductId"`
	DataBlockId     *uint      `json:"dataBlockId"`
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	ExpirationDate  *time.Time `json:"expirationDate,omitempty"`
	RolloverDate    *time.Time `json:"rolloverDate,omitempty"`
	Continuous       bool      `json:"continuous"`
	Month            string    `json:"month"`
	RolloverDelta    float64   `json:"rolloverDelta"`
}

//=============================================================================

type DataInstrumentFull struct {
	DataInstrument
	ProductSymbol  string `json:"productSymbol"`
	SystemCode     string `json:"systemCode"`
	ConnectionCode string `json:"connectionCode"`
}

//=============================================================================

type DataInstrumentExt struct {
	DataInstrument
	Status    *DBStatus         `json:"status"`
	DataFrom   datatype.IntDate `json:"dataFrom"`
	DataTo     datatype.IntDate `json:"dataTo"`
	Progress  *int8             `json:"progress"`
}

//=============================================================================

type DBStatus int

const (
	DBStatusWaiting    =  1
	DBStatusLoading    =  2
	DBStatusProcessing =  3
	DBStatusSleeping   =  4
	DBStatusEmpty      =  5
	DBStatusReady      =  0
	DBStatusError      = -1
)

//-----------------------------------------------------------------------------

type DataBlock struct {
	Id             uint             `json:"id" gorm:"primaryKey"`
	SystemCode     string           `json:"systemCode"`
	Root           string           `json:"root"`
	Symbol         string           `json:"symbol"`
	Status         DBStatus         `json:"status"`
	Global         bool             `json:"global"`
	DataFrom       datatype.IntDate `json:"dataFrom"`
	DataTo         datatype.IntDate `json:"dataTo"`
	Progress       int8             `json:"progress"`
}

//=============================================================================

type IngestionJob struct {
	Id                uint    `json:"id" gorm:"primaryKey"`
	DataInstrumentId  uint    `json:"dataInstrumentId"`
	DataBlockId       uint    `json:"dataBlockId"`
	Filename          string  `json:"filename"`
	Records           int     `json:"records"`
	Bytes             int64   `json:"bytes"`
	Timezone          string  `json:"timezone"`
	Parser            string  `json:"parser"`
	Error             string  `json:"error"`
}

//=============================================================================

type DJStatus int

const (
	DJStatusWaiting    DJStatus =  0
	DJStatusRunning    DJStatus =  1
	DJStatusError      DJStatus = -1
)

//-----------------------------------------------------------------------------

type DownloadJob struct {
	Id                uint             `json:"id" gorm:"primaryKey"`
	DataInstrumentId  uint             `json:"dataInstrumentId"`
	DataBlockId       uint             `json:"dataBlockId"`
	Status            DJStatus         `json:"status"`
	LoadFrom          datatype.IntDate `json:"loadFrom"`
	LoadTo            datatype.IntDate `json:"loadTo"`
	Priority          int              `json:"priority"`
	UserConnection    string           `json:"userConnection"`
	CurrDay           int              `json:"currDay"`
	TotDays           int              `json:"totDays"`
	Error             string           `json:"error"`
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
func (DataBlock)      TableName() string { return "data_block"      }
func (BrokerProduct)  TableName() string { return "broker_product"  }
func (IngestionJob)   TableName() string { return "ingestion_job"   }
func (DownloadJob)    TableName() string { return "download_job"    }
func (BiasAnalysis)   TableName() string { return "bias_analysis"   }
func (BiasConfig)     TableName() string { return "bias_config"     }

//=============================================================================
