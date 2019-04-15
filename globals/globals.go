package globals

import "sync"

var locks []*sync.RWMutex

// SetX initializes the global locks to a particular value.
//
// SetX can only be called once.
func SetX(x []*sync.RWMutex) {
	if locks != nil {
		panic("globals can only be set once for thread safety reasons")
	}
	locks = x
}

// GetX gets the current set of locks.
func GetX() []*sync.RWMutex {
	if locks == nil {
		panic("attempt to get uninitialized locks")
	}
	return locks
}

// Init initializes locks for a number of users
//
// Only for the convenience of Go initialization (not modeled in GoLayer).
func Init(numUsers uint64) {
	var ls []*sync.RWMutex
	for i := uint64(0); i < numUsers; i++ {
		ls = append(ls, new(sync.RWMutex))
	}
	SetX(ls)
}

// Shutdown simulates a crash, returning globals to an uninitialized state.
func Shutdown() {
	locks = nil
}
