package atomics

import "sync/atomic"

// Service is a variable of type String that represents the service being used.
// It can be set and retrieved using the Set() and Get() methods.
var Service = String{pointer: atomic.Value{}}
