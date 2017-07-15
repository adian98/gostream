package gocep

import (
	"testing"
)

func TestParser(t *testing.T) {
	q := "select * from MapEvent.length(10)"

	stmt, err := NewParser(q).Parse()
	if err != nil {
		t.Error(err)
		return
	}
	window := stmt.Build(1024)

	m := make(map[string]interface{})
	m["Value"] = "foobar"

	window.Input() <- MapEvent{m}
	event := <-window.Output()
	if event[0].RecordString("Value") != "foobar" {
		t.Error(event)
	}
}
