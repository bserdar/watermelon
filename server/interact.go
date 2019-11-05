package server

import (
	"sync"
)

var interactionMutex = sync.Mutex{}

// LockForInteraction prevents multiple goroutines from entering interaction
func LockForInteraction() {
	interactionMutex.Lock()
}

func UnlockForInteraction() {
	interactionMutex.Unlock()
}
