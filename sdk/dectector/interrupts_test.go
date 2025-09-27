package dectector

import (
	"sync"
	"testing"
	"time"
)

func TestInterruptDetection(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	// Start the interrupt manager in a separate goroutine
	interrupt.InterruptsManager()

	// Simulate agent speaking first
	wg.Add(1)
	go func() {
		defer wg.Done()

		interrupt.AgentSpoke(true)

		time.Sleep(4 * time.Second) // Agent keeps speaking for 1 second
		interrupt.AgentSpoke(false)
	}()

	// Simulate user speaking while agent is speaking
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // User starts speaking after 500ms
		for i := 0; i < 5; i++ {
			interrupt.UserSpoke(true)
			time.Sleep(100 * time.Microsecond)
		}
		time.Sleep(500 * time.Millisecond) // User keeps speaking
		interrupt.UserSpoke(false)
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Verify if an interrupt was detected
	if !interrupt.IsCooling() {
		t.Fatal("Interrupt was expected but not detected!")
	}
}
func TestUserStartsBeforeAgent(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	interrupt.InterruptsManager()

	// User starts speaking first
	wg.Add(1)
	go func() {
		defer wg.Done()
		interrupt.UserSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.UserSpoke(false)
	}()

	// Agent starts speaking later
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // Agent starts after 1 sec
		interrupt.AgentSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.AgentSpoke(false)
	}()

	wg.Wait()

	if interrupt.IsInterrupt() {
		t.Fatal("Unexpected interrupt detected!")
	}
}
func TestAgentAndUserStartTogether(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	interrupt.InterruptsManager()

	wg.Add(2)
	go func() {
		defer wg.Done()
		interrupt.AgentSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.AgentSpoke(false)
	}()

	go func() {
		defer wg.Done()
		interrupt.UserSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.UserSpoke(false)
	}()

	wg.Wait()

	if !interrupt.IsCooling() {
		t.Fatal("Interrupt was expected but not detected!")
	}
}
func TestUserInterruptsMultipleTimes(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	interrupt.InterruptsManager()

	wg.Add(2)
	go func() {
		defer wg.Done()
		interrupt.AgentSpoke(true)
		time.Sleep(4 * time.Second)
		interrupt.AgentSpoke(false)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // User starts speaking after 1 sec
		for i := 0; i < 3; i++ {
			interrupt.UserSpoke(true)
			time.Sleep(500 * time.Millisecond)
			interrupt.UserSpoke(false)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	wg.Wait()

	if !interrupt.IsCooling() {
		t.Fatal("Interrupt was expected but not detected!")
	}
}
func TestAlternatingSpeechNoInterrupt(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	interrupt.InterruptsManager()

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			interrupt.AgentSpoke(true)
			time.Sleep(500 * time.Millisecond)
			interrupt.AgentSpoke(false)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(250 * time.Millisecond) // Start slightly later
		for i := 0; i < 3; i++ {
			interrupt.UserSpoke(true)
			time.Sleep(500 * time.Millisecond)
			interrupt.UserSpoke(false)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	wg.Wait()

	if interrupt.IsInterrupt() {
		t.Fatal("Unexpected interrupt detected!")
	}
}
func TestAgentStopsBeforeUserStarts(t *testing.T) {
	interrupt := &Interrupt{AgentResponse: func(b bool, s string) {}}

	var wg sync.WaitGroup

	interrupt.InterruptsManager()

	wg.Add(2)
	go func() {
		defer wg.Done()
		interrupt.AgentSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.AgentSpoke(false)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second) // User starts after agent stops
		interrupt.UserSpoke(true)
		time.Sleep(2 * time.Second)
		interrupt.UserSpoke(false)
	}()

	wg.Wait()

	if interrupt.IsInterrupt() {
		t.Fatal("Unexpected interrupt detected!")
	}
}
