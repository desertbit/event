package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	e := New()
	var count int64

	e.On(func() {
		atomic.AddInt64(&count, 1)
	})
	e.On(func() {
		atomic.AddInt64(&count, 1)
	})
	e.On(func() {
		atomic.AddInt64(&count, 1)
	})
	e.Once(func() {
		atomic.AddInt64(&count, 1)
	})
	e.Once(func() {
		atomic.AddInt64(&count, 1)
	})

	e.TriggerWait()
	if count != 5 {
		t.Fatal(count)
	}

	count = 0
	e.TriggerWait()
	if count != 3 {
		t.Fatal(count)
	}

	f := func() {
		atomic.AddInt64(&count, 1)
	}
	e.On(f)

	count = 0
	e.TriggerWait()
	if count != 4 {
		t.Fatal(count)
	}

	e.Off(f)
	count = 0
	e.TriggerWait()
	if count != 3 {
		t.Fatal(count)
	}

	e.Off(f)
	count = 0
	e.TriggerWait()
	if count != 3 {
		t.Fatal(count)
	}

	e.Once(f)
	e.Off(f)
	count = 0
	e.TriggerWait()
	if count != 3 {
		t.Fatal(count)
	}
}

func TestEventTrigger(t *testing.T) {
	e := New()
	var count int64
	var wg sync.WaitGroup

	e.On(func() {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})
	e.On(func() {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})
	e.On(func() {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})
	e.Once(func() {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})
	e.Once(func() {
		atomic.AddInt64(&count, 1)
		wg.Done()
	})

	wg.Add(5)
	e.Trigger()
	wg.Wait()
	if count != 5 {
		t.Fatal(count)
	}

	wg.Add(3)
	count = 0
	e.Trigger()
	wg.Wait()
	if count != 3 {
		t.Fatal(count)
	}

}

func TestEventArgs(t *testing.T) {
	e := New()
	var count int64

	f := func(i int64, s string, n interface{}) {
		atomic.AddInt64(&count, i)
	}

	ff := func(i int64, s string, n interface{}) {
		atomic.AddInt64(&count, i)
	}

	e.On(f)
	e.Once(ff)

	e.TriggerWait(int64(2), "Hallo Welt", nil)
	if count != 4 {
		t.Fatal(count)
	}

	count = 0
	e.TriggerWait(int64(2), "Hallo Welt", nil)
	if count != 2 {
		t.Fatal(count)
	}

	e.Once(ff)
	e.Off(f)
	e.Off(ff)
	count = 0
	e.TriggerWait(int64(2), "Hallo Welt", nil)
	if count != 0 {
		t.Fatal(count)
	}
}

func TestEventPanic(t *testing.T) {
	e := New()
	var err error

	err = nil
	func() {
		defer func() {
			if r := recover(); nil != r {
				err = r.(error)
			}
		}()
		e.On(5)
	}()
	if err != ErrNotFunc {
		t.Fatal(err)
	}

	err = nil
	func() {
		defer func() {
			if r := recover(); nil != r {
				err = r.(error)
			}
		}()
		e.Once(5)
	}()
	if err != ErrNotFunc {
		t.Fatal(err)
	}

	err = nil
	func() {
		defer func() {
			if r := recover(); nil != r {
				err = r.(error)
			}
		}()
		e.Off(5)
	}()
	if err != ErrNotFunc {
		t.Fatal(err)
	}
}

func TestEventRecoverer(t *testing.T) {
	e := New()
	var err error

	e.Recoverer = func(i interface{}, rerr error) {
		err = rerr
	}

	e.On(func() {})
	e.TriggerWait(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.On(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.Once(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.Off(5)
	if err == nil {
		t.Fatal()
	}
}

func TestEventRecovererNew(t *testing.T) {
	var err error
	e := New(func(i interface{}, rerr error) {
		err = rerr
	})

	e.On(func() {})
	e.TriggerWait(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.On(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.Once(5)
	if err == nil {
		t.Fatal()
	}

	err = nil
	e.Off(5)
	if err == nil {
		t.Fatal()
	}
}

func TestEventConcurrentMutexLock(t *testing.T) {
	e := New()

	f := func() {
		time.Sleep(time.Second)
	}
	e.On(f)

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		e.TriggerWait()
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)

	e.On(f)
	e.Once(f)
	e.Off(f)

	if int(time.Since(start).Seconds()) != 0 {
		t.Fatalf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}

	wg.Wait()
}
