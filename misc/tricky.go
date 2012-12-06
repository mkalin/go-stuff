package main

import (
	"fmt"
)

func main() {
   done := make(chan bool)

   values := []string{"a", "b", "c"}
   for _, v := range values {
		u := v 
      go func(v string) {
         fmt.Println(v)
         done <- true
      }(u)
   }
	
   // wait for all goroutines to complete before exiting
   for _ = range values {
      <-done 
   }
}