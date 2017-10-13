package persistence

import (
	"encoding/json"
	"sync"
	"time"

	proto "subassembly/timer-controller/proto/notify"
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

func (this *HashMap) Listen(datasrc <-chan []byte) (err error) {
	go this.listen(datasrc)
	return
}

func (this *HashMap) listen(datasrc <-chan []byte) {
	var err error
	for {
		select {
		case data, ok := <-datasrc:
			if !ok {
				continue
			}
			err = this.Set(data)
			if err != nil {
				// TODO
				// log.Notice()
				continue
			}
		}

	}
}

func (this *HashMap) Set(data []byte) (err error) {
	val := &proto.TimerNotice{}

	err = json.Unmarshal(data, val)
	if err != nil {
		return
	}

	expireTime := val.SendUnixTime.Add(val.Expire)
	this.set(expireTime, data)

	return
}

func (this *HashMap) set(key time.Time, val []byte) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	vallist, ok := this.storeMedium[key]
	if !ok {
		vallist = make([][]byte, 0)
	}

	vallist = append(vallist, val)
	this.storeMedium[key] = vallist
}

// 获取少于当前时间的所有数据
func (this *HashMap) Get(now time.Time) (datas [][]byte) {
	// TODO
	// 改用list
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()

	datas = make([][]byte, 0)
	for t, tdatas := range this.storeMedium {
		if t.Before(now) {
			datas = append(datas, tdatas...)
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
	mlist = append(mlist, pairs...)

	// TODO

}
