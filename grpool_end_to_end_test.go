package grpool_test

import (
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/rhaseven7h/grpool"
)

func TestEndToEnd(t *testing.T) {
    t.Run("End To End", func(t *testing.T) {
        const totalCount = 10000
        var value int
        var mutex sync.Mutex

        grp := grpool.New(50, 500)
        grp.Start()
        go func() {
            for i := 0; i < totalCount; i++ {
                err := grp.Submit(func() {
                    // Increment must be locked, otherwise, potentially, we might
                    // not reach the expected total count, due to race conditions
                    mutex.Lock()
                    value++
                    mutex.Unlock()
                })
                if err != nil {
                    panic(err)
                }
            }
            grp.Stop()
        }()
        wasRunning := grp.IsRunning()
        grp.Wait()

        assert.True(t, wasRunning)
        assert.False(t, grp.IsRunning())
        assert.Equal(t, totalCount, value)
    })
}
