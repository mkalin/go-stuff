package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 8 directional moves ({n, m} is {row, col}: 0 is "same", 1 and -1 are "move one", either left/right or up/down
// The matrix is treated as a torus so that moves from any cell are possible.
var dirs = [][]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
var stop bool 

type Msg struct {
	row int   // new row
	col int   // new column
	ant *Ant  // encapsulates current row and column
}
var channels []chan *Msg

type Ant struct {
	id    rune
	moves int        // how many times an Ant has moved overall
	stays int        // how many times an Ant has stayed in its current cell
   cell  *Cell
	pipe  chan *Msg
}
var ants []*Ant

type Cell struct {
	count    int     // how many times an ant has moved into this cell
	row      int     // fixed 
	col      int     // fixed
	ant      *Ant
}

const dim int = 8
const border string = " "
const maxPause int = 32 // milliseconds

type Matrix [dim][dim]Cell
var matrix *Matrix

var rng *rand.Rand

func (ant *Ant) String() string {
	return fmt.Sprintf("Id: %c\tMoves: %5d   Stays: %5d   Cell: %v, %v", 
		ant.id, ant.moves, ant.stays, ant.cell.row, ant.cell.col)
}

func targetRC(row int, col int) (int, int) {
	pair := dirs[rng.Intn(len(dirs))]
	r, c := pair[0] + row, pair[1] + col
	return ((r % dim) + dim) % dim, ((c % dim) + dim) % dim
}

func updateBoard() {
	for {
		for _, channel := range channels {
			msg := <-channel
			if matrix[msg.row][msg.col].ant != nil { 
				msg.ant.stays++
			} else {
				matrix[msg.row][msg.col].ant = msg.ant
				matrix[msg.ant.cell.row][msg.ant.cell.col].ant = nil
				msg.ant.cell = &matrix[msg.row][msg.col]
				msg.ant.cell.count++
				msg.ant.moves++
			}
		}
	}
}

func randomStep(ant *Ant) {
	for {
		if stop { break }

		r, c := targetRC(ant.cell.row, ant.cell.col)
		msg := &Msg{row: r,
                  col: c,
                  ant: ant}
		ant.pipe <-msg
		zzz := rng.Intn(maxPause)
		time.Sleep(time.Duration(zzz) * time.Millisecond)
	}
}

func initialize(n int) {
	stop = false

	matrix = new(Matrix)
	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			matrix[i][j].row = i
			matrix[i][j].col = j
		}
	}
	
	// Randomly populate the matrix with N ants.
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	idC := 'a'
	channels = []chan *Msg{}
	ants = []*Ant{}
	i := 0

	for {
		if (i >= n) { break }

		row := rng.Intn(dim)
		col := rng.Intn(dim)
		if (matrix[row][col].ant == nil) {
			ant := &Ant { 
				id:    idC,
				moves: 0,
				stays: 0,
				cell:  &matrix[row][col],
		      pipe:  make(chan *Msg, 1)}
			matrix[row][col].ant = ant
			idC++
			channels = append(channels, ant.pipe)
			ants = append(ants, ant)
			i++
		}
	}
}

func dumpAnts() {
	fmt.Println("\nAnts:")
	for _, ant := range ants {
		fmt.Println(ant.String())
	}
}

func main() {
	initialize(dim + dim + 1)
	dumpAnts()

	go updateBoard()
	for _, ant := range ants {
		go randomStep(ant)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM) // control-C
	log.Println(<-ch)
	log.Println("Gracefully shutting down...")
	time.Sleep(time.Duration(1) * time.Second)

	dumpAnts()
	os.Exit(0) // kill all goroutines
}