# ** Capital One built this project to help our engineers as well as users in the community. We are no longer able to fully support the project. We have archived the project as of Jul 9 2019 where it will be available in a read-only state. Feel free to fork the project and maintain your own version. **

[![Go Report Card](https://goreportcard.com/badge/github.com/capitalone/go-future-context)](https://goreportcard.com/report/github.com/capitalone/go-future-context)
[![Sourcegraph](https://sourcegraph.com/github.com/capitalone/go-future-context/-/badge.svg)](https://sourcegraph.com/github.com/capitalone/go-future-context?badge)

# future

A simple Future (Promise) library for Go.

Usage
-----

Basic usage (wait forever):
```go
package main

import (
  "fmt"
  future "github.com/capitalone/go-future-context"
)

func ThingThatTakesALongTimeToCalculate(inVal int) (string, error) {
  //this does something but it's not that important
  return "Hello", nil
}

func main() {
  inVal := 200
  f := future.New(func() (interface{}, error) {
    return ThingThatTakesALongTimeToCalculate(inVal)
  })
  
  result, err := f.Get()
  fmt.Println(result, err)
}
```

Timeout usage (wait for specified amount of time):
```go
package main

import (
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
)

func ThingThatTakesALongTimeToCalculate(inVal int) (string, error) {
  //this does something but it's not that important
  return "Hello", nil
}

func main() {
  inVal := 200
  f := future.New(func() (interface{}, error) {
    return ThingThatTakesALongTimeToCalculate(inVal)
  })
  
  result, timeout, err := f.GetUntil(5 * time.Second)
  fmt.Println(result, timeout, err)
}
```

`timeout` will be true if the timeout was triggered. 

- The future methods `Get` and `GetUntil` can be called multiple times. 
- Once the value is calculated, the same value (and error) will be returned immediately.

Chaining usage (invoke a next method if the Promise doesn't return an error):
```go
package main

import (
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
)

func ThingThatTakesALongTimeToCalculate(inVal int) (int, error) {
	//this does something but it's not that important
	time.Sleep(5 * time.Second)
	return inVal * 2, nil
}

func main() {
	inVal := 200
	f := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	}).Then(func(i interface{}) (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(i.(int))
	})

    result, err := f.Get()
    fmt.Println(result, err)
}
```

You can use `Then` to chain together as many promises as you want.

- If an error is returned at any step along the way, the rest of the calls in the promise chain are skipped.
- Only the value and error of the last executed promise is returned; all others are lost.

Cancellation (stop waiting for a `Get` or `GetUntil` to complete):
```go
package main

import (
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
)

func ThingThatTakesALongTimeToCalculate(inVal int) (int, error) {
	//this does something but it's not that important
	time.Sleep(5 * time.Second)
	return inVal * 2, nil
}

func timeIt(f func()) int64 {
	start := time.Now()
	f()
	end := time.Now()
	dur := end.Unix() - start.Unix()
	return dur
}

func main() {
	inVal := 200
	f := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	})

	go func() {
		time.Sleep(2 * time.Second)
		f.Cancel()
	}()
	fmt.Println(timeIt(func() {
		result, err := f.Get()
		fmt.Println(result, err, f.IsCancelled())
	}))
}
```

- Calling `Cancel` after a `Get` or `GetUntil` has completed has no effect.
- Calling `Cancel` multiple times has no effect.
- When a future is cancelled, the process continues in the background but any data returned is not accessible.
- If `GetUntil` returns due to a timeout, it does not cancel the future. If you wish to cancel based on a `GetUntil` 
timeout, do the following: 
```go
	f := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	})
    val, timeout, err := f.GetUntil(2000)
    if timeout {
        f.Cancel()
    }
```

Context support:
Future contains support for the Context interface included in Go 1.7:
```go
package main

import (
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
	"context"
)

func ThingThatTakesALongTimeToCalculate(inVal int) (int, error) {
	//this does something but it's not that important
	time.Sleep(5 * time.Second)
	return inVal * 2, nil
}

func timeIt(f func()) int64 {
	start := time.Now()
	f()
	end := time.Now()
	dur := end.Unix() - start.Unix()
	return dur
}

func main() {
	inVal := 200
	ctx, _ := context.WithTimeout(context.Background(),2*time.Second)

	f := future.NewWithContext(ctx, func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	})

	fmt.Println(timeIt(func() {
		result, err := f.Get()
		fmt.Println(result, err, f.IsCancelled())
	}))
}
```

When a future is created using `NewWithContext`, it is cancelled when the Context's `Done` channel is closed, 
whether it is closed due to timeout or an explicit call to the `CancelFunc` returned by the Context factory functions.

Contributors:

We welcome your interest in Capital One’s Open Source Projects (the “Project”). Any Contributor to the project must accept and sign a CLA indicating agreement to the license terms. Except for the license granted in this CLA to Capital One and to recipients of software distributed by Capital One, you reserve all right, title, and interest in and to your contributions; this CLA does not impact your rights to use your own contributions for any other purpose.

[Link to Individual CLA](https://docs.google.com/forms/d/e/1FAIpQLSfwtl1s6KmpLhCY6CjiY8nFZshDwf_wrmNYx1ahpsNFXXmHKw/viewform)

[Link to Corporate CLA](https://docs.google.com/forms/d/e/1FAIpQLSeAbobIPLCVZD_ccgtMWBDAcN68oqbAJBQyDTSAQ1AkYuCp_g/viewform)

This project adheres to the [Open Source Code of Conduct](https://developer.capitalone.com/single/code-of-conduct/). By participating, you are expected to honor this code.
