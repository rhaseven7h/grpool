package grpool

import (
    "fmt"
    "sync"
)

func ExampleNew() {
    const totalCount = 10000
    var value int
    var mutex sync.Mutex

    grp := New(50, 500)
    grp.Start()
    go func() {
        for i := 0; i < totalCount; i++ {
            grp.pipe <- func() {
                // Increment must be locked, otherwise, potentially, we might
                // not reach the expected total count, due to race conditions
                mutex.Lock()
                value++
                mutex.Unlock()
            }
        }
        grp.Stop()
    }()
    fmt.Printf("Running? before Wait(): %t\n", grp.IsRunning())
    grp.Wait()
    fmt.Printf("Running? after Wait(): %t\n", grp.IsRunning())
}
