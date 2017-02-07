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
//See the License for the specific language governing permissions and limitations under the License. 

package future_test

import (
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
)

func Example_then() {
	inVal := 200

	ThingThatTakesALongTimeToCalculate := func(inVal int) (int, error) {
		//this does something but it's not that important
		time.Sleep(5 * time.Second)
		return inVal * 2, nil
	}

	f := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	}).Then(func(i interface{}) (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(i.(int))
	})

	result, err := f.Get()
	fmt.Println(result, err)

	//  results are instantaneous
	result, err = f.Get()
	fmt.Println(result, err)

	f2 := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	}).Then(func(i interface{}) (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(i.(int))
	})

	// times out during first step
	result, timeout, err := f2.GetUntil(2000)
	fmt.Println(result, timeout, err)

	f3 := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	}).Then(func(i interface{}) (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(i.(int))
	})

	// times out during second step
	result, timeout, err = f3.GetUntil(7000)
	fmt.Println(result, timeout, err)

	f4 := future.New(func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	}).Then(func(i interface{}) (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(i.(int))
	})

	// completes both steps
	result, timeout, err = f4.GetUntil(20000)
	fmt.Println(result, timeout, err)
}
