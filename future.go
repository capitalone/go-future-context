package future

import (
	"sync"
	"time"
)

type Interface interface {
	Get() (interface{}, error)
	GetUntil(ms int) (interface{}, bool, error)
}

func New(inFunc func() (interface{}, error)) Interface {
	f := futureImpl{}
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.val, f.err = inFunc()
	}()
	return &f
}

type futureImpl struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

func (f *futureImpl) Get() (interface{}, error) {
	f.wg.Wait()
	return f.val, f.err
}

func (f *futureImpl) GetUntil(ms int) (interface{}, bool, error) {
	c := make(chan struct{})
	go func() {
		f.Get()
		close(c)
	}()
	select {
	case <-c:
		return f.val, false, f.err
	case <-time.After(time.Duration(ms) * time.Millisecond):
		return nil, true, nil
	}
	//should never get here
	return nil, false, nil
}

