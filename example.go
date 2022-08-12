package stash

import (
	"encoding/json"
	"math/rand"
)

//this ensures that example is cacheable
var _ Cacheable = &Example{}

//Example is a type that can be used for simple tests or to
// understand how to make a struct Cacheable
//KIM: although we use JSON for simple serialization
// you can use whatever works like YAML, protobuf etc.
type Example struct {
	Int    int     `json:"int,omitempty"`
	Float  float64 `json:"float,omitempty"`
	String string  `json:"string,omitempty"`
}

func (e *Example) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Example) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, e)
}

//ExampleGenFloat64 will generate a random number of random float values if n is equal to 0
// not to exceed the constant TestMaxExamples, if n is provided, it will generate that many items
func ExampleGenFloat64(n int) []*Example {
	if n <= 0 {
		n = int(rand.Float64() * 1000)
	}
	values := make([]*Example, 0, n)
	for i := 0; i < n; i++ {
		values = append(values, &Example{Float: rand.Float64()})
	}
	return values
}
