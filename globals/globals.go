package globals

import "sync"

var locks []*sync.Mutex

// SetX initializes the global locks to a particular value.
//
// SetX can only be called once.
func SetX(x []*sync.Mutex) {
	if locks != nil {
		panic("globals can only be set once for thread safety reasons")
	}
	locks = x
}

// GetX gets the current set of locks.
func GetX() []*sync.Mutex {
	if locks == nil {
		panic("attempt to get uninitialized locks")
	}
	return locks
}

// Shutdown simulates a crash, returning globals to an uninitialized state.
func Shutdown() {
	locks = nil
}
