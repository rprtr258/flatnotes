package util

import "bytes"

func bytesHash(b []byte) uint64 {
	var hash uint64 = 5381
	for _, c := range b {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// BytesFilter is a efficient data structure for checking whether bytes exist or not.
// BytesFilter is thread-safe.
type BytesFilter struct {
	chars     [256]uint8
	threshold int
	slots     [][][]byte
}

// NewBytesFilter returns a new BytesFilter.
func NewBytesFilter(elements ...[]byte) *BytesFilter {
	s := &BytesFilter{
		threshold: 3,
		slots:     make([][][]byte, 64),
	}
	for _, element := range elements {
		s.Add(element)
	}
	return s
}

// Add adds given bytes to this set.
func (s *BytesFilter) Add(b []byte) {
	l := len(b)
	m := s.threshold
	if l < s.threshold {
		m = l
	}
	for i := 0; i < m; i++ {
		s.chars[b[i]] |= 1 << uint8(i)
	}
	h := bytesHash(b) % uint64(len(s.slots))
	slot := s.slots[h]
	if slot == nil {
		slot = [][]byte{}
	}
	s.slots[h] = append(slot, b)
}

// Extend copies this filter and adds given bytes to new filter.
func (s *BytesFilter) Extend(bs ...[]byte) *BytesFilter {
	newFilter := NewBytesFilter()
	newFilter.chars = s.chars
	newFilter.threshold = s.threshold
	for k, v := range s.slots {
		newSlot := make([][]byte, len(v))
		copy(newSlot, v)
		newFilter.slots[k] = v
	}
	for _, b := range bs {
		newFilter.Add(b)
	}
	return newFilter
}

// Contains return true if this set contains given bytes, otherwise false.
func (s *BytesFilter) Contains(b []byte) bool {
	l := len(b)
	m := s.threshold
	if l < s.threshold {
		m = l
	}
	for i := 0; i < m; i++ {
		if (s.chars[b[i]] & (1 << uint8(i))) == 0 {
			return false
		}
	}
	h := bytesHash(b) % uint64(len(s.slots))
	slot := s.slots[h]
	if len(slot) == 0 {
		return false
	}
	for _, element := range slot {
		if bytes.Equal(element, b) {
			return true
		}
	}
	return false
}
