# Event - A simple event emitter for Go

[![GoDoc](https://godoc.org/github.com/desertbit/event?status.svg)](https://godoc.org/github.com/desertbit/event)
[![Go Report Card](https://goreportcard.com/badge/github.com/desertbit/event)](https://goreportcard.com/report/github.com/desertbit/event)

This project was inspired by [the emission package](https://github.com/chuckpreslar/emission).
Instead of passing strings as event keys, I prefer to have one struct per event.

## Sample

```go
package main

import (
	"fmt"

	"github.com/desertbit/event"
)

func main() {
	e := event.New()
	e.On(onEvent)
	e.Once(onceEvent)

	e.TriggerWait("Hello World", 1)
	e.TriggerWait("Hello World", 2)

	e.Off(onEvent)
	e.TriggerWait("Hello World", 3)
}

func onEvent(s string, i int) {
	fmt.Println("on:", s, i)
}

func onceEvent(s string, i int) {
	fmt.Println("once:", s, i)
}
```
