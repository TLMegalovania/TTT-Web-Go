package carr

import "sync"

type cItem[v interface{}] struct {
	item   v
	locker *sync.RWMutex
}

type CArray[v interface{}] struct {
	arr []cItem[v]
}

func NewCArray[t interface{}](count int) *CArray[t] {
	res := &CArray[t]{arr: make([]cItem[t], count)}
	for k := range res.arr {
		res.arr[k].locker = &sync.RWMutex{}
	}
	return res
}

func (a *CArray[t]) Get(index int) t {
	c := a.arr[index]
	c.locker.RLock()
	res := c.item
	c.locker.RUnlock()
	return res
}

func (a *CArray[t]) Set(index int, setter func(value *t)) {
	c := a.arr[index]
	c.locker.Lock()
	v := c.item
	setter(&v)
	a.arr[index] = cItem[t]{item: v, locker: c.locker}
	c.locker.Unlock()
}

func (a *CArray[t]) Count() int {
	return len(a.arr)
}

type CTuple[t interface{}] struct {
	Key int
	Val t
}

func (a *CArray[t]) Iter() <-chan CTuple[t] {
	ch := make(chan CTuple[t])
	go func() {
		for i, item := range a.arr {
			item.locker.RLock()
			ch <- CTuple[t]{Key: i, Val: item.item}
			item.locker.RUnlock()
		}
		close(ch)
	}()
	return ch
}
