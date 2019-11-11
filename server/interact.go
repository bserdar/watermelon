package server

import (
	"sync"
)

var interactionMutex = sync.Mutex{}

// LockForInteraction prevents multiple goroutines from entering user interaction
func LockForInteraction() {
	interactionMutex.Lock()
}

// UnlockForInteraction unlocks the interaction lock
func UnlockForInteraction() {
	interactionMutex.Unlock()
}
