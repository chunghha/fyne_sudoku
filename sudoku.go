// sudoku.go
package main

import (
	"math/rand"
)

const gridSize = 9
const boxSize = 3

// isValid checks if placing 'num' at grid[row][col] is valid
func isValid(grid *[gridSize][gridSize]int, row, col, num int) bool {
	// Check row using range
	for c, cell := range grid[row] {
		// Skip the cell we are trying to place the number in
		if c == col {
			continue
		}
		if cell == num {
			return false
		}
	}
	// Check column using standard loop (range over columns is less direct)
	for r := 0; r < gridSize; r++ {
		// Skip the cell we are trying to place the number in
		if r == row {
			continue
		}
		if grid[r][col] == num {
			return false
		}
	}

	// Check 3x3 box
	startRow := row - row%boxSize
	startCol := col - col%boxSize
	for r := startRow; r < startRow+boxSize; r++ {
		for c := startCol; c < startCol+boxSize; c++ {
			// Skip the cell we are trying to place the number in
			if r == row && c == col {
				continue
			}
			if grid[r][c] == num {
				return false
			}
		}
	}

	return true
}

// findEmpty finds the next empty cell (value 0) using range
func findEmpty(grid *[gridSize][gridSize]int) (int, int, bool) {
	for r, rowData := range grid { // Iterate over rows
		for c, cell := range rowData { // Iterate over cells in the row
			if cell == 0 {
				return r, c, true // Found empty
			}
		}
	}
	return -1, -1, false // No empty cell found
}

// solveSudoku attempts to solve the grid using backtracking
func solveSudoku(grid *[gridSize][gridSize]int) bool {
	row, col, found := findEmpty(grid) // findEmpty now uses range
	if !found {
		return true // Grid is full, solved!
	}

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })

	for _, num := range numbers {
		// Check if placing num is valid *for the current state*
		// Temporarily place the number to check validity against itself in the box/row/col
		grid[row][col] = num
		if isValid(grid, row, col, num) {
			// If it was valid, proceed with recursion
			if solveSudoku(grid) {
				return true
			}
		}
		// Backtrack: If isValid failed or solveSudoku returned false
		grid[row][col] = 0
	}

	return false // Trigger backtracking
}

// generateSudoku creates a puzzle and its solution
func generateSudoku(difficulty int) ([gridSize][gridSize]int, [gridSize][gridSize]int) {
	var solutionGrid [gridSize][gridSize]int
	solveSudoku(&solutionGrid) // solveSudoku uses range internally via findEmpty

	puzzleGrid := solutionGrid
	cellsToRemove := difficulty
	if cellsToRemove < 10 {
		cellsToRemove = 10
	}
	if cellsToRemove > 55 {
		cellsToRemove = 55
	}

	removed := 0
	maxTotalAttempts := gridSize * gridSize * 3

	// This loop picks random cells, range doesn't apply directly
	for removed < cellsToRemove && maxTotalAttempts > 0 {
		row := rand.Intn(gridSize)
		col := rand.Intn(gridSize)
		maxTotalAttempts--

		if puzzleGrid[row][col] != 0 {
			// Optional: Add uniqueness check here if desired
			puzzleGrid[row][col] = 0
			removed++
		}
	}

	return puzzleGrid, solutionGrid
}
