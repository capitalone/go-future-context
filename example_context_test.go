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

package future_test

import (
	"context"
	"fmt"
	future "github.com/capitalone/go-future-context"
	"time"
)

func Example_context() {
	inVal := 200

	ThingThatTakesALongTimeToCalculate := func(inVal int) (int, error) {
		//this does something but it's not that important
		time.Sleep(5 * time.Second)
		return inVal * 2, nil
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	f := future.NewWithContext(ctx, func() (interface{}, error) {
		return ThingThatTakesALongTimeToCalculate(inVal)
	})

	result, err := f.Get()
	fmt.Println(result, err, f.IsCancelled())
}
