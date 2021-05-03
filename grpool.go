package grpool

import (
    "errors"
    "sync"
)

// GRPool contains the configuration for the worker pool.
type GRPool struct {
    running    bool
    routines   uint
    bufferSize uint
    pipe       chan func()
    pipeOpen   bool
    waitGroup  sync.WaitGroup
}

// New creates a new GRPool object to control a set of worker
// go routines.
//
// The routines parameter is the number of go routines to be launched.
// The bufferSize parameter is the number of jobs to be queued.
//
// Submitting more than bufferSize jobs is allowed, but probably
// will cause a lock on the Submit() method caller.
//
// It is also possible to specify 0 as the bufferSize, to
// create an unbuffered channel, and therefore, allow locking calls.
func New(routines uint, bufferSize uint) *GRPool {
    return &GRPool{
        routines:   routines,
        bufferSize: bufferSize,
        pipe:       make(chan func(), bufferSize),
        pipeOpen:   true,
    }
}

func (grp *GRPool) worker() {
    for job := range grp.pipe {
        job()
    }
    grp.waitGroup.Done()
}

// Start instantiates workers for this pool.
func (grp *GRPool) Start() {
    grp.running = true
    grp.waitGroup.Add(int(grp.routines))
    for i := uint(0); i < grp.routines; i++ {
        go grp.worker()
    }
}

// Wait blocks until all workers are finished.
// You would usually have a Stop() somewhere in your code
// to ensure this call gets unblocked at some point in time.
func (grp *GRPool) Wait() {
    grp.waitGroup.Wait()
    grp.running = false
}

// IsRunning returns true if there are workers running.
func (grp *GRPool) IsRunning() bool {
    return grp.running
}

// Stop closes the pipe of jobs, essentially stopping processing
// of any new jobs.
func (grp *GRPool) Stop() {
    close(grp.pipe)
    grp.pipeOpen = false
}

// Submit received a new job to be enqueued to be processed
// by this pool's workers. It will return an error if the pipe
// has already been closed by Stop(), or nil if the job was
// enqueued successfully.
func (grp *GRPool) Submit(job func()) error {
    if !grp.pipeOpen {
        return errors.New("job pipe is not open, Stop() has been called on this pool already")
    }
    grp.pipe <- job
    return nil
}
