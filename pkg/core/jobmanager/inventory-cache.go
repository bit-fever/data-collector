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

type InventoryCache struct {
	sync.RWMutex
	adaptersMap map[string]*AdapterCache
}

//=============================================================================

func NewInventoryCache() *InventoryCache {
	ic := InventoryCache{
		adaptersMap: make(map[string]*AdapterCache),
	}

	return &ic
}

//=============================================================================
//===
//=== API methods
//===
//=============================================================================

func (ic *InventoryCache) getDataBlock(systemCode, root, symbol string) *db.DataBlock {
	ic.RLock()
	ac, found := ic.adaptersMap[systemCode]
	ic.RUnlock()

	if found {
		return ac.getDataBlock(root, symbol)
	}

	return nil
}

//=============================================================================

func (ic *InventoryCache) addDataBlock(db *db.DataBlock) {
	ic.Lock()

	ac, found := ic.adaptersMap[db.SystemCode]
	if !found {
		ac = NewAdapterCache()
		ic.adaptersMap[db.SystemCode] = ac
	}

	ic.Unlock()
	ac.addDataBlock(db)
}

//=============================================================================
