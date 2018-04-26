//    Copyright 2018 Google LLC
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package server

import (
	"container/heap"

	"github.com/danchia/ddb/sst"
)

// mergingIter is an iterator that merges from many sst.Iter.
type mergingIter struct {
	h *iterHeap

	curKey   string
	curTs    int64
	curValue []byte
}

func newMergingIter(iters []*sst.Iter) (*mergingIter, error) {
	mi := &mergingIter{}
	for _, iter := range iters {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if hasNext {
			*mi.h = append(*mi.h, iter)
		}
	}
	heap.Init(mi.h)
	return mi, nil
}

// Next advances the iterator. Returns true if there is a next value.
func (i *mergingIter) Next() (bool, error) {
	iter := heap.Pop(i.h).(*sst.Iter)
	i.curKey = iter.Key()
	i.curTs = iter.Timestamp()
	i.curValue = iter.Value()

	hasNext, err := iter.Next()
	if err != nil {
		return false, err
	}

	if hasNext {
		heap.Push(i.h, iter)
	}

	return i.h.Len() > 0, nil
}

type iterHeap []*sst.Iter

func (h iterHeap) Len() int { return len(h) }

func (h iterHeap) Less(i, j int) bool {
	it1 := h[i]
	it2 := h[j]

	if it1.Key() != it2.Key() {
		return it1.Key() < it2.Key()
	}

	return it1.Timestamp() < it2.Timestamp()
}

func (h iterHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *iterHeap) Push(x interface{}) {
	*h = append(*h, x.(*sst.Iter))
}

func (h *iterHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
