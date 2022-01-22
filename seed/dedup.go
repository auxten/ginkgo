package seed

import (
	"bytes"
	"sort"
)

func SortDeDup(l []Host) []Host {
	n := len(l)
	if n <= 1 {
		return l
	}
	sort.Slice(l, func(i, j int) bool {
		c := bytes.Compare(l[i].IP[:], l[j].IP[:])
		if c != 0 {
			return c < 0
		} else {
			return l[i].Port < l[j].Port
		}
	})

	j := 1
	for i := 1; i < n; i++ {
		if l[i] != l[i-1] {
			l[j] = l[i]
			j++
		}
	}

	return l[0:j]
}
