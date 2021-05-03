package grpool

import (
    "regexp"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestGRPool_Start(t *testing.T) {
    t.Run("Start", func(t *testing.T) {
        var value int
        grp := &GRPool{
            running:   false,
            routines:  1,
            pipe:      make(chan func()),
            waitGroup: sync.WaitGroup{},
        }
        grp.Start()
        assert.True(t, grp.running)
        grp.pipe <- func() {
            value = 777
        }
        <-time.NewTimer(100 * time.Millisecond).C
        assert.Equal(t, 777, value)
        close(grp.pipe)
        grp.waitGroup.Wait()
        assert.True(t, true) // Just make sure we get here
    })
}

func TestGRPool_Stop(t *testing.T) {
    t.Run("Stop", func(t *testing.T) {
        grp := &GRPool{
            pipe: make(chan func()),
        }
        go func() {
            t := time.NewTimer(time.Second)
            <-t.C
            grp.Stop()
        }()
        <-grp.pipe
        assert.True(t, true) // Just check we got here
    })
}

func TestGRPool_Submit(t *testing.T) {
    for _, tt := range []struct {
        name           string
        markPipeAsOpen bool
        expectedValue  int
        expectedError  *regexp.Regexp
    }{
        {
            name:           "Success",
            markPipeAsOpen: true,
            expectedValue:  777,
            expectedError:  nil,
        },
        {
            name:           "Pipe Closed",
            markPipeAsOpen: false,
            expectedValue:  -1,
            expectedError:  regexp.MustCompile("pipe is not open"),
        },
    } {
        t.Run(tt.name, func(t *testing.T) {
            tt := tt
            var value int
            var err error = nil
            grp := &GRPool{
                pipe:     make(chan func()),
                pipeOpen: tt.markPipeAsOpen,
            }
            go func() {
                <-time.NewTimer(100 * time.Millisecond).C
                err = grp.Submit(func() {
                    value = tt.expectedValue
                })
                <-time.NewTimer(100 * time.Millisecond).C
                grp.Stop()
            }()
            job := <-grp.pipe
            if job != nil {
                job()
            }
            if tt.expectedError != nil {
                if assert.NotNil(t, err) {
                    assert.Regexp(t, tt.expectedError, err.Error())
                }
            } else {
                if assert.Nil(t, err) {
                    assert.Equal(t, tt.expectedValue, value)
                }
            }
        })
    }
}

func TestGRPool_worker(t *testing.T) {
    t.Run("Worker", func(t *testing.T) {
        var value int
        grp := &GRPool{
            pipe:      make(chan func()),
            waitGroup: sync.WaitGroup{},
        }
        grp.waitGroup.Add(1)
        go grp.worker()
        job := func() {
            value = 777
        }
        go func() {
            done := false
            jobSendTimer := time.NewTimer(200 * time.Millisecond)
            pipeCloseTimer := time.NewTimer(500 * time.Millisecond)
            for !done {
                select {
                case <-jobSendTimer.C:
                    grp.pipe <- job
                case <-pipeCloseTimer.C:
                    close(grp.pipe)
                    done = true
                }
            }
        }()
        grp.waitGroup.Wait()
        assert.Equal(t, 777, value)
    })
}

func TestNew(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        grp := New(5, 50)
        if assert.NotNil(t, grp) {
            assert.False(t, grp.running)
            assert.Equal(t, uint(5), grp.routines)
            assert.Equal(t, uint(50), grp.bufferSize)
            assert.Equal(t, 50, cap(grp.pipe))
        }
    })
}

func TestGRPool_Wait(t *testing.T) {
    t.Run("Wait", func(t *testing.T) {
        grp := &GRPool{
            waitGroup: sync.WaitGroup{},
        }
        grp.waitGroup.Add(1)
        grp.running = true
        go func() {
            t := time.NewTimer(100 * time.Millisecond)
            <-t.C
            grp.waitGroup.Done()
        }()
        grp.Wait()
        assert.False(t, grp.running) // Just check that we got here
    })
}

func TestGRPool_IsRunning(t *testing.T) {
    t.Run("Is Running?", func(t *testing.T) {
        grp := &GRPool{}
        assert.False(t, grp.IsRunning())
        grp.running = true
        assert.True(t, grp.IsRunning())
    })
}
