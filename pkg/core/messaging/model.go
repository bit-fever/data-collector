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

//=============================================================================

type ProductData struct {
	Id            uint    `json:"id"`
	ConnectionId  uint    `json:"connectionId"`
	ExchangeId    uint    `json:"exchangeId"`
	Username      string  `json:"username"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Increment     float64 `json:"increment"`
}

//=============================================================================

type Connection struct {
	Id                   uint    `json:"id"`
	Username             string  `json:"username"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	SystemCode           string  `json:"systemCode"`
	SystemName           string  `json:"systemName"`
	InstanceCode         string  `json:"instanceCode"`
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

type ProductDataMessage struct {
	ProductData ProductData `json:"productData"`
	Connection  Connection  `json:"connection"`
	Exchange    Exchange    `json:"exchange"`
}

//=============================================================================
