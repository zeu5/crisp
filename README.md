# crisp

A concurrency test runner for go code

## Example

The following test does not fail when run with the normal testing engine. The obvious thread interleaving where the go routine is scheduled before the main thread executes the check is never explored.

``` go
import (
    "testing"
    "github.com/zeu5/crisp"
    "github.com/zeu5/crisp/sync"
)

func TestMain(t *testing.T) {
    target := func() {
        l := new(sync.Mutex)
        val := 0

        crisp.Go(func(){
            l.Lock()
            defer l.Unlock()
            val = 1
        })

        if val != 0 {
            t.Fatal("Error")
        }
    }

    crisp.CheckRandom(target, 100)
}
```
