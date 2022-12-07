package trigger

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	ERR_TARGET_NOT_FOUND = "Target unknown to the Dispatcher"
)

type Trigger struct {
	Target   string        // a valid Triggerable.Name()
	Action   string        // any action understandable by the Triggerable
	Duration time.Duration // a duration meaningful to the Triggerable
	Message  string        // an optional message; possibly an error or a status update
	ReportCh chan Trigger  // the channel on which the Triggerable should report
	Error    bool          // whether this trigger is an error
}

func (t Trigger) String() string {
	ss := strings.Builder{}
	ss.Grow(512)
	ss.WriteString("Trigger\n")
	ss.WriteString("\tTarget: ")
	ss.WriteString(t.Target)
	ss.WriteString("\nAction: ")
	ss.WriteString(t.Action)
	ss.WriteString("\nDuration: ")
	ss.WriteString(t.Duration.String())
	ss.WriteString("\nMessage: ")
	ss.WriteString(t.Message)
	ss.WriteString("\nError: ")
	ss.WriteString(strconv.FormatBool(t.Error))
	return ss.String()
}

type dispatch struct {
	triggerCh    chan Trigger  // the channel on which the Dispatcher will receive Triggers
	triggerables []Triggerable // a slice of Triggerables addressable by the Dispatcher
}

type Dispatcher interface {
	AddToDispatch(t ...Triggerable) // pass any Triggerable who you want to be addressable by this Dispatcher
	Dispatch()                      // Dispatcher will consume (dispatch).triggerCh
}

type Triggerable interface {
	Name() string      // Dispatcher will use this to match incoming Triggers to intended receivers
	Execute(t Trigger) // Dispatcher will pass a Trigger for the Triggerable to execute
	// Trigger() chan Trigger // return the channel on which you're listening for Triggers

}

// NewDispatch returns a Dispatcher listening for Triggers on a passed-in channel
func NewDispatch(triggerCh chan Trigger) Dispatcher {
	return &dispatch{
		triggerCh:    triggerCh,
		triggerables: make([]Triggerable, 0),
	}
}

// AddToDispatch makes a Dispatcher aware of a slice of Triggerables
func (d *dispatch) AddToDispatch(t ...Triggerable) {
	if len(t) > 0 {
		println("Adding triggerables to dispatch")
		d.triggerables = append(d.triggerables, t...)
	}
}

// Dispatch is a goroutine accepting Triggers on a channel,
// matches the received Trigger.Target to a Triggerable known to the Dispatcher,
// and concurrently calls the Triggerable to Execute(Trigger)
func (d *dispatch) Dispatch() {
	fmt.Println("Dispatching – " + strconv.FormatInt(int64(len(d.triggerables)), 10) + " Valid Targets")
	for i := range d.triggerables {
		fmt.Println("			" + d.triggerables[i].Name())
	}
	for {
		select {
		case t := <-d.triggerCh:
			if t.Target == "?" {
				ss := strings.Builder{}
				ss.Grow(512)
				ss.WriteString("Valid targets: ")
				for _, n := range d.triggerables {
					ss.WriteString(n.Name())
					ss.WriteString(", ")
				}
				t.Message = ss.String()
				fmt.Println(t.Message)
				t.ReportCh <- t
				continue
			}
			r, err := d.findTarget(t)
			if err != nil {
				fmt.Println("dispatcher error - " + err.Error())
				t.Target = "MISO"
				t.Action = "ErrorReport"
				t.Message = err.Error()
				t.ReportCh <- t
				continue
			}
			go func() {
				fmt.Println("executing on " + r.Name() + " from trigger " + t.Target)
				r.Execute(t)
			}()
		}
	}
}

func (d *dispatch) findTarget(t Trigger) (Triggerable, error) {
	for i := range d.triggerables {
		if t.Target == d.triggerables[i].Name() {
			return d.triggerables[i], nil
		}
	}
	ss := strings.Builder{}
	ss.Grow(len(d.triggerables) * 16)
	ss.WriteString("named: ")
	for i := range d.triggerables {
		ss.WriteString(d.triggerables[i].Name())
		ss.WriteString(", ")
	}
	return nil, errors.New(string(t.Target + " " + ERR_TARGET_NOT_FOUND + " (" + strconv.FormatInt(int64(len(d.triggerables)), 10) + " known Triggerables " + ss.String() + ") "))
}
