// sudoku.go
package main

import (
	"math/rand"
	// NOTE: Use of built-in 'max' and 'min' requires Go 1.21 or later.
)

const gridSize = 9
const boxSize = 3

// isValid checks if placing 'num' at grid[row][col] is valid.
func isValid(grid *[gridSize][gridSize]int, row, col, num int) bool {
	// Check row using range
	for c, cell := range grid[row] {
		if c != col && cell == num {
			return false
		}
	}
	// Check column using range
	for r := range grid {
		if r != row && grid[r][col] == num {
			return false
		}
	}
	// Check 3x3 box
	startRow := row - row%boxSize
	startCol := col - col%boxSize
	for r := startRow; r < startRow+boxSize; r++ {
		for c := startCol; c < startCol+boxSize; c++ {
			if r != row || c != col {
				if grid[r][c] == num {
					return false
				}
			}
		}
	}
	return true
}

// findEmpty finds the next empty cell (value 0) using range.
func findEmpty(grid *[gridSize][gridSize]int) (int, int, bool) {
	for r, rowData := range grid {
		for c, cell := range rowData {
			if cell == 0 {
				return r, c, true
			}
		}
	}
	return -1, -1, false
}

// solveSudoku attempts to solve the grid using backtracking.
func solveSudoku(grid *[gridSize][gridSize]int) bool {
	row, col, found := findEmpty(grid)
	if !found {
		return true
	}

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })

	for _, num := range numbers {
		grid[row][col] = num
		if isValid(grid, row, col, num) {
			if solveSudoku(grid) {
				return true
			}
		}
		grid[row][col] = 0 // Backtrack
	}
	return false
}

// generateSudoku creates a puzzle and its solution.
func generateSudoku(difficulty int) ([gridSize][gridSize]int, [gridSize][gridSize]int) {
	var solutionGrid [gridSize][gridSize]int
	solveSudoku(&solutionGrid) // Generate a full solution

	puzzleGrid := solutionGrid // Start with the solution
	cellsToRemove := difficulty

	// Use max to ensure minimum difficulty (Requires Go 1.21+)
	cellsToRemove = max(cellsToRemove, 10)

	// *** Use min to cap maximum difficulty (Requires Go 1.21+) ***
	cellsToRemove = min(cellsToRemove, 55)

	removed := 0
	maxTotalAttempts := gridSize * gridSize * 3 // Limit attempts

	for removed < cellsToRemove && maxTotalAttempts > 0 {
		row := rand.Intn(gridSize)
		col := rand.Intn(gridSize)
		maxTotalAttempts--

		if puzzleGrid[row][col] != 0 {
			puzzleGrid[row][col] = 0
			removed++
		}
	}
	return puzzleGrid, solutionGrid
}
