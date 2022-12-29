# Trigger
Trigger is a package useful for relaying structured messages (Triggers) between peripherals in a microcontroller + mqtt setting. 

There is also a branch 'nonTinyGo' which is useful when building emulators of a device that run in mainline Go, which uses the `fmt` package. 

Trigger provides a Trigger struct and interfaces Triggerable and Dispatcher. 

### Trigger
Triggers pass an Action to an intended Target, along with an optional Duration & Message.
```go
type Trigger struct {
    Target string
    Action string
    Duration time.Duration
    Message string
    ReportCh chan Trigger
    Error bool
}
```

### Dispatcher
A Dispatcher listens for incoming Triggers on a passed-in channel and passes Triggers to the intended Triggerable from the Trigger.Target. Dispatcher implements the following methods:
```go
type Dispatcher interface {
    AddToDispatch(t ...Triggerable)
    Dispatch()
}
```
where `Dispatch()` is intended to be a goroutine.


### Triggerable
Triggerables must implement the following methods in order to gain the ability to be triggered by Triggers.
```go
type Triggerable interface {
    Name() string
    Execute(t Trigger)
}
```
