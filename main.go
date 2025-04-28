// main.go
package main

import (
	"fmt"
	"image/color" // Required for custom colors
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme" // Required for theme elements
	"fyne.io/fyne/v2/widget"
)

// --- Custom Theme Definition ---
type myTheme struct{ fyne.Theme }

var _ fyne.Theme = (*myTheme)(nil)

func (m *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	forcedVariant := theme.VariantLight
	baseColor := theme.DefaultTheme().Color(name, forcedVariant) // Get base color for light variant

	switch name {
	case theme.ColorNameForeground:
		return color.Black // Force black text for enabled widgets
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Force dark grey for disabled text
	// Make Input Border, Background, and Focus visuals transparent
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
var entryWidgets [gridSize][gridSize]*widget.Entry
var backgroundRects [gridSize][gridSize]*canvas.Rectangle

// Define colors
var colorBlock1 color.Color
var colorBlock2 color.Color
var colorCorrect = color.NRGBA{R: 180, G: 255, B: 180, A: 255}   // Light Green
var colorIncorrect = color.NRGBA{R: 255, G: 180, B: 180, A: 255} // Light Red

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

	// Calculate Colors based on the Applied Theme
	currentTheme := myApp.Settings().Theme()
	colorBlock1 = currentTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	colorBlock2 = offsetColor(colorBlock1, -15)

	// --- Sudoku Grid UI (using range) ---
	gridContainer := container.NewGridWithColumns(gridSize)
	for r := range entryWidgets {
		for c := range entryWidgets[r] {
			blockRow, blockCol := r/boxSize, c/boxSize
			var bgColor color.Color
			if (blockRow+blockCol)%2 == 0 {
				bgColor = colorBlock1
			} else {
				bgColor = colorBlock2
			}

			rect := canvas.NewRectangle(bgColor)
			backgroundRects[r][c] = rect

			entry := widget.NewEntry()
			entry.Validator = validation.NewRegexp(`^[1-9]?$`, "Must be 1-9 or empty")
			entry.PlaceHolder = ""
			// entry.TextAlign = fyne.TextAlignCenter // Keep commented out to avoid build errors
			entry.TextStyle = fyne.TextStyle{}

			entryWidgets[r][c] = entry

			// *** USE container.NewCenter as requested ***
			centeredEntry := container.NewCenter(entry)

			// Stack the background and the centered entry
			cellWidget := container.NewStack(rect, centeredEntry)
			gridContainer.Add(cellWidget)
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

// updateGridUI - Resets background colors and validation state
func updateGridUI(a fyne.App, win fyne.Window) {
	currentTheme := a.Settings().Theme()
	colorBlock1 = currentTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	colorBlock2 = offsetColor(colorBlock1, -15)

	for r, rowWidgets := range entryWidgets {
		for c, entry := range rowWidgets {
			rect := backgroundRects[r][c]
			if entry == nil || rect == nil {
				continue
			}

			// Reset background color
			blockRow, blockCol := r/boxSize, c/boxSize
			var bgColor color.Color
			if (blockRow+blockCol)%2 == 0 {
				bgColor = colorBlock1
			} else {
				bgColor = colorBlock2
			}
			rect.FillColor = bgColor
			rect.Refresh()

			// Clear any previous validation state (icon/outline)
			_ = entry.Validate()

			if currentPuzzle[r][c] != 0 { // Given number
				entry.SetText(strconv.Itoa(currentPuzzle[r][c]))
				entry.TextStyle = fyne.TextStyle{Bold: true}
				entry.Disable()
			} else { // Empty cell
				entry.SetText("")
				entry.TextStyle = fyne.TextStyle{}
				entry.Enable()
			}
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	}
}

// checkSolution - Applies color feedback ONLY via background
func checkSolution(a fyne.App, win fyne.Window) {
	correct := true
	complete := true
	var firstErrorEntry fyne.Focusable

	// Recalculate base colors needed for reset
	currentTheme := a.Settings().Theme()
	colorBlock1 = currentTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	colorBlock2 = offsetColor(colorBlock1, -15)

	// Reset all background colors and clear entry validation icons before checking
	for r, rowRects := range backgroundRects {
		for c, rect := range rowRects {
			if rect == nil {
				continue
			}
			// Reset background
			blockRow, blockCol := r/boxSize, c/boxSize
			if (blockRow+blockCol)%2 == 0 {
				rect.FillColor = colorBlock1
			} else {
				rect.FillColor = colorBlock2
			}
			rect.Refresh()
			// Clear entry's visual validation state
			if entryWidgets[r][c] != nil {
				_ = entryWidgets[r][c].Validate()
			}
		}
	}

	// Check entries and apply feedback colors to background
	for r, rowWidgets := range entryWidgets {
		for c, entry := range rowWidgets {
			rect := backgroundRects[r][c] // Get corresponding rect
			if entry == nil || rect == nil {
				continue
			}
			// DO NOT call entry.Validate() here - we use background color instead

			if !entry.Disabled() { // Check user-editable cells
				valStr := entry.Text
				if valStr == "" {
					complete = false
					// Keep default background for empty cells (already reset above)
					continue
				}

				val, err := strconv.Atoi(valStr)
				if err != nil || val < 1 || val > 9 {
					// Invalid input
					rect.FillColor = colorIncorrect // Set background to Red
					rect.Refresh()
					if firstErrorEntry == nil {
						firstErrorEntry = entry
					}
					correct, complete = false, false
					continue
				}

				// Valid input, check against solution
				if val != currentSolution[r][c] {
					// Incorrect number
					rect.FillColor = colorIncorrect // Set background to Red
					rect.Refresh()
					if firstErrorEntry == nil {
						firstErrorEntry = entry
					}
					correct = false
				} else {
					// Correct number
					rect.FillColor = colorCorrect // Set background to Green
					rect.Refresh()
				}
			}
		}
	}

	// Show result dialog
	if !complete && correct {
		dialog.ShowInformation("Check Result", "The puzzle is not yet complete.", win)
	} else if !correct {
		dialog.ShowError(fmt.Errorf("incorrect solution - check colored cells"), win)
		if firstErrorEntry != nil && win.Canvas() != nil {
			win.Canvas().Focus(firstErrorEntry)
		}
	} else { // complete and correct
		dialog.ShowInformation("Check Result", "Congratulations! The solution is correct!", win)
		// Optionally disable all on success
		for _, rowWidgets := range entryWidgets {
			for _, entry := range rowWidgets {
				if entry != nil {
					entry.Disable()
				}
			}
		}
	}
}

// showSolution - Resets background colors and validation state
func showSolution(a fyne.App, win fyne.Window) {
	currentTheme := a.Settings().Theme()
	colorBlock1 = currentTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	colorBlock2 = offsetColor(colorBlock1, -15)

	for r, rowWidgets := range entryWidgets {
		for c, entry := range rowWidgets {
			rect := backgroundRects[r][c]
			if entry == nil || rect == nil {
				continue
			}

			// Reset background color
			blockRow, blockCol := r/boxSize, c/boxSize
			if (blockRow+blockCol)%2 == 0 {
				rect.FillColor = colorBlock1
			} else {
				rect.FillColor = colorBlock2
			}
			rect.Refresh()

			// Clear any previous validation state (icon/outline)
			_ = entry.Validate()

			isUserEditable := !entry.Disabled()
			userValue := entry.Text
			correctValueStr := strconv.Itoa(currentSolution[r][c])

			if isUserEditable && userValue != correctValueStr { // Revealed number
				entry.SetText(correctValueStr)
				entry.TextStyle = fyne.TextStyle{Italic: true}
				entry.Disable()
			} else if !isUserEditable { // Pre-filled puzzle number
				entry.TextStyle = fyne.TextStyle{Bold: true}
				entry.Disable()
			} else { // User entered correct number
				entry.TextStyle = fyne.TextStyle{}
				entry.Disable()
			}
		}
	}
	if win.Canvas() != nil {
		win.Canvas().Focus(nil)
	}
}
