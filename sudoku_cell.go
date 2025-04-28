// sudoku_cell.go
package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Define font size increase relative to theme default
const fontSizeIncrease float32 = 4.0

// sudokuCell represents a single interactive cell in the Sudoku grid.
type sudokuCell struct {
	widget.BaseWidget

	text             string
	style            fyne.TextStyle
	bgColor          color.Color
	textColor        color.Color // Current text color (can be feedback color)
	defaultTextColor color.Color // Color when not showing feedback
	solutionValue    int         // Correct value for this cell
	disabled         bool
	focused          bool

	onChanged func(string)
	onTapped  func()
}

// NewSudokuCell creates a new Sudoku cell widget.
func NewSudokuCell() *sudokuCell {
	cell := &sudokuCell{}
	cell.ExtendBaseWidget(cell)
	cell.textColor = color.Black // Initial default
	cell.defaultTextColor = color.Black
	return cell
}

// --- Widget Interface ---

func (c *sudokuCell) CreateRenderer() fyne.WidgetRenderer {
	currentTheme := fyne.CurrentApp().Settings().Theme()
	initialTextColor := currentTheme.Color(theme.ColorNameForeground, theme.VariantLight)

	text := canvas.NewText(c.text, initialTextColor)
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = theme.TextSize() + fontSizeIncrease // Increased font size
	text.TextStyle = fyne.TextStyle{Bold: true}         // Always bold

	rect := canvas.NewRectangle(c.bgColor)

	renderer := &sudokuCellRenderer{
		widget:     c,
		background: rect,
		text:       text,
		objects:    []fyne.CanvasObject{rect, text},
	}
	renderer.Refresh()
	return renderer
}

// --- Interaction ---
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
		typedValue, err := strconv.Atoi(c.text)
		if err == nil && c.solutionValue != 0 {
			if typedValue == c.solutionValue {
				c.SetTextColor(colorTextCorrect)
			} else {
				c.SetTextColor(colorTextIncorrect)
			}
		} else {
			c.SetTextColor(c.defaultTextColor)
		} // Fallback
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
			c.SetTextColor(c.defaultTextColor) // Reset color
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

// --- Disableable interface ---
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
	style.Bold = true // Always bold
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
func (c *sudokuCell) SetTextColor(textColor color.Color) {
	if c.textColor != textColor {
		c.textColor = textColor
		c.Refresh()
	}
}
func (c *sudokuCell) SetDefaultTextColor(textColor color.Color) { c.defaultTextColor = textColor }
func (c *sudokuCell) SetSolutionValue(value int)                { c.solutionValue = value }

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
func (r *sudokuCellRenderer) Refresh() {
	r.background.FillColor = r.widget.bgColor
	r.text.Text = r.widget.text
	r.text.TextSize = theme.TextSize() + fontSizeIncrease // Ensure size is correct
	r.text.TextStyle = fyne.TextStyle{Bold: true}         // Always bold
	if r.widget.style.Italic {                            // Apply italic if set
		r.text.TextStyle.Italic = true
	}
	r.text.Color = r.widget.textColor // Use the stored text color

	r.background.Refresh()
	r.text.Refresh()
}
func (r *sudokuCellRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *sudokuCellRenderer) Destroy()                     {}
