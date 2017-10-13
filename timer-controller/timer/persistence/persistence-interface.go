package persistence

import (
	"time"
)

type Persistence interface {
	// 存储数据
	Set(key time.Time, data []byte)

	// 获取少于当前时间的所有数据
	Get(now time.Time) [][]byte

	// 删除时间范围内的所有数据, 范围: [start1, end1), [start2, end2) ...
	Delete(start time.Time, end time.Time, pairs ...time.Time)
}

type DeleteTimeList []time.Time

func (this DeleteTimeList) Len() int {
	return len(this)
}

func (this DeleteTimeList) Swap(i int, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this DeleteTimeList) Less(i int, j int) bool {
	return this[i].Before(this[j])
}
