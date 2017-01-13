package queue

import (
	"runtime"
	"sync/atomic"
)

type Memory struct { //数组解构
	value interface{} //数据值
	m     bool        // mark位
}

func NewQueue(max uint64) *Queue {
	q := new(Queue)
	q.maxlen = minLen(max)
	q.capM = q.maxlen - 1
	q.mp = make([]Memory, q.maxlen)
	return q
}

// lock free queue
type Queue struct {
	maxlen uint64 //最大长度
	capM   uint64
	putB   uint64   //生产位
	getB   uint64   //消费位
	mp     []Memory //数组
}

func (q *Queue) Maxlen() uint64 {
	return q.maxlen
}

func (q *Queue) Len() uint64 {
	var putB, getB uint64
	var max uint64
	getB = q.getB
	putB = q.putB

	if putB >= getB {
		max = putB - getB
	} else {
		max = q.capM + putB - getB
	}

	return max
}

// put queue functions
func (q *Queue) Put(val interface{}) (ok bool, cnt uint64) {
	var putB, putBNew, getB, posCnt uint64
	var men *Memory
	capM := q.capM
	for {
		getB = q.getB
		putB = q.putB

		if putB >= getB {
			posCnt = putB - getB
		} else {
			posCnt = capM + putB - getB
		}

		if posCnt >= capM {
			runtime.Gosched()
			return false, posCnt
		}

		putBNew = putB + 1
		if atomic.CompareAndSwapUint64(&q.putB, putB, putBNew) {
			break
		} else {
			runtime.Gosched()
		}
	}

	men = &q.mp[putBNew&capM]

	for {
		if !men.m {
			men.value = val
			men.m = true
			return true, posCnt + 1
		} else {
			runtime.Gosched()
		}
	}
}

// get queue functions
func (q *Queue) Get() (val interface{}, ok bool, cnt uint64) {
	var putB, getB, getBNew, posCnt uint64
	var men *Memory
	capM := q.capM
	for {
		putB = q.putB
		getB = q.getB

		if putB >= getB {
			posCnt = putB - getB
		} else {
			posCnt = capM + putB - getB
		}

		if posCnt < 1 {
			runtime.Gosched()
			return nil, false, posCnt
		}

		getBNew = getB + 1
		if atomic.CompareAndSwapUint64(&q.getB, getB, getBNew) {
			break
		} else {
			runtime.Gosched()
		}
	}

	men = &q.mp[getBNew&capM]

	for {
		if men.m {
			val = men.value
			men.m = false
			return val, true, posCnt - 1
		} else {
			runtime.Gosched()
		}
	}
}

// round 到最近的2的倍数
func minLen(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
