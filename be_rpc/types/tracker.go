package types

import (
	"golang.org/x/exp/constraints"
	"sort"
)

type Tracker[T constraints.Ordered] map[T]bool

func (m Tracker[T]) Add(value T) {
	m[value] = true
}

func (m Tracker[T]) Has(value T) bool {
	_, found := m[value]
	return found
}

func (m Tracker[T]) ToSlice() []T {
	res := make([]T, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

func (m Tracker[T]) ToSortedSlice() []T {
	slice := m.ToSlice()
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	return slice
}
