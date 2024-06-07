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

package messaging

import (
	"encoding/json"
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/data-collector/pkg/db"
	"gorm.io/gorm"
	"log/slog"
)

//=============================================================================

func InitMessageListener() {
	slog.Info("Starting message listeners...")

	go msg.ReceiveMessages(msg.QuInventoryUpdatesToCollector, handleMessage)
}

//=============================================================================

func handleMessage(m *msg.Message) bool {

	slog.Info("New message received", "origin", m.Origin, "type", m.Type, "source", m.Source)

	if m.Origin == msg.OriginDb {
		if m.Source == msg.SourceProductData {
			pdm := ProductDataMessage{}
			err := json.Unmarshal(m.Entity, &pdm)
			if err != nil {
				slog.Error("Dropping badly formatted message!", "entity", string(m.Entity))
				return true
			}

			if m.Type == msg.TypeCreate {
				return addProductData(&pdm)
			}
		}
	}

	slog.Error("Dropping message with unknown origin/type!", "origin", m.Origin, "type", m.Type)
	return true
}

//=============================================================================

func addProductData(pdm *ProductDataMessage) bool {
	slog.Info("addProductData: Product for data change received", "sourceId", pdm.ProductData.Id)

	err := db.RunInTransaction(func(tx *gorm.DB) error {
		pd := &db.ProductData{}

		pd.SourceId             = pdm.ProductData.Id
		pd.Symbol               = pdm.ProductData.Symbol
		pd.Username             = pdm.ProductData.Username
		pd.SystemCode           = pdm.Connection.SystemCode
		pd.ConnectionCode       = pdm.Connection.Code
		pd.SupportsMultipleData = pdm.Connection.SupportsMultipleData
		pd.Timezone             = pdm.Exchange.Timezone

		return db.AddProductData(tx, pd)
	})

	if err != nil {
		slog.Error("Raised error while processing message")
	} else {
		slog.Info("addProductData: Operation complete")
	}

	return err == nil
}

//=============================================================================
