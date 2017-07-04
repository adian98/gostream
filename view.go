package gocep

import "sort"

type View interface {
	Apply(event []Event) []Event
}

type First struct{}

func (f First) Apply(event []Event) (stream []Event) {
	if len(event) == 0 {
		return stream
	}
	return append(stream, event[0])
}

type Last struct{}

func (f Last) Apply(event []Event) (stream []Event) {
	if len(event) == 0 {
		return stream
	}
	return append(stream, event[len(event)-1])
}

type Limit struct {
	Offset int
	Limit  int
}

func (f Limit) Apply(event []Event) []Event {
	if len(event) < f.Offset+f.Limit {
		return event
	}
	return event[f.Offset : f.Offset+f.Limit]
}

type SortableInt struct {
	event []Event
	name  string
}

func (s SortableInt) Len() int {
	return len(s.event)
}

func (s SortableInt) Less(i, j int) bool {
	return s.event[i].IntValue(s.name) < s.event[j].IntValue(s.name)
}

func (s SortableInt) Swap(i, j int) {
	s.event[i], s.event[j] = s.event[j], s.event[i]
}

type SortInt struct {
	Name    string
	Reverse bool
}

func (f SortInt) Apply(event []Event) []Event {
	data := SortableInt{event, f.Name}
	if f.Reverse {
		sort.Sort(sort.Reverse(data))
		return data.event
	}
	sort.Sort(data)
	return data.event
}

type SortableFloat struct {
	event []Event
	name  string
}

func (s SortableFloat) Len() int {
	return len(s.event)
}

func (s SortableFloat) Less(i, j int) bool {
	return s.event[i].Float32Value(s.name) < s.event[j].Float32Value(s.name)
}

func (s SortableFloat) Swap(i, j int) {
	s.event[i], s.event[j] = s.event[j], s.event[i]
}

type SortFloat struct {
	Name    string
	Reverse bool
}

func (f SortFloat) Apply(event []Event) []Event {
	data := SortableFloat{event, f.Name}
	if f.Reverse {
		sort.Sort(sort.Reverse(data))
		return data.event
	}
	sort.Sort(data)
	return data.event
}

type HavingLargerThanInt struct {
	Name  string
	Value int
}

func (f HavingLargerThanInt) Apply(event []Event) []Event {
	if event[len(event)-1].Record[f.Name].(int) > f.Value {
		return event
	}
	return []Event{}
}

type HavingLargerThanFloat struct {
	Name  string
	Value float32
}

func (f HavingLargerThanFloat) Apply(event []Event) []Event {
	if event[len(event)-1].Record[f.Name].(float32) > f.Value {
		return event
	}
	return []Event{}
}

type HavingLessThanInt struct {
	Name  string
	Value int
}

func (f HavingLessThanInt) Apply(event []Event) []Event {
	if event[len(event)-1].Record[f.Name].(int) < f.Value {
		return event
	}
	return []Event{}
}

type HavingLessThanFloat struct {
	Name  string
	Value float32
}

func (f HavingLessThanFloat) Apply(event []Event) []Event {
	if event[len(event)-1].Record[f.Name].(float32) < f.Value {
		return event
	}
	return []Event{}
}