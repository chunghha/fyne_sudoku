// main.go
package main

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- Application Version ---
const version = "v0.1.1"

// --- Custom Theme Definition ---
type myTheme struct{ fyne.Theme }

var _ fyne.Theme = (*myTheme)(nil)

func (m *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	forcedVariant := theme.VariantLight
	baseColor := theme.DefaultTheme().Color(name, forcedVariant)
	switch name {
	case theme.ColorNameForeground:
		return color.Black
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Dark Grey
	case theme.ColorNameInputBorder, theme.ColorNameInputBackground, theme.ColorNameFocus:
		return color.Transparent // Make entry visuals transparent
	default:
		return baseColor
	}
}
func (m *myTheme) Font(style fyne.TextStyle) fyne.Resource    { return theme.DefaultTheme().Font(style) }
func (m *myTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }
func (m *myTheme) Size(name fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(name) }

// --- End Custom Theme Definition ---

// Global variables
var currentPuzzle [gridSize][gridSize]int
var currentSolution [gridSize][gridSize]int
var cellWidgets [gridSize][gridSize]*sudokuCell // Holds custom cell widgets

// Define colors
var colorBlock1 = color.NRGBA{R: 230, G: 230, B: 240, A: 255}    // Bluish Light Grey
var colorBlock2 = color.NRGBA{R: 230, G: 240, B: 230, A: 255}    // Greenish Light Grey
var colorTextCorrect = color.NRGBA{R: 0, G: 150, B: 0, A: 255}   // Dark Green
var colorTextIncorrect = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Dark Red

// main function
func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&myTheme{}) // Apply Custom Theme

	myWindow := myApp.NewWindow("Sudoku Generator " + version)
	myWindow.Resize(fyne.NewSize(500, 700)) // Keep user's preferred size

	// --- Sudoku Grid UI (using custom widget) ---
	gridContainer := container.NewGridWithColumns(gridSize)
	for r := range cellWidgets {
		for c := range cellWidgets[r] {
			row, col := r, c // Capture loop variables for closure

			cell := NewSudokuCell()
			// Set initial background
			blockRow, blockCol := r/boxSize, c/boxSize
			if (blockRow+blockCol)%2 == 0 {
				cell.SetBackgroundColor(colorBlock1)
			} else {
				cell.SetBackgroundColor(colorBlock2)
			}

			// Set callback to handle focus request
			cell.onTapped = func() {
				if myWindow.Canvas() != nil {
					myWindow.Canvas().Focus(cellWidgets[row][col])
				}
			}

			cellWidgets[r][c] = cell
			gridContainer.Add(cell)
		}
	}

	// --- Control Buttons ---
	newPuzzleButton := widget.NewButton("New Puzzle (Easy)", func() { loadNewPuzzle(myApp, 35, myWindow) })
	newPuzzleMediumButton := widget.NewButton("New Puzzle (Medium)", func() { loadNewPuzzle(myApp, 45, myWindow) })
	newPuzzleHardButton := widget.NewButton("New Puzzle (Hard)", func() { loadNewPuzzle(myApp, 55, myWindow) })
	checkButton := widget.NewButton("Check Solution", func() { checkSolution(myApp, myWindow) })
	solveButton := widget.NewButton("Show Solution", func() { showSolution(myApp, myWindow) })

	buttonBox := container.NewVBox(
		newPuzzleButton, newPuzzleMediumButton, newPuzzleHardButton,
		widget.NewSeparator(),
		checkButton, solveButton,
	)

	// --- Layout ---
	content := container.NewBorder(nil, buttonBox, nil, nil, gridContainer)

	// --- Initial Load ---
	loadNewPuzzle(myApp, 45, myWindow)

	myWindow.SetContent(content)
	myWindow.SetFixedSize(true)
	myWindow.ShowAndRun()
}

// loadNewPuzzle generates a new puzzle and updates the UI
func loadNewPuzzle(a fyne.App, difficulty int, win fyne.Window) {
	currentPuzzle, currentSolution = generateSudoku(difficulty)
	updateGridUI(a, win)
}

