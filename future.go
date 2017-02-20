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
	"time"
)

// Interface represents a future. No concrete implementation is
// exposed; all access to a future is via this interface.
type Interface interface {
	// Cancel prevents a future that hasn't completed from returning a
	// value. Any current or future calls to Get or GetUntil will return
	// immediately.
	//
	// If the future has already completed or has already been
	// cancelled, calling Cancel will do nothing.
	// After a successful cancel, IsCancelled returns true.
	//
	// Calling Cancel on a future that has not completed does not stop the
	// currently running function. However, any chained functions will not
	// be run and the values returned by the current function are not accessible.
	Cancel()

	// IsCancelled indicates if a future terminated due to cancellation.
	// If Cancel was called and the future's work was not completed, IsCancelled
	// returns true. Otherwise, it returns false
	IsCancelled() bool

	// Get returns the values calculated by the future. It will pause until
	// the future is cancelled or until the value is calculated.
	// If Get is invoked multiple times, the same value will be returned each time.
	// Subsequent calls to Get will return instantaneously.
	//
	// When the future is cancelled, nil is returned for both the value and the error.
	Get() (interface{}, error)

	// GetUntil waits for up to Duration d for the future to complete. If the
	// future completes before the Duration completes, the value and error are returned
	// and timeout is returned as false. If the Duration completes before the future
	// returns, nil is returned for the value and the error and timeout is returned
	// as true.
	//
	// When the future is cancelled, nil is returned for both the value and the error.
	GetUntil(d time.Duration) (interface{}, bool, error)

	// Then allows multiple function calls to be chained together into a single future.
	// Each call is run in order, with the output of the previous call passed into
	// the next function in the chain. If an error occurs at any step in the chain,
	// processing ceases and the error is returned via Get or GetUntil.
	//
	// If Cancel is called before the chain completes, the currently running function
	// will complete silently in the background and all unexecuted functions will
	// not run.
	Then(func(interface{}) (interface{}, error)) Interface
}

// New creates a new Future that wraps the provided function.
func New(inFunc func() (interface{}, error)) Interface {
	return NewWithContext(context.Background(), inFunc)
}

func newInner(cancelChan <-chan struct{}, cancelFunc context.CancelFunc, inFunc func() (interface{}, error)) Interface {
	f := futureImpl{
		done:       make(chan struct{}),
		cancelChan: cancelChan,
		cancelFunc: cancelFunc,
	}
	go func() {
		go func() {
			f.val, f.err = inFunc()
			close(f.done)
		}()
		select {
		case <-f.done:
			//do nothing, just waiting to see which will happen first
		case <-f.cancelChan:
			//do nothing, leave val and err nil
		}
	}()
	return &f
}

// NewWithContext creates a new Future that wraps the provided function and
// cancels when the Done channel of the provided context is closed.
func NewWithContext(ctx context.Context, inFunc func() (interface{}, error)) Interface {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return newInner(cancelCtx.Done(), cancelFunc, inFunc)
}

type futureImpl struct {
	done       chan struct{}
	cancelChan <-chan struct{}
	cancelFunc context.CancelFunc
	val        interface{}
	err        error
}

func (f *futureImpl) Cancel() {
	select {
	case <-f.done:
		return //already finished
	case <-f.cancelChan:
		return //already cancelled
	default:
		f.cancelFunc()
	}
}

func (f *futureImpl) IsCancelled() bool {
	select {
	case <-f.cancelChan:
		return true
	default:
		return false
	}
}

func (f *futureImpl) Get() (interface{}, error) {
	select {
	case <-f.done:
		return f.val, f.err
	case <-f.cancelChan:
		//on cancel, just fall out
	}
	return nil, nil
}

func (f *futureImpl) GetUntil(d time.Duration) (interface{}, bool, error) {
	select {
	case <-f.done:
		val, err := f.Get()
		return val, false, err
	case <-time.After(d):
		return nil, true, nil
	case <-f.cancelChan:
		//on cancel, just fall out
	}
	return nil, false, nil
}

func (f *futureImpl) Then(next func(interface{}) (interface{}, error)) Interface {
	nextFuture := newInner(f.cancelChan, f.cancelFunc, func() (interface{}, error) {
		result, err := f.Get()
		if f.IsCancelled() || err != nil {
			return result, err
		}
		return next(result)
	})
	return nextFuture
}
