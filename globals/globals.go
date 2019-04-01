package globals

import "sync"

var locks []*sync.RWMutex

func SetX(x []*sync.RWMutex) {
	if locks != nil {
		panic("globals can only be set once for thread safety reasons")
	}
	locks = x
}

func GetX() []*sync.RWMutex {
	return locks
}

// Initialize locks for a number of users
//
// Only for the convenience of Go initialization (not modeled in GoLayer).
func Init(numUsers uint64) {
	var ls []*sync.RWMutex
	for i := uint64(0); i < numUsers; i++ {
		ls = append(ls, new(sync.RWMutex))
	}
	SetX(ls)
}

// Simulate a crash, returning globals to an uninitialized state.
func Shutdown() {
	locks = nil
}
