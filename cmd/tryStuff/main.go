package main

import "fmt"

type someType struct {
	Key int
}

func (instance *someType) increment() {
	old := instance.Key
	instance.Key = old + 1
}

func main() {
	someInstance := &someType{Key: 1}

	someInstance.increment()

	fmt.Printf("Value %d\n", someInstance.Key)
}
