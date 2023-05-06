package incursions

type IncursionList []Incursion

func (list *IncursionList) Find(inc Incursion) *Incursion {
	for i, incursion := range *list {
		if incursion.Equal(inc) {
			return &(*list)[i]
		}
	}
	return nil
}

func (list *IncursionList) Empty() bool { return len(*list) == 0 }

func (list *IncursionList) Remove(i int) {
	*list = append((*list)[:i], (*list)[i+1:]...)
}

func (list *IncursionList) RemoveFunc(fun func(inc Incursion) bool) {
	for i, inc := range *list {
		if fun(inc) {
			list.Remove(i)
			return
		}
	}
}
