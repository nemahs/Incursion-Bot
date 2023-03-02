package Utils

type QueueDataType struct {
	Distance int
	SystemID int
}

type Queue struct {
	list []QueueDataType
}

func (q *Queue) Add(val QueueDataType) {
	if q.Contains(val) {
		return
	}

	for i, v := range q.list {
		if q.less(val, v) {
			q.list = append(q.list[:i+1], q.list[i:]...)
			q.list[i] = val
			return
		}
	}

	q.list = append(q.list, val)
}

func (q *Queue) IsEmpty() bool {
	return len(q.list) == 0
}

func (q *Queue) less(l QueueDataType, r QueueDataType) bool {
	if l.Distance < r.Distance {
		return true
	}

	if l.Distance > r.Distance {
		return false
	}

	return l.SystemID > r.SystemID
}

func (q *Queue) Pop() QueueDataType {
	ret := q.list[0]
	q.list = q.list[1:]
	return ret
}

func (q *Queue) Remove(val int) {
	var result []QueueDataType
	for _, v := range q.list {
		if v.SystemID != val {
			result = append(result, v)
		}
	}

	q.list = result
}

func (q *Queue) Contains(val QueueDataType) bool {
	for _, v := range q.list {
		if v.SystemID == val.SystemID {
			return true
		}
	}

	return false
}
