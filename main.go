package ploop

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

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
