package scheduler

// WorkerGen is an algorithm to schedule worker machine
// generator pick worker machine and attach jobs there
type WorkerGen interface {
	Next() interface{}
}
