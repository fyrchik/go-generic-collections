package ploop

import (
	"math/rand"
	"sync"
	"testing"
)

const iterSize = 10

func testPolicy(p Policy, t *testing.T) {
	numbers := make([]int, iterSize)
	for i := 0; i < iterSize; i++ {
		numbers[i] = rand.Int()
	}

	m := &sync.Map{}
	RunWithPolicy(p, numbers, func(args ...interface{}) {
		// type assertions for additional checking
		m.Store(args[0].(int), args[1].(int))
	})

	m.Range(func(key, value interface{}) bool {
		i, _ := key.(int)
		if numbers[i] != value {
			t.Fatalf("Value for key %d is equal to %d, expected %d", i, numbers[i], value)
		}
		return true
	})
}

func TestSequential(t *testing.T) {
	testPolicy(&Sequential{}, t)
}

func TestParallel(t *testing.T) {
	testPolicy(&AllParallel{}, t)
}

func TestBoundedParallel(t *testing.T) {
	max := int32(3)
	testPolicy(&BoundedParallel{
		max: max,
		wg:  sync.WaitGroup{},
		c:   make(chan struct{}, max),
	}, t)
}
