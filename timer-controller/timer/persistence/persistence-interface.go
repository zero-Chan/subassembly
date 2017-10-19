package persistence

import (
	"sort"
	"time"
)

type Persistence interface {
	Close() error

	// 存储数据
	Set(key time.Time, data []byte) error

	// 获取少于当前时间的所有数据
	Get(now time.Time) (datas [][]byte, err error)

	// 删除时间范围内的所有数据, 范围: [start1, end1), [start2, end2) ...
	Delete(start time.Time, end time.Time, pairs ...time.Time) (err error)
}

type DeleteTimeList []time.Time

func NewDeleteTimeList(start time.Time, end time.Time, others ...time.Time) (list DeleteTimeList) {
	list = make(DeleteTimeList, 0)
	list = append(list, start, end)

	// if have other timer, we should have double data to control
	// e,g:
	// [start, end1), [end1+1, start2), [start2+1, end)...
	if others != nil && len(others) != 0 {
		list = append(list, list...)
		list = append(list, others...)
		list = append(list, others...)
	}

	sort.Sort(list)

	// the first enum and the last enum have only one.
	if len(list) > 2 {
		list = list[1 : len(list)-1]
	}

	return list
}

func (this DeleteTimeList) Len() int {
	return len(this)
}

func (this DeleteTimeList) Swap(i int, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this DeleteTimeList) Less(i int, j int) bool {
	return this[i].Before(this[j])
}
