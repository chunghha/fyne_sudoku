// sudoku_cell.go
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// sudokuCell represents a single interactive cell in the Sudoku grid.
type sudokuCell struct {
	widget.BaseWidget

	text      string
	style     fyne.TextStyle
	bgColor   color.Color
	textColor color.Color // Added field for text color control
	disabled  bool
	focused   bool

	// Callbacks
	onChanged func(string)
	onTapped  func()
}

// NewSudokuCell creates a new Sudoku cell widget.
func NewSudokuCell() *sudokuCell {
	cell := &sudokuCell{}
	cell.ExtendBaseWidget(cell)
	// Initialize text color (will be overridden by theme/logic later)
	cell.textColor = color.Black
	return cell
}

// --- Widget Interface ---

// CreateRenderer is required by the Widget interface.
func (c *sudokuCell) CreateRenderer() fyne.WidgetRenderer {
	// Get the current theme instance
	currentTheme := fyne.CurrentApp().Settings().Theme()
	// Use the THEME METHOD with NAME and VARIANT
	// Use widget's textColor field for initial color if needed, or theme default
	// Let's use the theme default initially, it will be set correctly by updateGridUI
	initialTextColor := currentTheme.Color(theme.ColorNameForeground, theme.VariantLight)

	text := canvas.NewText(c.text, initialTextColor) // Use initial color
	text.Alignment = fyne.TextAlignCenter
	// Default to Bold
	text.TextStyle = fyne.TextStyle{Bold: true}
	// Apply specific style if set
	if c.style.Bold || c.style.Italic || c.style.Monospace || c.style.Symbol || c.style.TabWidth > 0 {
		text.TextStyle = c.style
		// Ensure bold is always true unless explicitly overridden? Or combine?
		if c.style.Italic {
			text.TextStyle.Bold = true // Make revealed text bold italic
		} else {
			text.TextStyle.Bold = true // Ensure non-italic is bold
		}
	}

	rect := canvas.NewRectangle(c.bgColor)

	renderer := &sudokuCellRenderer{
		widget:     c,
		background: rect,
		text:       text,
		objects:    []fyne.CanvasObject{rect, text},
	}
	renderer.Refresh() // Initial refresh
	return renderer
}

// --- Interaction --- (Code remains the same)
func (c *sudokuCell) Tapped(*fyne.PointEvent) {
	if !c.disabled && c.onTapped != nil {
		c.onTapped()
	}
}
func (c *sudokuCell) FocusGained() {
	if !c.disabled {
		c.focused = true
		c.Refresh()
	}
}
func (c *sudokuCell) FocusLost() { c.focused = false; c.Refresh() }
func (c *sudokuCell) TypedRune(r rune) {
	if c.disabled || !c.focused {
		return
	}
	if r >= '1' && r <= '9' {
		c.text = string(r)
		c.Refresh()
		if c.onChanged != nil {
			c.onChanged(c.text)
		}
	}
}
func (c *sudokuCell) TypedKey(key *fyne.KeyEvent) {
	if c.disabled || !c.focused {
		return
	}
	switch key.Name {
	case fyne.KeyBackspace, fyne.KeyDelete:
		if c.text != "" {
			c.text = ""
			c.Refresh()
			if c.onChanged != nil {
				c.onChanged(c.text)
			}
		}
	}
}
func (c *sudokuCell) MouseIn(*desktop.MouseEvent)    {}
func (c *sudokuCell) MouseOut()                      {}
func (c *sudokuCell) MouseMoved(*desktop.MouseEvent) {}
func (c *sudokuCell) MouseUp(*desktop.MouseEvent)    {}
func (c *sudokuCell) MouseDown(*desktop.MouseEvent) {
	if !c.disabled && c.onTapped != nil {
		c.onTapped()
	}
}

// --- Disableable interface --- (Code remains the same)
func (c *sudokuCell) Enable()        { c.disabled = false; c.Refresh() }
func (c *sudokuCell) Disable()       { c.disabled = true; c.Refresh() }
func (c *sudokuCell) Disabled() bool { return c.disabled }

// --- Custom Setters ---
func (c *sudokuCell) SetText(text string) {
	if c.text != text {
		c.text = text
		c.Refresh()
	}
}
func (c *sudokuCell) SetStyle(style fyne.TextStyle) {
	// Ensure Bold is maintained or added correctly
	if style.Italic {
		style.Bold = true // Make italic also bold
	} else {
		style.Bold = true // Ensure non-italic is bold
	}
	if c.style != style {
		c.style = style
		c.Refresh()
	}
}
func (c *sudokuCell) SetBackgroundColor(bgColor color.Color) {
	if c.bgColor != bgColor {
		c.bgColor = bgColor
		c.Refresh()
	}
}

// Add Setter for Text Color
func (c *sudokuCell) SetTextColor(textColor color.Color) {
	if c.textColor != textColor {
		c.textColor = textColor
		c.Refresh()
	}
}

// --- Renderer ---
type sudokuCellRenderer struct {
	widget     *sudokuCell
	background *canvas.Rectangle
	text       *canvas.Text
	objects    []fyne.CanvasObject
}

func (r *sudokuCellRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.text.Resize(size)
	r.text.Move(fyne.NewPos(0, 0))
}
func (r *sudokuCellRenderer) MinSize() fyne.Size {
	minTextSize := fyne.MeasureText("8", r.text.TextSize, r.text.TextStyle)
	padding := theme.Padding() * 2
	return fyne.NewSize(minTextSize.Width+padding, minTextSize.Height+padding)
}

// Refresh updates the visual elements based on the widget's state.
func (r *sudokuCellRenderer) Refresh() {
	// *** REMOVED unused currentTheme variable ***
	// currentTheme := fyne.CurrentApp().Settings().Theme()

	r.background.FillColor = r.widget.bgColor
	r.text.Text = r.widget.text

	// Ensure Bold style is applied correctly
	currentStyle := r.widget.style
	if currentStyle.Italic {
		currentStyle.Bold = true // Italic should also be bold
	} else {
		currentStyle.Bold = true // Default to bold
	}
	r.text.TextStyle = currentStyle

	// *** Use the widget's textColor field ***
	r.text.Color = r.widget.textColor

	// Optional: Visual feedback for focus
	// if r.widget.focused { ... }

	r.background.Refresh()
	r.text.Refresh()
}
func (r *sudokuCellRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *sudokuCellRenderer) Destroy()                     {}
