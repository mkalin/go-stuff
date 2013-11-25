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
	row  int
	col  int
	ant *Ant
}
var channels []chan *Msg

type Ant struct {
	id    rune
	moves int     // how many times an Ant has moved overall
	stays int     // how many times an Ant has stayed in its current cell
	cell  *Cell   // current cell
	pipe  chan *Msg
}

type Cell struct {
	count int // how many times an ant has moved into this cell
	row   int
	col   int
	ant  *Ant
}

const dim int = 8
const border string = " "

type Matrix [dim][dim]Cell
var matrix *Matrix

func displayBoard() {
	display := " * "
	fmt.Println()
	for i := 0; i < dim; i++ {
		fmt.Print("\t")
		for j := 0; j < dim; j++ {
			if matrix[i][j].ant != nil {
				display = fmt.Sprintf(" %c ", matrix[i][j].ant.id)
			}
			fmt.Print(border + display + border)
		}
		fmt.Println()
	}
	fmt.Println()
}

func targetRC(row int, col int) (int, int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ind := rng.Intn(len(dirs))
	pair := dirs[ind]
	r, c := pair[0] + row, pair[1] + col
	return ((r % dim) + dim) % dim, ((c % dim) + dim) % dim
}

func updateBoard() {
	for _, channel := range channels {
		select {
		case msg, ok := <-channel:
			if ok { 
				if matrix[msg.row][msg.col].ant != nil { // already occupied?
					msg.ant.stays++
				} else {
					msg.ant.cell = &matrix[msg.row][msg.col]
					msg.ant.cell.ant = msg.ant
					msg.ant.cell.count++
					msg.ant.moves++
				}
			}
		default:
		}
		displayBoard()
	}
}

func randomStep(cell *Cell) {
	for {
		if stop { break }

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		r, c := targetRC(cell.row, cell.col)
		msg := &Msg{row: r, 
                  col: c, 
                  ant: cell.ant}
		cell.ant.pipe<- msg
		zzz := rng.Intn(30)
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
	
	// Randomly populate the matrix with at most N ants.
	// (It's possible but unlikely that, during initialization, 
   // one ant might displace another in a cell.)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idC := 'a'
	channels = []chan *Msg{}
	for i := 0; i < n; i++ {
		row := r.Intn(dim)
		col := r.Intn(dim)
		ant := &Ant { 
			id:    idC,
			moves: 0,
			stays: 0,
			cell:  &matrix[row][col],
		   pipe:  make(chan *Msg)}
		matrix[row][col].ant = ant
		idC++
		channels = append(channels, ant.pipe)
	}
}

func main() {
	initialize(dim + dim + 1)
	//simulate() // end with control-C

	r, c := targetRC(0, 0)
	fmt.Println(fmt.Sprintf("%v %v ==> %v %v", 0, 0, r, c))

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM) // control-C
	log.Println(<-ch)
	log.Println("Gracefully shutting down...")

	time.Sleep(time.Duration(1) * time.Second)
	os.Exit(0) // kill all goroutines
}