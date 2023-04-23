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

	assert.Equal(t, out.Changed(), []string{"a"})
	assert.Equal(t, out.Events(), Events{
		&Event{
			Name:    "event1",
			Outputs: Outputs{out[0]},
		},
		&Event{
			Name:    "event2",
			Outputs: Outputs{out[0]},
		},
	})
	assert.Equal(t, out.Failed(), []string{"b"})
	assert.Equal(t, out.Removed(), []string{"a", "b"})
}
