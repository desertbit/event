// Package event is a simple event emitter for Go.
package event

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// ErrNotFunc presented when an invalid argument is provided as a listener function.
var ErrNotFunc = errors.New("listener value is not a function")

// Recoverer for panics. The interface defines the listener function.
type Recoverer func(interface{}, error)

// Event implementation.
type Event struct {
	Recoverer Recoverer

	mutex sync.Mutex
	funcs map[reflect.Value]bool // The boolean value defines if the event should be called only once.
}

// New creates a new single event.
// Optionally pass a recoverer.
func New(r ...Recoverer) *Event {
	e := &Event{
		funcs: make(map[reflect.Value]bool),
	}

	if len(r) > 0 {
		e.Recoverer = r[0]
	}

	return e
}

// On subscribes a function as new event listener.
func (e *Event) On(listener interface{}) {
	e.addListener(listener, false)
}

// Once subscribes a function as new event listener.
// This function will be called only once.
func (e *Event) Once(listener interface{}) {
	e.addListener(listener, true)
}

func (e *Event) addListener(listener interface{}, once bool) {
	fn := reflect.ValueOf(listener)
	if fn.Kind() != reflect.Func {
		if e.Recoverer != nil {
			e.Recoverer(listener, ErrNotFunc)
		} else {
			panic(ErrNotFunc)
		}
	}

	e.mutex.Lock()
	e.funcs[fn] = once
	e.mutex.Unlock()
}

// Off unsubscribes a function listener.
func (e *Event) Off(listener interface{}) {
	fn := reflect.ValueOf(listener)
	if fn.Kind() != reflect.Func {
		if e.Recoverer != nil {
			e.Recoverer(listener, ErrNotFunc)
		} else {
			panic(ErrNotFunc)
		}
	}

	e.mutex.Lock()
	delete(e.funcs, fn)
	e.mutex.Unlock()
}

// Trigger the event. This will call all function listeners.
func (e *Event) Trigger(arguments ...interface{}) {
	e.trigger(nil, arguments...)
}

// TriggerWait triggers the event and waits for all functions to execute.
func (e *Event) TriggerWait(arguments ...interface{}) {
	var wg sync.WaitGroup
	e.trigger(&wg, arguments...)
	wg.Wait()
}

func (e *Event) trigger(wg *sync.WaitGroup, arguments ...interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if wg != nil {
		wg.Add(len(e.funcs))
	}

	for fn, once := range e.funcs {
		if once {
			delete(e.funcs, fn)
		}

		go func(fn reflect.Value) {
			if wg != nil {
				defer wg.Done()
			}

			// Recover from potential panics, supplying them to a
			// Recoverer if one has been set, else allowing
			// the panic to occur.
			if e.Recoverer != nil {
				defer func() {
					if r := recover(); nil != r {
						e.Recoverer(fn.Interface(), fmt.Errorf("%v", r))
					}
				}()
			}

			var values []reflect.Value

			for i := 0; i < len(arguments); i++ {
				if arguments[i] == nil {
					values = append(values, reflect.New(fn.Type().In(i)).Elem())
				} else {
					values = append(values, reflect.ValueOf(arguments[i]))
				}
			}

			fn.Call(values)
		}(fn)
	}
}
