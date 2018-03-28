package main

import (
	"sync"
	"fmt"
	"reflect"
	"strings"
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

// AllParallel policy spawns a separate goroutine in every Execute
// and waits for all of them to finish
type AllParallel struct {
	wg sync.WaitGroup
}

func (s *AllParallel) Execute(f Func, args ...interface{}) {
	s.wg.Add(1)
	go func(a ...interface{}) {
		defer s.wg.Done()
		f(a...)
	}(args...)
}

func (s *AllParallel) Wait() {
	s.wg.Wait()
}

func RunAllAndWait(iter interface{}, f func(...interface{})) {
	var ap Policy = &AllParallel{}
	switch t := reflect.ValueOf(iter); t.Type().Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			ap.Execute(f, i, t.Index(i).Interface())
		}
	case reflect.Chan:
		for {
			v, ok := t.Recv()
			if !ok {
				break
			}
			ap.Execute(f, v.Interface())
		}
	case reflect.Map:
		for _, k := range t.MapKeys() {
			ap.Execute(f, k.Interface(), t.MapIndex(k).Interface())
		}
	}
	ap.Wait()
}

func CastSlice(f interface{}) []interface{} {
	if reflect.TypeOf(f).Kind() != reflect.Slice {
		panic("Not a slice!")
	}

	sl := reflect.ValueOf(f)
	ret := make([]interface{}, sl.Len())
	for i := 0; i < sl.Len(); i++ {
		ret[i] = sl.Index(i).Interface()
	}
	return ret
}

func main() {
	strSlice := strings.Split("this is another sentence", " ")
	strChan  := make(chan string, len(strSlice))
	strMap   := make(map[int]string)
	for i, s := range strSlice {
		strMap[i] = s
		strChan <- s
	}
	close(strChan)
	RunAllAndWait(strSlice, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Slice[%02d] = %s\n", index, value)
	})
	RunAllAndWait(strMap, func(args ...interface{}) {
		index := args[0].(int)
		value := args[1].(string)
		fmt.Printf("Map[%0d] = %s\n", index, value)
	})
	RunAllAndWait(strChan, func(args ...interface{}) {
		value := args[0].(string)
		fmt.Printf("Channel recv = %s\n", value)
	})
}