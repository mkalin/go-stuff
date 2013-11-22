package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 8 directional moves ({n, m} is {row, col}: 0 is "same", 1 and -1 are "move one")
var dirs = [][]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}

type Ant struct {
	id    rune
	moves int     // how many times an Ant has moved overall
	stays int     // how many times an Ant has stayed in its current cell
	cell  *Cell   // current cell
}

func (ant *Ant) Id() rune {
	return ant.id
}

type Cell struct {
	count    int // how many times an ant has moved into this cell
	occupant *Ant
}

const dim int = 8
const border string = " "

type Matrix [dim][dim]Cell

var matrix *Matrix

func displayBoard() {
	var display string
	fmt.Println()
	for i := 0; i < dim; i++ {
		fmt.Print("\t")
		for j := 0; j < dim; j++ {
			if matrix[i][j].occupant == nil {
				display = " * " 
			} else {
				display = fmt.Sprintf(" %c ", matrix[i][j].occupant.id)
			}
			fmt.Print(border + display + border)
		}
		fmt.Println()
	}
	fmt.Println()
}

func initialize(n int) {
	matrix = new(Matrix)
	
	// Randomly populate the matrix with N ants.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idC := 'a'
	for i := 0; i < n; i++ {
		row := r.Intn(dim)
		col := r.Intn(dim)
		ant := &Ant { 
			id:    idC,
			moves: 0,
			stays: 0,
		   cell:  &matrix[row][col]}
		matrix[row][col].occupant = ant
		idC++
	}
}

func main() {
	initialize((dim * dim) / 4)
	displayBoard()
}