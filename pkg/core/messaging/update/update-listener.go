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

package update

import (
	"encoding/json"
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
	"log/slog"
)

//=============================================================================

func HandleUpdateMessage(m *msg.Message) bool {

	slog.Info("New message received", "source", m.Source,  "type", m.Type)

	if m.Source == msg.SourceDataProduct {
		dpm := DataProductMessage{}
		err := json.Unmarshal(m.Entity, &dpm)
		if err != nil {
			slog.Error("Dropping badly data formatted message (DataProduct)!", "entity", string(m.Entity))
			return true
		}

		if m.Type == msg.TypeCreate {
			return addDataProduct(&dpm)
		}

	} else if m.Source == msg.SourceBrokerProduct {
		bpm := BrokerProductMessage{}
		err := json.Unmarshal(m.Entity, &bpm)
		if err != nil {
			slog.Error("Dropping badly formatted message (BrokerProduct)!", "entity", string(m.Entity))
			return true
		}

		if m.Type == msg.TypeCreate || m.Type == msg.TypeUpdate {
			return setBrokerProduct(&bpm)
		}
	}

	slog.Error("Dropping message with unknown source/type!", "source", m.Source, "type", m.Type)
	return true
}

//=============================================================================

func addDataProduct(dpm *DataProductMessage) bool {
	slog.Info("addDataProduct: Data product change received", "id", dpm.DataProduct.Id)

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		pd := &db.DataProduct{}

		pd.Id                   = dpm.DataProduct.Id
		pd.Symbol               = dpm.DataProduct.Symbol
		pd.Username             = dpm.DataProduct.Username
		pd.SystemCode           = dpm.Connection.SystemCode
		pd.ConnectionCode       = dpm.Connection.Code
		pd.Connected            = dpm.Connection.Connected
		pd.SupportsMultipleData = dpm.Connection.SupportsMultipleData
		pd.Timezone             = dpm.Exchange.Timezone

		return db.AddDataProduct(tx, pd)
	})

	if err != nil {
		slog.Error("Raised error while processing message")
	} else {
		slog.Info("addDataProduct: Operation complete")
	}

	return err == nil
}

//=============================================================================

func setBrokerProduct(bpm *BrokerProductMessage) bool {
	slog.Info("setBrokerProduct: Broker product change received", "id", bpm.BrokerProduct.Id)

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		bp := &db.BrokerProduct{}

		bp.Id               = bpm.BrokerProduct.Id
		bp.Symbol           = bpm.BrokerProduct.Symbol
		bp.Username         = bpm.BrokerProduct.Username
		bp.Name             = bpm.BrokerProduct.Name
		bp.PointValue       = bpm.BrokerProduct.PointValue
		bp.CostPerOperation = bpm.BrokerProduct.CostPerOperation
		bp.CurrencyCode     = bpm.Currency.Code

		return db.UpdateBrokerProduct(tx, bp)
	})

	if err != nil {
		slog.Error("Raised error while processing message")
	} else {
		slog.Info("setBrokerProduct: Operation complete")
	}

	return err == nil
}

//=============================================================================
