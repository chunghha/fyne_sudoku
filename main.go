// main.go
package main

import (
	"fmt"
	"image/color" // Required for custom colors
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/data/validation" // No longer needed
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme" // Required for theme elements
	"fyne.io/fyne/v2/widget"
)

// --- Custom Theme Definition (Simplified) ---
type myTheme struct{ fyne.Theme }

var _ fyne.Theme = (*myTheme)(nil)

func (m *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	forcedVariant := theme.VariantLight
	baseColor := theme.DefaultTheme().Color(name, forcedVariant) // Get base color for light variant

	switch name {
	case theme.ColorNameForeground:
		return color.Black // Force black text for enabled widgets
	case theme.ColorNameDisabled:
		// Use a visible color for disabled text in our light theme
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Dark Grey
	// Keep Input Border, Background, and Focus visuals transparent
	case theme.ColorNameInputBorder, theme.ColorNameInputBackground, theme.ColorNameFocus:
		return color.Transparent
	default:
		// Use the standard light theme color otherwise
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
var baseGray = color.NRGBA{R: 235, G: 235, B: 235, A: 255} // Custom light grey
var colorBlock1 color.Color                                // Will be set to baseGray
var colorBlock2 color.Color                                // Will be offset from baseGray
// Define colors for TEXT feedback
var colorTextCorrect = color.NRGBA{R: 0, G: 150, B: 0, A: 255}   // Dark Green
var colorTextIncorrect = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Dark Red

// Helper function to slightly adjust a color
func offsetColor(c color.Color, offset int) color.Color {
	nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
	clamp := func(val int) uint8 {
		if val < 0 {
			return 0
		}
		if val > 255 {
			return 255
		}
		return uint8(val)
	}
	r, g, b := int(nrgba.R)+offset, int(nrgba.G)+offset, int(nrgba.B)+offset
	return color.NRGBA{R: clamp(r), G: clamp(g), B: clamp(b), A: nrgba.A}
}

// main function
func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&myTheme{}) // Apply Custom Theme

	myWindow := myApp.NewWindow("Sudoku Generator")
	myWindow.Resize(fyne.NewSize(500, 700)) // Keep user's preferred size

	// Calculate Colors
	colorBlock1 = baseGray
	colorBlock2 = offsetColor(colorBlock1, -20) // Adjust offset as needed for contrast

	// --- Sudoku Grid UI (using custom widget) ---
	gridContainer := container.NewGridWithColumns(gridSize)
	for r := range cellWidgets {
		for c := range cellWidgets[r] {
			// Need indices for closure
			row, col := r, c

			cell := NewSudokuCell()
			// Set initial background (will be updated in updateGridUI)
			blockRow, blockCol := r/boxSize, c/boxSize
			if (blockRow+blockCol)%2 == 0 {
				cell.SetBackgroundColor(colorBlock1)
			} else {
				cell.SetBackgroundColor(colorBlock2)
			}

			// Set callback to handle focus request
			cell.onTapped = func() {
				if myWindow.Canvas() != nil {
					myWindow.Canvas().Focus(cellWidgets[row][col]) // Focus the tapped cell
				}
			}

			cellWidgets[r][c] = cell // Store the custom cell widget
			gridContainer.Add(cell)  // Add the cell directly
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
	myWindow.SetFixedSize(true) // Keep fixed size unless user wants resize
	myWindow.ShowAndRun()
}

// loadNewPuzzle
func loadNewPuzzle(a fyne.App, difficulty int, win fyne.Window) {
	currentPuzzle, currentSolution = generateSudoku(difficulty)
	updateGridUI(a, win)
}

// updateGridUI - Updates custom cell widgets, sets default text color
func updateGridUI(a fyne.App, win fyne.Window) {
	currentTheme := a.Settings().Theme()
	colorBlock1 = baseGray
	colorBlock2 = offsetColor(colorBlock1, -20)
	// Get default text colors from theme
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
				cell.SetStyle(fyne.TextStyle{})    // Let renderer handle bold
				cell.SetTextColor(defaultDisColor) // Set disabled text color
				cell.Disable()
			} else { // Empty cell
				cell.SetText("")
				cell.SetStyle(fyne.TextStyle{})   // Let renderer handle bold
				cell.SetTextColor(defaultFgColor) // Set default foreground text color
				cell.Enable()
			}
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	} // Unfocus on new puzzle
}

