package future

import (
	"testing"
	"time"
	"fmt"
)

func TestFutureGet(t *testing.T) {
	fb := func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	}
	f := New(fb)
	start := time.Now()
	v, err := f.Get()
	end := time.Now()
	fmt.Println(v, err, end.Unix() - start.Unix())
}

func TestFutureGetUntil(t *testing.T) {
	fb := func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	}
	f := New(fb)
	start := time.Now()
	v, timeout, err := f.GetUntil(3000)
	end := time.Now()
	fmt.Println(v, timeout, err, end.Unix() - start.Unix())
	v, timeout, err = f.GetUntil(3000)
	fmt.Println(v, timeout, err, end.Unix() - start.Unix())
}
