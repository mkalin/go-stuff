package main

import (
	"fmt"
)

func fibS(c, quit chan int) {
	x, y := 1, 1

	for {
		// select blocks until a case is ready
		select {
		case c <- x:
			x, y = y, x + y
		case  <- quit:
			fmt.Println("quit")
			return
		}
	}
}

func fib(n int, c chan int) {
	defer close(c)
	x, y := 1, 1
	for i := 0; i < n; i++ {
		c <- x
		x, y = y, x + y
	}
}

func buff(n int) {
	c := make(chan int, n)
	for i := 0; i < n; i++ {
		c <- i + 1
	}
	for i := 0; i < n; i++ {
		fmt.Println(<- c)
	}
}

func sum(a []int, c chan int) {
	sum := 0
	for i := 0; i < len(a); i++ {
		sum += a[i]
	}
	c <- sum
}

func main() {
	a := []int{11, 5, 8, -9, 7, -4, 3, -1, -14, 23}
	l := len(a) / 2
	c := make(chan int)

	go sum(a[:l], c)
	go sum(a[l:], c)
	// unbuffered channels block automatically
	s1, s2 := <- c, <- c 
	close(c)
	fmt.Println(s1, s2, s1 + s2)

	// buffered channels
	buff(2)

	fmt.Println()

	// richer example--buffered channel
	c = make(chan int, 10)
	go fib(cap(c), c) 

	// Read the channel values
	// range channel == go until channel closes
	for i := range c {
		fmt.Println(i)
	}

	fmt.Println()

	// Control fibS from main
	quit := make(chan int)
	cS := make(chan int)
	go fibS(cS, quit) // nothing happens until data are sent
	for i := 0; i < 10; i++ {
		fmt.Println(<- cS)		
	}
	quit <- 0 // any integer value would do

	close(cS)
	close(quit)
}