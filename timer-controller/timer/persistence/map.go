package persistence

import (
	"sync"
	"time"
)

type HashMap struct {
	// map[ExpireTime]msg
	// 纳秒级别，备份一条就好,而且delete是通过时间
	storeMedium map[time.Time][]byte

	rwMutex sync.RWMutex
}

func CreateHashMap() (hashmap HashMap) {
	hashmap = HashMap{
		storeMedium: make(map[time.Time][]byte),
	}

	return
}

func NewHashMap() (hashmap *HashMap) {
	m := CreateHashMap()
	return &m
}

func (this *HashMap) Set(key time.Time, val []byte) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	this.storeMedium[key] = val
}

// 获取少于当前时间的所有数据
func (this *HashMap) Get(now time.Time) (datas [][]byte) {
	// TODO
	// 改用list
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()

	datas = make([][]byte, 0)
	for t, tdata := range this.storeMedium {
		if t.Before(now) {
			datas = append(datas, tdata)
		}
	}

	return
}

// 删除时间范围内的所有数据, 范围: [start1, end1), [start2, end2) ...
func (this *HashMap) Delete(start time.Time, end time.Time, pairs ...time.Time) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	mlist := make(DeleteTimeList, 0)
	mlist = append(mlist, start, end)

	if len(pairs)%2 == 0 {
		mlist = append(mlist, pairs...)
	}

	deletetimes := make(DeleteTimeList, 0)

	for t, _ := range this.storeMedium {
		for i := 0; i < len(mlist); i++ {
			start := mlist[i]
			i++
			end := mlist[i]

			if t.After(start) && t.Before(end) {
				deletetimes = append(deletetimes, t)
			}
		}
	}

	for _, t := range deletetimes {
		delete(this.storeMedium, t)
	}
}
