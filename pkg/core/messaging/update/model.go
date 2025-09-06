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

package update

import "github.com/bit-fever/data-collector/pkg/db"

//=============================================================================
//===
//=== General entities
//===
//=============================================================================

type Currency struct {
	Id   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

//=============================================================================

type Connection struct {
	Id                   uint    `json:"id"`
	Username             string  `json:"username"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	SystemCode           string  `json:"systemCode"`
	SystemName           string  `json:"systemName"`
	Connected            bool    `json:"connected"`
	SupportsData         bool    `json:"supportsData"`
	SupportsMultipleData bool    `json:"supportsMultipleData"`
	SupportsInventory    bool    `json:"supportsInventory"`
}

//=============================================================================

type Exchange struct {
	Id         uint   `json:"id"`
	CurrencyId uint   `json:"currencyId"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Timezone   string `json:"timezone"`
}

//=============================================================================
//===
//=== Data product
//===
//=============================================================================

type DataProduct struct {
	Id              uint             `json:"id"`
	ConnectionId    uint             `json:"connectionId"`
	ExchangeId      uint             `json:"exchangeId"`
	Username        string           `json:"username"`
	Symbol          string           `json:"symbol"`
	Name            string           `json:"name"`
	MarketType      string           `json:"marketType"`
	ProductType     string           `json:"productType"`
	Months          string           `json:"months"`
	RolloverTrigger db.DPRollTrigger `json:"rolloverTrigger"`
}

//=============================================================================

type DataProductMessage struct {
	DataProduct DataProduct `json:"dataProduct"`
	Connection  Connection  `json:"connection"`
	Exchange    Exchange    `json:"exchange"`
}

//=============================================================================
//===
//=== Broker product
//===
//=============================================================================

type BrokerProduct struct {
	Id               uint     `json:"id"`
	Username         string   `json:"username"`
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	PointValue       float32  `json:"pointValue"`
	CostPerOperation float32  `json:"costPerOperation"`
}

//=============================================================================

type BrokerProductMessage struct {
	BrokerProduct BrokerProduct `json:"brokerProduct"`
	Connection    Connection    `json:"connection"`
	Exchange      Exchange      `json:"exchange"`
	Currency      Currency      `json:"currency"`
}

//=============================================================================
