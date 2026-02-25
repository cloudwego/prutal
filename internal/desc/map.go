/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package desc

import (
	"sync/atomic"
)

// mapStructDesc represents a read-lock-free hashmap for *StructDesc like sync.Map.
// it's NOT designed for writes.
// Each slot is an atomic.Pointer so writes only reallocate the affected slot.
type mapStructDesc struct {
	buckets [mapStructDescBuckets + 1]atomic.Pointer[[]mapStructDescItem]
}

// XXX: fixed size to make it simple,
// we may not so many structs that need to rehash it
const mapStructDescBuckets = 0xffff

type mapStructDescItem struct {
	abiType uintptr
	sd      *StructDesc
}

// Get ...
func (m *mapStructDesc) Get(abiType uintptr) *StructDesc {
	p := m.buckets[abiType&mapStructDescBuckets].Load()
	if p == nil {
		return nil
	}
	dd := *p
	for i := range dd {
		if dd[i].abiType == abiType {
			return dd[i].sd
		}
	}
	return nil
}

// Set ...
// It's slow, should be used in rare write cases
func (m *mapStructDesc) Set(abiType uintptr, sd *StructDesc) {
	if m.Get(abiType) == sd {
		return
	}
	bk := abiType & mapStructDescBuckets
	var old []mapStructDescItem
	if p := m.buckets[bk].Load(); p != nil {
		old = *p
	}
	// alloc len+1 cap upfront since append is the common path
	ns := make([]mapStructDescItem, len(old), len(old)+1)
	copy(ns, old)
	for i := range ns {
		if ns[i].abiType == abiType {
			ns[i].sd = sd
			m.buckets[bk].Store(&ns)
			return
		}
	}
	ns = append(ns, mapStructDescItem{abiType: abiType, sd: sd})
	m.buckets[bk].Store(&ns)
}