// updateGridUI updates the custom cell widgets based on the new puzzle
func updateGridUI(a fyne.App, win fyne.Window) {
	currentTheme := a.Settings().Theme()
	defaultFgColor := currentTheme.Color(theme.ColorNameForeground, theme.VariantLight)
	defaultDisColor := currentTheme.Color(theme.ColorNameDisabled, theme.VariantLight)

	for r, rowCells := range cellWidgets {
		for c, cell := range rowCells {
			if cell == nil {
				continue
			}

			// Set background color
			blockRow, blockCol := r/boxSize, c/boxSize
			var bgColor color.Color
			if (blockRow+blockCol)%2 == 0 {
				bgColor = colorBlock1
			} else {
				bgColor = colorBlock2
			}
			cell.SetBackgroundColor(bgColor)

			if currentPuzzle[r][c] != 0 { // Given number
				cell.SetText(strconv.Itoa(currentPuzzle[r][c]))
				cell.SetStyle(fyne.TextStyle{}) // Renderer handles bold
				cell.SetDefaultTextColor(defaultDisColor)
				cell.SetTextColor(defaultDisColor)
				cell.SetSolutionValue(0)
				cell.Disable()
			} else { // Empty cell
				cell.SetText("")
				cell.SetStyle(fyne.TextStyle{}) // Renderer handles bold
				cell.SetDefaultTextColor(defaultFgColor)
				cell.SetTextColor(defaultFgColor)
				cell.SetSolutionValue(currentSolution[r][c])
				cell.Enable()
			}
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	} // Unfocus on new puzzle
}

// checkSolution checks completion and correctness based on cell text colors
func checkSolution(a fyne.App, win fyne.Window) {
	correct := true
	complete := true
	var firstErrorCell fyne.Focusable

	// Get default text color from theme (only need foreground for reset)
	currentTheme := a.Settings().Theme()
	defaultFgColor := currentTheme.Color(theme.ColorNameForeground, theme.VariantLight)
	// defaultDisColor := currentTheme.Color(theme.ColorNameDisabled, theme.VariantLight) // REMOVED - Not used in this function's logic

	// Reset TEXT colors before checking (only for enabled cells)
	for _, rowCells := range cellWidgets {
		for _, cell := range rowCells {
			if cell == nil {
				continue
			}
			if !cell.Disabled() {
				// If text is present but color is feedback color, reset it
				// Otherwise, ensure it's the default foreground
				if cell.text != "" && (cell.textColor == colorTextCorrect || cell.textColor == colorTextIncorrect) {
					// Re-evaluate based on current text
					typedValue, err := strconv.Atoi(cell.text)
					if err == nil && cell.solutionValue != 0 {
						if typedValue == cell.solutionValue {
							cell.SetTextColor(colorTextCorrect)
						} else {
							cell.SetTextColor(colorTextIncorrect)
						}
					} else {
						cell.SetTextColor(defaultFgColor)
					} // Fallback
				} else { // Includes empty cells and cells already default color
					cell.SetTextColor(defaultFgColor)
				}
			}
			// Disabled cell colors are handled by updateGridUI/showSolution
		}
	}

	// Check cells for completion and correctness (based on current text color)
	for _, rowCells := range cellWidgets {
		for _, cell := range rowCells {
			if cell == nil {
				continue
			}
			if !cell.Disabled() { // Check user-editable cells
				if cell.text == "" {
					complete = false // Found an empty cell
				} else if cell.textColor == colorTextIncorrect {
					correct = false // Found an incorrect number
					if firstErrorCell == nil {
						firstErrorCell = cell
					} // Track first error
				}
			}
		}
	}

	// Show result dialog
	if !complete {
		dialog.ShowInformation("Check Result", "The puzzle is not yet complete.", win)
	} else if !correct {
		dialog.ShowError(fmt.Errorf("incorrect solution - check red numbers"), win)
		if firstErrorCell != nil && win.Canvas() != nil {
			win.Canvas().Focus(firstErrorCell)
		}
	} else { // complete and correct
		dialog.ShowInformation("Check Result", "Congratulations! The solution is correct!", win)
		// Disable all cells on success
		for _, rowCells := range cellWidgets {
			for _, cell := range rowCells {
				if cell != nil {
					cell.Disable()
				}
			}
		}
	}
}

// showSolution reveals the solution and updates cell states
func showSolution(a fyne.App, win fyne.Window) {
	currentTheme := a.Settings().Theme()
	defaultDisColor := currentTheme.Color(theme.ColorNameDisabled, theme.VariantLight)

	for r, rowCells := range cellWidgets {
		for c, cell := range rowCells {
			if cell == nil {
				continue
			}

			// Reset background color
			blockRow, blockCol := r/boxSize, c/boxSize
			if (blockRow+blockCol)%2 == 0 {
				cell.SetBackgroundColor(colorBlock1)
			} else {
				cell.SetBackgroundColor(colorBlock2)
			}

			isUserEditable := !cell.Disabled()
			correctValueStr := strconv.Itoa(currentSolution[r][c])

			// Set text color to default disabled color
			cell.SetTextColor(defaultDisColor)

			if isUserEditable { // Only modify cells the user could edit
				cell.SetText(correctValueStr)
				cell.SetStyle(fyne.TextStyle{Italic: true}) // Renderer handles bold
			} else { // Pre-filled puzzle number
				cell.SetStyle(fyne.TextStyle{}) // Renderer handles bold
			}
			cell.Disable() // Disable all cells
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	}
}
