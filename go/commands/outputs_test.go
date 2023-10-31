package commands

import (
	"testing"

	"github.com/candiddev/shared/go/assert"
)

func TestOutputs(t *testing.T) {
	out := Outputs{
		&Output{
			Change:     "change",
			Changed:    true,
			ChangeFail: false,
			Check:      "check",
			CheckFail:  true,
			Events:     []string{"event2", "event1"},
			ID:         "a",
			Remove:     "remove",
			Removed:    true,
		},
		&Output{
			Change:     "change",
			ChangeFail: true,
			Check:      "check",
			CheckFail:  true,
			ID:         "b",
			Remove:     "remove",
			Removed:    true,
		},
	}

	ids, outputs := out.Changed()

	assert.Equal(t, ids, []string{"a"})
	assert.Equal(t, outputs, []string{"change"})
	assert.Equal(t, out.Events(), Events{
		Event{
			Name:    "event1",
			Outputs: []Output{*out[0]},
		},
		Event{
			Name:    "event2",
			Outputs: []Output{*out[0]},
		},
	})
	assert.Equal(t, out.Failed(), []string{"b"})

	ids, outputs = out.Removed()

	assert.Equal(t, ids, []string{"a", "b"})
	assert.Equal(t, outputs, []string{"remove", "remove"})
}
