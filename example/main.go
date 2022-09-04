package main

import (
	"machine"
	"time"

	trig "github.com/eyelight/trigger"
)

var (
	btn        = machine.D3
	dispatchCh = make(chan trig.Trigger, 1)
	fakeMqttCh = make(chan trig.Trigger, 1)
)

func main() {
	time.Sleep(3 * time.Second)
	light1 := NewPeripheral("LightBulb1")
	light2 := NewPeripheral("LightBulb2")
	dispatch := trig.NewDispatch(dispatchCh)
	dispatch.AddToDispatch(&light1, &light2)
	misoMqtt := NewFakeMqtt(fakeMqttCh)
	go dispatch.Dispatch()
	go misoMqtt.ListenAndSend()

	triggerFromMqttA := trig.Trigger{
		Target:   "LightBulb1",
		Action:   "FakeOn",
		Duration: time.Duration(5 * time.Second),
		Report:   true,
		ReportCh: fakeMqttCh,
	}
	triggerFromMqttB := trig.Trigger{
		Target:   "LightBulb2",
		Action:   "FakeOff",
		Duration: time.Duration(0),
		Report:   true,
		ReportCh: fakeMqttCh,
	}
	triggerFromMqttC := trig.Trigger{
		Target:   "LightBulb3",
		Action:   "FakeToggle",
		Duration: time.Duration(0),
		Report:   true,
		ReportCh: fakeMqttCh,
	}
	triggerFromMqttD := trig.Trigger{
		Target:   "LightBulb1",
		Action:   "FakeOn",
		Duration: time.Duration(0),
		Report:   true,
		ReportCh: fakeMqttCh,
	}

	dispatchCh <- triggerFromMqttA
	time.Sleep(2 * time.Second)
	dispatchCh <- triggerFromMqttB
	time.Sleep(1 * time.Second)
	dispatchCh <- triggerFromMqttC
	time.Sleep(200 * time.Millisecond)
	dispatchCh <- triggerFromMqttD
	select {}
}

type responder struct {
	name string
}

func NewPeripheral(n string) responder {
	return responder{
		name: n,
	}
}

func (r *responder) Name() string {
	return r.name
}

// func (r *responder) Trigger() chan trig.Trigger {
// 	return r.ch
// }

func (r *responder) Execute(t trig.Trigger) {
	if t.Target != r.name {
		t.Error, t.Report = true, false
		t.Message = string("error - " + r.name + " received a trigger intended for " + t.Target)
		t.ReportCh <- t
		return
	}
	switch t.Action {
	case "FakeOn":
		t.Error, t.Report = false, false
		t.Message = string(r.name + " executing FakeOn at " + time.Now().String() + " for duration " + t.Duration.String())
		if t.Duration > 0 {
			go func() {
				println("anon")
				time.Sleep(t.Duration)
				t.Message = string(r.name + " executed FakeOff after " + t.Duration.String())
				t.ReportCh <- t
			}()
		}
		t.ReportCh <- t
		println(r.name + " doing FakeOn")
	case "FakeOff":
		t.Error, t.Report = false, false
		t.Message = string(r.name + " executing FakeOff at " + time.Now().String())
		t.ReportCh <- t
		println(r.name + " doing FakeOff")
	case "FakeToggle":
		t.Error, t.Report = false, false
		t.Message = string(r.name + " executing FakeToggle at " + time.Now().String())
		t.ReportCh <- t
		println(r.name + " doing FakeToggle")
	}
}

type fakeMqtt struct {
	input chan trig.Trigger
}

func NewFakeMqtt(ch chan trig.Trigger) fakeMqtt {
	return fakeMqtt{
		input: ch,
	}
}

func (m *fakeMqtt) ListenAndSend() {
	for {
		select {
		case t := <-m.input:
			println("	Mqtt MISO Reponse: " + t.Message)
		}
	}
}
