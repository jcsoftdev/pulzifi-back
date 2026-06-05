// Package scheduler internal tests for goroutine panic recovery. A panic in a
// scheduled goroutine without recover() crashes the whole worker process, which
// silently stops ALL tenants' monitoring until the process is restarted.
package scheduler

import (
	"testing"
	"time"
)

// TestSafeGo_RecoversPanicWithoutCrashing proves a panic inside the scheduled
// function is recovered. Without recover(), the panic would abort the test
// binary itself — so reaching the assertion proves the panic was contained.
func TestSafeGo_RecoversPanicWithoutCrashing(t *testing.T) {
	ran := make(chan struct{})
	safeGo("test-panic", func() {
		defer close(ran)
		panic("boom")
	})

	select {
	case <-ran:
		// The deferred close ran during panic unwinding; safeGo's recover then
		// contained the panic. If it had not, the process would have crashed.
		time.Sleep(20 * time.Millisecond) // let safeGo's deferred recover finish
	case <-time.After(2 * time.Second):
		t.Fatal("safeGo did not run the function")
	}
}

// TestSafeGo_RunsFunctionNormally proves safeGo still executes non-panicking work.
func TestSafeGo_RunsFunctionNormally(t *testing.T) {
	got := make(chan int, 1)
	safeGo("test-normal", func() { got <- 42 })

	select {
	case v := <-got:
		if v != 42 {
			t.Fatalf("got %d, want 42", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("safeGo did not run the function")
	}
}
