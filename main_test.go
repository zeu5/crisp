package crisp

import (
	"testing"
)

func TestMain(t *testing.T) {
	for i := 0; i < 100; i++ {
		// l := new(sync.Mutex)
		val := 0

		go func() {
			// l.Lock()
			// defer l.Unlock()
			val = 1
		}()
		if val != 0 {
			t.Fatalf("Error, val: %d", val)
		}
	}
}

func TestMain2(t *testing.T) {
	CheckRandom(func() {
		// l := new(sync.Mutex)
		val := 0

		Go(func() {
			// l.Lock()
			// defer l.Unlock()
			val = 1
		})
		Assert(val == 0, "val should be 0")
	}, t, 100)
}
