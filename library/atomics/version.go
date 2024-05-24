package atomics

import "sync/atomic"

// Version is a variable of type String that represents the version of the software.
// It can be set and retrieved using the Set() and Get() methods.
var Version = String{pointer: atomic.Value{}}
