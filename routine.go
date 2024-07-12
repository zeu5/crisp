package crisp

type routineState int

const (
	blocked routineState = iota
	finished
	available
	runToCompletion
)

type routine struct {
	f     func()
	state routineState
	id    int64

	in  chan struct{}
	out chan bool
}

func newRoutine(f func()) *routine {
	return &routine{
		f:     f,
		state: available,
		id:    -1,

		in:  make(chan struct{}),
		out: make(chan bool),
	}
}

func (r *routine) run() {
	<-r.in
	r.f()
	r.state = finished
	r.out <- true
	close(r.in)
	close(r.out)
}

func (r *routine) block() {
	r.state = blocked
}

func (r *routine) resume() bool {
	if r.state == finished {
		return true
	}
	r.in <- struct{}{}
	return <-r.out
}

func (r *routine) yield() {
	if r.state == runToCompletion {
		return
	}
	r.out <- false
	<-r.in
}

func (r *routine) done() bool {
	return r.state == finished
}

func (r *routine) resumePermanently() {
	r.state = runToCompletion
	close(r.in)
}