// checkSolution - Applies TEXT color feedback
func checkSolution(a fyne.App, win fyne.Window) {
	correct := true
	complete := true
	var firstErrorCell fyne.Focusable

	// Get default text colors from theme
	currentTheme := a.Settings().Theme()
	defaultFgColor := currentTheme.Color(theme.ColorNameForeground, theme.VariantLight)
	defaultDisColor := currentTheme.Color(theme.ColorNameDisabled, theme.VariantLight)

	// Reset TEXT colors before checking
	// *** CORRECTED LOOP: Use blank identifiers ***
	for _, rowCells := range cellWidgets {
		for _, cell := range rowCells {
			if cell == nil {
				continue
			}
			// Reset text color based on disabled state
			if cell.Disabled() {
				cell.SetTextColor(defaultDisColor)
			} else {
				cell.SetTextColor(defaultFgColor)
			}
			// Background reset is handled by updateGridUI or showSolution
		}
	}

	// Check cells and apply feedback TEXT colors
	for r, rowCells := range cellWidgets {
		for c, cell := range rowCells {
			if cell == nil {
				continue
			}

			if !cell.Disabled() { // Check user-editable cells
				valStr := cell.text // Access internal text directly for check
				if valStr == "" {
					complete = false
					cell.SetTextColor(defaultFgColor) // Ensure empty cells have default color
					continue
				}

				val, err := strconv.Atoi(valStr)
				// Basic validation
				if err != nil || val < 1 || val > 9 {
					// Invalid input
					cell.SetTextColor(colorTextIncorrect) // Set text to Red
					if firstErrorCell == nil {
						firstErrorCell = cell
					}
					correct, complete = false, false
					continue
				}

				// Valid input, check against solution
				if val != currentSolution[r][c] {
					// Incorrect number
					cell.SetTextColor(colorTextIncorrect) // Set text to Red
					if firstErrorCell == nil {
						firstErrorCell = cell
					}
					correct = false
				} else {
					// Correct number
					cell.SetTextColor(colorTextCorrect) // Set text to Green
				}
			}
		}
	}

	// Show result dialog
	if !complete && correct {
		dialog.ShowInformation("Check Result", "The puzzle is not yet complete.", win)
	} else if !correct {
		dialog.ShowError(fmt.Errorf("incorrect solution - check colored text"), win) // Updated message
		if firstErrorCell != nil && win.Canvas() != nil {
			win.Canvas().Focus(firstErrorCell)
		}
	} else { // complete and correct
		dialog.ShowInformation("Check Result", "Congratulations! The solution is correct!", win)
		// Optionally disable all on success
		for _, rowCells := range cellWidgets {
			for _, cell := range rowCells {
				if cell != nil {
					cell.Disable()
				}
			}
		}
	}
}

// showSolution - Updates custom cells, sets text color
func showSolution(a fyne.App, win fyne.Window) {
	// Recalculate colors
	colorBlock1 = baseGray
	colorBlock2 = offsetColor(colorBlock1, -20)
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
			userValue := cell.text // Access internal text
			correctValueStr := strconv.Itoa(currentSolution[r][c])

			// Set default disabled color first
			cell.SetTextColor(defaultDisColor)

			if isUserEditable && userValue != correctValueStr { // Revealed number
				cell.SetText(correctValueStr)
				cell.SetStyle(fyne.TextStyle{Italic: true}) // Renderer handles bold
				cell.Disable()
			} else if !isUserEditable { // Pre-filled puzzle number
				cell.SetStyle(fyne.TextStyle{}) // Renderer handles bold
				cell.Disable()                  // Ensure disabled
			} else { // User entered correct number
				cell.SetStyle(fyne.TextStyle{}) // Renderer handles bold
				cell.Disable()
			}
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	}
}
