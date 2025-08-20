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

package jobmanager

import (
	"sync"

	"github.com/bit-fever/data-collector/pkg/db"
)

//=============================================================================

type AdapterCache struct {
	sync.RWMutex
	connections []string
	productsMap map[string]*ProductCache
}

//=============================================================================

func NewAdapterCache() *AdapterCache {
	return &AdapterCache{
		productsMap: make(map[string]*ProductCache),
	}
}

//=============================================================================
//===
//=== API methods
//===
//=============================================================================

func (ac *AdapterCache) getDataBlock(root string, symbol string) *db.DataBlock {
	ac.RLock()
	pc, found := ac.productsMap[root]
	ac.RUnlock()

	if found {
		return pc.getDataBlock(symbol)
	}

	return nil
}

//=============================================================================

func (ac *AdapterCache) addDataBlock(db *db.DataBlock) {
	ac.Lock()

	pc, found := ac.productsMap[db.Root]
	if !found {
		pc = NewProductCache(db.Root)
		ac.productsMap[db.Root] = pc
	}

	ac.Unlock()
	pc.addDataBlock(db)
}

//=============================================================================
