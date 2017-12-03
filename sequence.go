package factory

import (
	"sync"
)

type sequence struct {
	first        int64
	value        int64
	roundStarted bool
	mux          sync.Mutex
}

// peek returns the current cursor number of the sequence
func (seq *sequence) peek() int64 {
	seq.mux.Lock()
	defer seq.mux.Unlock()

	if seq.value < seq.first {
		return seq.first
	}
	return seq.value
}

// next move the cursor of the sequence to next number
func (seq *sequence) next() int64 {
	seq.mux.Lock()
	defer seq.mux.Unlock()
	if seq.value < seq.first || (seq.value == seq.first && !seq.roundStarted) {
		seq.value = seq.first
		seq.roundStarted = true
	} else {
		seq.value = seq.value + 1
	}

	return seq.value
}

// rewind moves the cursor of the sequence to the start number of the sequence
func (seq *sequence) rewind() {
	seq.mux.Lock()
	defer seq.mux.Unlock()
	seq.value = seq.first
	seq.roundStarted = false

	return
}
