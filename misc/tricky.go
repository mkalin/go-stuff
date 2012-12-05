package main

import (
	"fmt"
)

func main() {
   done := make(chan bool)

   values := []string{"a", "b", "c"}
   for _, v := range values {
		v := v // necessary to get all three
      go func() {
         fmt.Println(v)
         done <- true
      }()
   }
	
   // wait for all goroutines to complete before exiting
   for _ = range values {
      <-done 
   }
}