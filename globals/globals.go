package globals

import "sync"

var locks []*sync.RWMutex

func Set(x []*sync.RWMutex) {
	if locks != nil {
		panic("globals can only be set once for thread safety reasons")
	}
	locks = x
}

func Get() []*sync.RWMutex {
	return locks
}

// Initialize locks for a number of users
func Init(users uint64) {
	var ls []*sync.RWMutex
	for i := uint64(0); i < users; i++ {
		ls = append(ls, new(sync.RWMutex))
	}
	Set(ls)
}

// Simulate a crash, returning globals to an uninitialized state.
func Shutdown() {
	locks = nil
}
