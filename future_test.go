//Copyright 2016 Capital One Services, LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and limitations under the License.
// SPDX-Copyright: Copyright (c) Capital One Services, LLC
// SPDX-License-Identifier: Apache-2.0

package future

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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
	dur := end.Unix() - start.Unix()
	fmt.Println(v, err, dur)
	assert.Nil(t, err)
	assert.Equal(t, 10, v)
	assert.Equal(t, int64(5), dur)
}

func TestFutureGetUntil(t *testing.T) {
	fb := func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	}
	f := New(fb)
	start := time.Now()
	v, timeout, err := f.GetUntil(3 * time.Second)
	end := time.Now()
	dur := end.Unix() - start.Unix()
	fmt.Println(v, timeout, err, dur)
	assert.Nil(t, err)
	assert.Nil(t, v)
	assert.True(t, timeout)
	assert.True(t, dur < 5)

	start2 := time.Now()
	v, timeout, err = f.GetUntil(10 * time.Second)
	end2 := time.Now()
	dur2 := end2.Unix() - start2.Unix()
	fmt.Println(v, timeout, err, dur2)
	assert.Nil(t, err)
	assert.Equal(t, 10, v)
	assert.False(t, timeout)
	assert.True(t, dur < int64(10))
}

func TestThen(t *testing.T) {
	f := New(func() (interface{}, error) {
		return 10, nil
	}).Then(func(i interface{}) (interface{}, error) {
		return 2 * i.(int), nil
	}).Then(func(i interface{}) (interface{}, error) {
		return 2 + i.(int), nil
	})
	result, err := f.Get()
	fmt.Println(result, err)
	assert.Nil(t, err)
	assert.Equal(t, 22, result)

	g := New(func() (interface{}, error) {
		return nil, errors.New("This is an error")
	}).Then(func(i interface{}) (interface{}, error) {
		return 2 * i.(int), nil
	})
	result, err = g.Get()
	fmt.Println(result, err)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "This is an error", err.Error())

	h := New(func() (interface{}, error) {
		return 10, nil
	}).Then(func(i interface{}) (interface{}, error) {
		return nil, errors.New("This is also an error")
	})
	result, err = h.Get()
	fmt.Println(result, err)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "This is also an error", err.Error())
}

func TestCancel(t *testing.T) {
	f := New(func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	})
	go func() {
		time.Sleep(2 * time.Second)
		f.Cancel()
	}()
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
}

func TestCancelChain(t *testing.T) {
	var run1 int64
	var run2 int64
	f := New(func() (interface{}, error) {
		fmt.Println("1")
		time.Sleep(5 * time.Second)
		atomic.AddInt64(&run1, 1)
		fmt.Println("1 done")
		return 10, nil
	}).Then(func(i interface{}) (interface{}, error) {
		fmt.Println("2", i)
		time.Sleep(2 * time.Second)
		atomic.AddInt64(&run2, 1)
		return i.(int) * 2, nil
	})
	go func() {
		time.Sleep(2 * time.Second)
		f.Cancel()
		fmt.Println("cancelled", f.IsCancelled())
	}()
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
	time.Sleep(7 * time.Second)
	assert.Equal(t, int64(1), atomic.LoadInt64(&run1))
	assert.Equal(t, int64(0), atomic.LoadInt64(&run2))
}

func TestCancelTimer(t *testing.T) {
	f := New(func() (interface{}, error) {
		time.Sleep(10 * time.Second)
		return 10, nil
	})
	go func() {
		time.Sleep(7 * time.Second)
		f.Cancel()
	}()
	result, timeout, err := f.GetUntil(5 * time.Second)
	fmt.Println(result, timeout, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, timeout)
	assert.False(t, f.IsCancelled())

	result, timeout, err = f.GetUntil(5 * time.Second)
	fmt.Println(result, timeout, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.False(t, timeout)
	assert.True(t, f.IsCancelled())
}

func TestCancelAfterDone(t *testing.T) {
	f := New(func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	})
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Equal(t, 10, result)
	assert.Nil(t, err)
	assert.False(t, f.IsCancelled())
	f.Cancel()
	result, err = f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Equal(t, 10, result)
	assert.Nil(t, err)
	assert.False(t, f.IsCancelled())
}

func TestCancelTwice(t *testing.T) {
	f := New(func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	})
	go func() {
		time.Sleep(2 * time.Second)
		f.Cancel()
	}()
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
	f.Cancel()
	result, err = f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
}

func TestNewWithContextTimeout(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	f := NewWithContext(ctx, func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	})
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
	assert.Equal(t, context.DeadlineExceeded, ctx.Err())
}

func TestNewWithContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	f := NewWithContext(ctx, func() (interface{}, error) {
		time.Sleep(5 * time.Second)
		return 10, nil
	})
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()
	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
	assert.Nil(t, result)
	assert.Nil(t, err)
	assert.True(t, f.IsCancelled())
	assert.Equal(t, context.Canceled, ctx.Err())
}

// test case from Dave Cheney's bug report
// since this test runs forever, I have
// it marked as skipped.
func TestCancelConcurrent(t *testing.T) {
	t.SkipNow()
	loop := func() {
		const N = 2000
		start := make(chan int)
		var done sync.WaitGroup
		done.Add(N)
		f := New(func() (interface{}, error) { select {}; return 1, nil })
		for i := 0; i < N; i++ {
			go func() {
				defer done.Done()
				<-start
				f.Cancel()
			}()
		}
		close(start)
		done.Wait()
	}
	for {
		loop()
	}

}
