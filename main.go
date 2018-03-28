package main

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Func is generic function type to invoke
type Func func(...interface{})

// Policy incapsulates all parallelizing logic
type Policy interface {
	// Execute executes specified function with provided arguments
	Execute(Func, ...interface{})

	// Wait waits for all previously executed workers to finish
	Wait()
}

// Sequential just execute functions in the same thread
// It is here mosly for illustrative and testing purposes
type Sequential struct{}

func (s *Sequential) Execute(f Func, args ...interface{}) {
	f(args...)
}

func (s *Sequential) Wait() {}

func RunSequential(iter interface{}, f Func) {
	RunWithPolicy(&Sequential{}, iter, f)
}

// AllParallel policy spawns a separate goroutine in every Execute
// and waits for all of them to finish
type AllParallel struct {
	wg sync.WaitGroup
}

func (ap *AllParallel) Execute(f Func, args ...interface{}) {
	ap.wg.Add(1)
	go func(a ...interface{}) {
		defer ap.wg.Done()
		f(a...)
	}(args...)
}

func (ap *AllParallel) Wait() {
	ap.wg.Wait()
}

func RunParallel(iter interface{}, f Func) {
	RunWithPolicy(&AllParallel{}, iter, f)
}

// BoundedParallel policy works like AllParallel, but also ensures
// that no more than N workers are spawned simultaneously.
// If there are already N goroutines spawned, Execute blocks until one
// of them exits.
type BoundedParallel struct {
	max int32
	n   int32
	wg  sync.WaitGroup
	c   chan struct{}
}

func (bp *BoundedParallel) Execute(f Func, args ...interface{}) {
	bp.wg.Add(1)
	for {
		fmt.Println(bp.n, bp.max)
		if bp.n < bp.max && atomic.CompareAndSwapInt32(&bp.n, bp.n, bp.n+1) {
			go func(a ...interface{}) {
				defer func() {
					atomic.CompareAndSwapInt32(&bp.n, bp.n, bp.n-1)
					bp.wg.Done()
					bp.c <- struct{}{}
				}()
				f(a...)
			}(args...)
			break
		}
		select {
		case <-bp.c:
		}
	}
}

func (bp *BoundedParallel) Wait() {
	bp.wg.Wait()
}

func RunBoundedParallel(max int32, iter interface{}, f Func) {
	RunWithPolicy(&BoundedParallel{
		n:   0,
		max: max,
		c:   make(chan struct{}, max),
	}, iter, f)
}

func RunWithPolicy(p Policy, iter interface{}, f Func) {
	switch t := reflect.ValueOf(iter); t.Type().Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			p.Execute(f, i, t.Index(i).Interface())
		}
	case reflect.Chan:
		for {
			v, ok := t.Recv()
			if !ok {
				break
			}
			p.Execute(f, v.Interface())
		}
	case reflect.Map:
		for _, k := range t.MapKeys() {
			p.Execute(f, k.Interface(), t.MapIndex(k).Interface())
		}
	}
	p.Wait()
}

func main() {
	strSlice := strings.Split("this is another sentence", " ")
	strChan := make(chan string, len(strSlice))
	strMap := make(map[int]string)
	for i, s := range strSlice {
		strMap[i] = s
		strChan <- s
	}
	close(strChan)

	fmt.Println("\nParallel test:")
	RunParallel(strSlice, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Slice[%02d] = %s\n", index, value)
	})
	RunParallel(strMap, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Map[%0d] = %s\n", index, value)
	})
	RunParallel(strChan, func(args ...interface{}) {
		value := args[0].(string)
		fmt.Printf("Channel recv = %s\n", value)
	})

	fmt.Println("\nSequential test:")
	strChan = make(chan string, len(strSlice))
	for _, s := range strSlice {
		strChan <- s
	}
	close(strChan)
	RunSequential(strSlice, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Slice[%02d] = %s\n", index, value)
	})
	RunSequential(strMap, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Map[%0d] = %s\n", index, value)
	})
	RunSequential(strChan, func(args ...interface{}) {
		value := args[0].(string)
		fmt.Printf("Channel recv = %s\n", value)
	})

	fmt.Println("\nBoundedParallel test:")
	strChan = make(chan string, len(strSlice))
	for _, s := range strSlice {
		strChan <- s
	}
	close(strChan)
	max := int32(2)
	RunBoundedParallel(max, strSlice, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Slice[%02d] = %s\n", index, value)
		time.Sleep(time.Second)
	})
	RunBoundedParallel(max, strMap, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Map[%0d] = %s\n", index, value)
		time.Sleep(time.Second)
	})
	RunBoundedParallel(max, strChan, func(args ...interface{}) {
		value := args[0].(string)
		fmt.Printf("Channel recv = %s\n", value)
		time.Sleep(time.Second)
	})
}
