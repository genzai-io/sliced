package ring

import (
	"bytes"
	"fmt"

	"github.com/slice-d/genzai/proto/store"
)

const (
	Slots = 16384
)

func SlotRangeString(s *store.SlotRange) string {
	return fmt.Sprintf("[%d] %d -> %d", s.Slice, s.Low, s.High)
}

//
func RebalanceString(r *store.Rebalance) string {
	buf := &bytes.Buffer{}
	for _, s := range r.Tasks {
		buf.Write([]byte(TaskToString(s)))
		buf.WriteString("\n")
	}
	return buf.String()
}

func TaskHigh(m *store.Rebalance_Task) int32 {
	return m.Low + m.Count
}

func TaskToString(m *store.Rebalance_Task) string {
	return fmt.Sprintf("%d[%d, %d] -> %d", m.From, m.Low, m.Low+m.Count, m.To)
}

type Ring struct {
	Slots  [Slots]int32
	Ranges []*store.SlotRange
}

// Creates a new Ring that evenly distributes the slots.
func Balanced(numOfSlices int32) *Ring {
	if numOfSlices < 1 {
		numOfSlices = 1
	}
	if numOfSlices > Slots {
		numOfSlices = Slots
	}
	slices := make([]*store.SlotRange, 0)
	perSlice := Slots / numOfSlices

	for i := int32(0); i < numOfSlices; i++ {
		s := &store.SlotRange{
			Slice: i,
			Low:   perSlice * i,
			High:  perSlice*i + perSlice,
		}

		if i == numOfSlices-1 {
			s.High = Slots
		}

		slices = append(slices, s)
	}

	return newRing(slices)
}

// Constructs a new Ring from Ranges
func newRing(slices []*store.SlotRange) *Ring {
	r := &Ring{}

	r.Ranges = slices

	// Fill ring
	for _, s := range r.Ranges {
		for slot := s.Low; slot < s.High; slot++ {
			r.Slots[slot] = s.Slice
		}
	}

	return r
}

func (r *Ring) Slice(key []byte) int32 {
	return r.Slots[int(CRC16(key))]
}

// Creates a migration plan to transform from one ring to another.
func (from *Ring) Migrate(to *Ring) (*store.Rebalance, error) {
	changes := make([]*store.Rebalance_Task, 0)
	var change *store.Rebalance_Task

	for slot := int32(0); slot < Slots; slot++ {
		fromSlice := from.Slots[slot]
		toSlice := to.Slots[slot]

		if fromSlice == toSlice {
			continue
		}

		if change != nil {
			if change.From == fromSlice && change.To == toSlice {
				if change.Low+change.Count == slot {
					change.Count++
				} else {
					changes = append(changes, change)
					change = &store.Rebalance_Task{
						From:  fromSlice,
						To:    toSlice,
						Low:   slot,
						Count: 1,
					}
				}
			} else {
				changes = append(changes, change)
				change = &store.Rebalance_Task{
					From:  fromSlice,
					To:    toSlice,
					Low:   slot,
					Count: 1,
				}
			}
		} else {
			change = &store.Rebalance_Task{
				From:  fromSlice,
				To:    toSlice,
				Low:   slot,
				Count: 1,
			}
		}
	}

	if change != nil {
		changes = append(changes, change)
	}

	return &store.Rebalance{
		Tasks: changes,
	}, nil
}

// Make a pretty string
func (r *Ring) String() string {
	buf := &bytes.Buffer{}
	for _, s := range r.Ranges {
		buf.Write([]byte(SlotRangeString(s)))
		buf.WriteString("\n")
	}
	return buf.String()
}
