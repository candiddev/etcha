package commands

import (
	"sort"

	"github.com/candiddev/shared/go/cli"
)

// Output is the output from running commands.
type Output struct {
	Change          cli.CmdOutput
	Changed         bool
	ChangeFail      bool
	Check           cli.CmdOutput
	Checked         bool
	CheckFailChange bool
	CheckFailRemove bool
	Events          []string
	ID              string
	ParentID        string
	Remove          cli.CmdOutput
	Removed         bool
	RemoveFail      bool
}

// Outputs is a list of Command IDs and the associated outputs from a run.
type Outputs []*Output

// Event is a list of IDs and a name.
type Event struct {
	Name    string
	Outputs []Output
}

// Events is multiple events.
type Events []Event

// CheckFail returns a list of IDs that have failed check.
func (o Outputs) CheckFail(remove bool) []string {
	var out []string

	for _, u := range o {
		if (remove && u.CheckFailRemove) || (!remove && u.CheckFailChange) {
			out = append(out, u.ID)
		}
	}

	return out
}

// Changed returns a list of IDs that have Changed.
func (o Outputs) Changed() (ids, outputs []string) {
	for _, u := range o {
		if u.Changed {
			ids = append(ids, u.ID)
			outputs = append(outputs, u.Change.String())
		}
	}

	return ids, outputs
}

// Events returns a sorted list of events that were fired.
func (o Outputs) Events() Events {
	em := map[string][]Output{}

	for _, u := range o {
		if u != nil {
			for _, e := range u.Events {
				em[e] = append(em[e], *u)
			}
		}
	}

	names := []string{}

	for k := range em {
		names = append(names, k)
	}

	sort.Strings(names)

	ev := []Event{}

	for _, s := range names {
		ev = append(ev, Event{
			Name:    s,
			Outputs: em[s],
		})
	}

	return ev
}

// Failed returns a list of IDs that have Failed.
func (o Outputs) Failed() []string {
	var out []string

	for _, u := range o {
		if u.ChangeFail || u.RemoveFail {
			out = append(out, u.ID)
		}
	}

	return out
}

// Removed returns a list of IDs and their outputs that have Removed.
func (o Outputs) Removed() (ids, outputs []string) {
	for _, u := range o {
		if u.Removed {
			ids = append(ids, u.ID)
			outputs = append(outputs, u.Remove.String())
		}
	}

	return ids, outputs
}
