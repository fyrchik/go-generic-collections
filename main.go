package main

import (
	"sync"
	"fmt"
	"reflect"
	"strings"
)

type Policy func(func(interface{}, interface{}), interface{}, ...interface{})

func RunAllAndWait(iter interface{}, f func(...interface{})) {
	wg := &sync.WaitGroup{}
	switch t := reflect.ValueOf(iter); t.Type().Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			wg.Add(1)
			go func(a ...interface{}) {
				defer wg.Done()
				f(a...)
			}(i, t.Index(i).Interface())
		}
	case reflect.Chan:
		for {
			v, ok := t.Recv()
			if !ok {
				break
			}
			wg.Add(1)
			go func(a ...interface{}) {
				defer wg.Done()
				f(a...)
			}(v.Interface())
		}
	case reflect.Map:
		for _, k := range t.MapKeys() {
			wg.Add(1)
			go func(a ...interface{}) {
				defer wg.Done()
				f(a...)
			}(k.Interface(), t.MapIndex(k).Interface())
		}
	}
	wg.Wait()
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