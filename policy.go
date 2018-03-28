package ploop

import (
	"sync"
	"sync/atomic"
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

func NewBoundedParallel(max int32) *BoundedParallel {
	return &BoundedParallel{
		max: max,
		c: make(chan struct{}, max),
	}
}

func (bp *BoundedParallel) Execute(f Func, args ...interface{}) {
	bp.wg.Add(1)
	for {
		if n := bp.n; n < bp.max {
			if atomic.CompareAndSwapInt32(&bp.n, n, n+1) {
				go func(a ...interface{}) {
					defer func() {
						for n := bp.n; !atomic.CompareAndSwapInt32(&bp.n, n, n-1); {}
						bp.wg.Done()
						bp.c <- struct{}{}
					}()
					f(a...)
				}(args...)
				break
			}
		} else {
			select {
			case <-bp.c:
			}
		}
	}
}

func (bp *BoundedParallel) Wait() {
	bp.wg.Wait()
}

func RunBoundedParallel(max int32, iter interface{}, f Func) {
	RunWithPolicy(NewBoundedParallel(max), iter, f)
}
