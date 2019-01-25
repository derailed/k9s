package tview

import (
	"math"

	"github.com/gdamore/tcell"
)

// gridItem represents one primitive and its possible position on a grid.
type gridItem struct {
	Item                        Primitive // The item to be positioned. May be nil for an empty item.
	Row, Column                 int       // The top-left grid cell where the item is placed.
	Width, Height               int       // The number of rows and columns the item occupies.
	MinGridWidth, MinGridHeight int       // The minimum grid width/height for which this item is visible.
	Focus                       bool      // Whether or not this item attracts the layout's focus.

	visible    bool // Whether or not this item was visible the last time the grid was drawn.
	x, y, w, h int  // The last position of the item relative to the top-left corner of the grid. Undefined if visible is false.
}

// Grid is an implementation of a grid-based layout. It works by defining the
// size of the rows and columns, then placing primitives into the grid.
//
// Some settings can lead to the grid exceeding its available space. SetOffset()
// can then be used to scroll in steps of rows and columns. These offset values
// can also be controlled with the arrow keys (or the "g","G", "j", "k", "h",
// and "l" keys) while the grid has focus and none of its contained primitives
// do.
//
// See https://github.com/rivo/tview/wiki/Grid for an example.
type Grid struct {
	*Box

	// The items to be positioned.
	items []*gridItem

	// The definition of the rows and columns of the grid. See
	// SetRows()/SetColumns() for details.
	rows, columns []int

	// The minimum sizes for rows and columns.
	minWidth, minHeight int

	// The size of the gaps between neighboring primitives. This is automatically
	// set to 1 if borders is true.
	gapRows, gapColumns int

	// The number of rows and columns skipped before drawing the top-left corner
	// of the grid.
	rowOffset, columnOffset int

	// Whether or not borders are drawn around grid items. If this is set to true,
	// a gap size of 1 is automatically assumed (which is filled with the border
	// graphics).
	borders bool

	// The color of the borders around grid items.
	bordersColor tcell.Color
}

// NewGrid returns a new grid-based layout container with no initial primitives.
//
// Note that Box, the superclass of Grid, will have its background color set to
// transparent so that any grid areas not covered by any primitives will leave
// their background unchanged. To clear a Grid's background before any items are
// drawn, set it to the desired color:
//
//   grid.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
func NewGrid() *Grid {
	g := &Grid{
		Box:          NewBox().SetBackgroundColor(tcell.ColorDefault),
		bordersColor: Styles.GraphicsColor,
	}
	g.focus = g
	return g
}

// SetRows defines how the rows of the grid are distributed. Each value defines
// the size of one row, starting with the leftmost row. Values greater 0
// represent absolute row widths (gaps not included). Values less or equal 0
// represent proportional row widths or fractions of the remaining free space,
// where 0 is treated the same as -1. That is, a row with a value of -3 will
// have three times the width of a row with a value of -1 (or 0). The minimum
// width set with SetMinSize() is always observed.
//
// Primitives may extend beyond the rows defined explicitly with this function.
// A value of 0 is assumed for any undefined row. In fact, if you never call
// this function, all rows occupied by primitives will have the same width.
// On the other hand, unoccupied rows defined with this function will always
// take their place.
//
// Assuming a total width of the grid of 100 cells and a minimum width of 0, the
// following call will result in rows with widths of 30, 10, 15, 15, and 30
// cells:
//
//   grid.SetRows(30, 10, -1, -1, -2)
//
// If a primitive were then placed in the 6th and 7th row, the resulting widths
// would be: 30, 10, 10, 10, 20, 10, and 10 cells.
//
// If you then called SetMinSize() as follows:
//
//   grid.SetMinSize(15, 20)
//
// The resulting widths would be: 30, 15, 15, 15, 20, 15, and 15 cells, a total
// of 125 cells, 25 cells wider than the available grid width.
func (g *Grid) SetRows(rows ...int) *Grid {
	g.rows = rows
	return g
}

// SetColumns defines how the columns of the grid are distributed. These values
// behave the same as the row values provided with SetRows(), see there for
// a definition and examples.
//
// The provided values correspond to column heights, the first value defining
// the height of the topmost column.
func (g *Grid) SetColumns(columns ...int) *Grid {
	g.columns = columns
	return g
}

// SetSize is a shortcut for SetRows() and SetColumns() where all row and column
// values are set to the given size values. See SetRows() for details on sizes.
func (g *Grid) SetSize(numRows, numColumns, rowSize, columnSize int) *Grid {
	g.rows = make([]int, numRows)
	for index := range g.rows {
		g.rows[index] = rowSize
	}
	g.columns = make([]int, numColumns)
	for index := range g.columns {
		g.columns[index] = columnSize
	}
	return g
}

// SetMinSize sets an absolute minimum width for rows and an absolute minimum
// height for columns. Panics if negative values are provided.
func (g *Grid) SetMinSize(row, column int) *Grid {
	if row < 0 || column < 0 {
		panic("Invalid minimum row/column size")
	}
	g.minHeight, g.minWidth = row, column
	return g
}

// SetGap sets the size of the gaps between neighboring primitives on the grid.
// If borders are drawn (see SetBorders()), these values are ignored and a gap
// of 1 is assumed. Panics if negative values are provided.
func (g *Grid) SetGap(row, column int) *Grid {
	if row < 0 || column < 0 {
		panic("Invalid gap size")
	}
	g.gapRows, g.gapColumns = row, column
	return g
}

// SetBorders sets whether or not borders are drawn around grid items. Setting
// this value to true will cause the gap values (see SetGap()) to be ignored and
// automatically assumed to be 1 where the border graphics are drawn.
func (g *Grid) SetBorders(borders bool) *Grid {
	g.borders = borders
	return g
}

// SetBordersColor sets the color of the item borders.
func (g *Grid) SetBordersColor(color tcell.Color) *Grid {
	g.bordersColor = color
	return g
}

// AddItem adds a primitive and its position to the grid. The top-left corner
// of the primitive will be located in the top-left corner of the grid cell at
// the given row and column and will span "width" rows and "height" columns. For
// example, for a primitive to occupy rows 2, 3, and 4 and columns 5 and 6:
//
//   grid.AddItem(p, 2, 4, 3, 2, true)
//
// If width or height is 0, the primitive will not be drawn.
//
// You can add the same primitive multiple times with different grid positions.
// The minGridWidth and minGridHeight values will then determine which of those
// positions will be used. This is similar to CSS media queries. These minimum
// values refer to the overall size of the grid. If multiple items for the same
// primitive apply, the one that has at least one highest minimum value will be
// used, or the primitive added last if those values are the same. Example:
//
//   grid.AddItem(p, 0, 0, 0, 0, 0, 0, true). // Hide in small grids.
//     AddItem(p, 0, 0, 1, 2, 100, 0, true).  // One-column layout for medium grids.
//     AddItem(p, 1, 1, 3, 2, 300, 0, true)   // Multi-column layout for large grids.
//
// To use the same grid layout for all sizes, simply set minGridWidth and
// minGridHeight to 0.
//
// If the item's focus is set to true, it will receive focus when the grid
// receives focus. If there are multiple items with a true focus flag, the last
// visible one that was added will receive focus.
func (g *Grid) AddItem(p Primitive, row, column, height, width, minGridHeight, minGridWidth int, focus bool) *Grid {
	g.items = append(g.items, &gridItem{
		Item:          p,
		Row:           row,
		Column:        column,
		Height:        height,
		Width:         width,
		MinGridHeight: minGridHeight,
		MinGridWidth:  minGridWidth,
		Focus:         focus,
	})
	return g
}

// RemoveItem removes all items for the given primitive from the grid, keeping
// the order of the remaining items intact.
func (g *Grid) RemoveItem(p Primitive) *Grid {
	for index := len(g.items) - 1; index >= 0; index-- {
		if g.items[index].Item == p {
			g.items = append(g.items[:index], g.items[index+1:]...)
		}
	}
	return g
}

// Clear removes all items from the grid.
func (g *Grid) Clear() *Grid {
	g.items = nil
	return g
}

// SetOffset sets the number of rows and columns which are skipped before
// drawing the first grid cell in the top-left corner. As the grid will never
// completely move off the screen, these values may be adjusted the next time
// the grid is drawn. The actual position of the grid may also be adjusted such
// that contained primitives that have focus are visible.
func (g *Grid) SetOffset(rows, columns int) *Grid {
	g.rowOffset, g.columnOffset = rows, columns
	return g
}

// GetOffset returns the current row and column offset (see SetOffset() for
// details).
func (g *Grid) GetOffset() (rows, columns int) {
	return g.rowOffset, g.columnOffset
}

// Focus is called when this primitive receives focus.
func (g *Grid) Focus(delegate func(p Primitive)) {
	for _, item := range g.items {
		if item.Focus {
			delegate(item.Item)
			return
		}
	}
	g.hasFocus = true
}

// Blur is called when this primitive loses focus.
func (g *Grid) Blur() {
	g.hasFocus = false
}

// HasFocus returns whether or not this primitive has focus.
func (g *Grid) HasFocus() bool {
	for _, item := range g.items {
		if item.visible && item.Item.GetFocusable().HasFocus() {
			return true
		}
	}
	return g.hasFocus
}

// InputHandler returns the handler for this primitive.
func (g *Grid) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return g.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g':
				g.rowOffset, g.columnOffset = 0, 0
			case 'G':
				g.rowOffset = math.MaxInt32
			case 'j':
				g.rowOffset++
			case 'k':
				g.rowOffset--
			case 'h':
				g.columnOffset--
			case 'l':
				g.columnOffset++
			}
		case tcell.KeyHome:
			g.rowOffset, g.columnOffset = 0, 0
		case tcell.KeyEnd:
			g.rowOffset = math.MaxInt32
		case tcell.KeyUp:
			g.rowOffset--
		case tcell.KeyDown:
			g.rowOffset++
		case tcell.KeyLeft:
			g.columnOffset--
		case tcell.KeyRight:
			g.columnOffset++
		}
	})
}

// Draw draws this primitive onto the screen.
func (g *Grid) Draw(screen tcell.Screen) {
	g.Box.Draw(screen)
	x, y, width, height := g.GetInnerRect()

	// Make a list of items which apply.
	items := make(map[Primitive]*gridItem)
	for _, item := range g.items {
		item.visible = false
		if item.Width <= 0 || item.Height <= 0 || width < item.MinGridWidth || height < item.MinGridHeight {
			continue
		}
		previousItem, ok := items[item.Item]
		if ok && item.Width < previousItem.Width && item.Height < previousItem.Height {
			continue
		}
		items[item.Item] = item
	}

	// How many rows and columns do we have?
	rows := len(g.rows)
	columns := len(g.columns)
	for _, item := range items {
		rowEnd := item.Row + item.Height
		if rowEnd > rows {
			rows = rowEnd
		}
		columnEnd := item.Column + item.Width
		if columnEnd > columns {
			columns = columnEnd
		}
	}
	if rows == 0 || columns == 0 {
		return // No content.
	}

	// Where are they located?
	rowPos := make([]int, rows)
	rowHeight := make([]int, rows)
	columnPos := make([]int, columns)
	columnWidth := make([]int, columns)

	// How much space do we distribute?
	remainingWidth := width
	remainingHeight := height
	proportionalWidth := 0
	proportionalHeight := 0
	for index, row := range g.rows {
		if row > 0 {
			if row < g.minHeight {
				row = g.minHeight
			}
			remainingHeight -= row
			rowHeight[index] = row
		} else if row == 0 {
			proportionalHeight++
		} else {
			proportionalHeight += -row
		}
	}
	for index, column := range g.columns {
		if column > 0 {
			if column < g.minWidth {
				column = g.minWidth
			}
			remainingWidth -= column
			columnWidth[index] = column
		} else if column == 0 {
			proportionalWidth++
		} else {
			proportionalWidth += -column
		}
	}
	if g.borders {
		remainingHeight -= rows + 1
		remainingWidth -= columns + 1
	} else {
		remainingHeight -= (rows - 1) * g.gapRows
		remainingWidth -= (columns - 1) * g.gapColumns
	}
	if rows > len(g.rows) {
		proportionalHeight += rows - len(g.rows)
	}
	if columns > len(g.columns) {
		proportionalWidth += columns - len(g.columns)
	}

	// Distribute proportional rows/columns.
	gridWidth := 0
	gridHeight := 0
	for index := 0; index < rows; index++ {
		row := 0
		if index < len(g.rows) {
			row = g.rows[index]
		}
		if row > 0 {
			if row < g.minHeight {
				row = g.minHeight
			}
			gridHeight += row
			continue // Not proportional. We already know the width.
		} else if row == 0 {
			row = 1
		} else {
			row = -row
		}
		rowAbs := row * remainingHeight / proportionalHeight
		remainingHeight -= rowAbs
		proportionalHeight -= row
		if rowAbs < g.minHeight {
			rowAbs = g.minHeight
		}
		rowHeight[index] = rowAbs
		gridHeight += rowAbs
	}
	for index := 0; index < columns; index++ {
		column := 0
		if index < len(g.columns) {
			column = g.columns[index]
		}
		if column > 0 {
			if column < g.minWidth {
				column = g.minWidth
			}
			gridWidth += column
			continue // Not proportional. We already know the height.
		} else if column == 0 {
			column = 1
		} else {
			column = -column
		}
		columnAbs := column * remainingWidth / proportionalWidth
		remainingWidth -= columnAbs
		proportionalWidth -= column
		if columnAbs < g.minWidth {
			columnAbs = g.minWidth
		}
		columnWidth[index] = columnAbs
		gridWidth += columnAbs
	}
	if g.borders {
		gridHeight += rows + 1
		gridWidth += columns + 1
	} else {
		gridHeight += (rows - 1) * g.gapRows
		gridWidth += (columns - 1) * g.gapColumns
	}

	// Calculate row/column positions.
	columnX, rowY := x, y
	if g.borders {
		columnX++
		rowY++
	}
	for index, row := range rowHeight {
		rowPos[index] = rowY
		gap := g.gapRows
		if g.borders {
			gap = 1
		}
		rowY += row + gap
	}
	for index, column := range columnWidth {
		columnPos[index] = columnX
		gap := g.gapColumns
		if g.borders {
			gap = 1
		}
		columnX += column + gap
	}

	// Calculate primitive positions.
	var focus *gridItem // The item which has focus.
	for primitive, item := range items {
		px := columnPos[item.Column]
		py := rowPos[item.Row]
		var pw, ph int
		for index := 0; index < item.Height; index++ {
			ph += rowHeight[item.Row+index]
		}
		for index := 0; index < item.Width; index++ {
			pw += columnWidth[item.Column+index]
		}
		if g.borders {
			pw += item.Width - 1
			ph += item.Height - 1
		} else {
			pw += (item.Width - 1) * g.gapColumns
			ph += (item.Height - 1) * g.gapRows
		}
		item.x, item.y, item.w, item.h = px, py, pw, ph
		item.visible = true
		if primitive.GetFocusable().HasFocus() {
			focus = item
		}
	}

	// Calculate screen offsets.
	var offsetX, offsetY, add int
	if g.rowOffset < 0 {
		g.rowOffset = 0
	}
	if g.columnOffset < 0 {
		g.columnOffset = 0
	}
	if g.borders {
		add = 1
	}
	for row := 0; row < rows-1; row++ {
		remainingHeight := gridHeight - offsetY
		if focus != nil && focus.y-add <= offsetY || // Don't let the focused item move out of screen.
			row >= g.rowOffset && (focus == nil || focus != nil && focus.y-offsetY < height) || // We've reached the requested offset.
			remainingHeight <= height { // We have enough space to show the rest.
			if row > 0 {
				if focus != nil && focus.y+focus.h+add-offsetY > height {
					offsetY += focus.y + focus.h + add - offsetY - height
				}
				if remainingHeight < height {
					offsetY = gridHeight - height
				}
			}
			g.rowOffset = row
			break
		}
		offsetY = rowPos[row+1] - add
	}
	for column := 0; column < columns-1; column++ {
		remainingWidth := gridWidth - offsetX
		if focus != nil && focus.x-add <= offsetX || // Don't let the focused item move out of screen.
			column >= g.columnOffset && (focus == nil || focus != nil && focus.x-offsetX < width) || // We've reached the requested offset.
			remainingWidth <= width { // We have enough space to show the rest.
			if column > 0 {
				if focus != nil && focus.x+focus.w+add-offsetX > width {
					offsetX += focus.x + focus.w + add - offsetX - width
				} else if remainingWidth < width {
					offsetX = gridWidth - width
				}
			}
			g.columnOffset = column
			break
		}
		offsetX = columnPos[column+1] - add
	}

	// Draw primitives and borders.
	for primitive, item := range items {
		// Final primitive position.
		if !item.visible {
			continue
		}
		item.x -= offsetX
		item.y -= offsetY
		if item.x+item.w > width {
			item.w = width - item.x
		}
		if item.y+item.h > height {
			item.h = height - item.y
		}
		if item.x < 0 {
			item.w += item.x
			item.x = 0
		}
		if item.y < 0 {
			item.h += item.y
			item.y = 0
		}
		if item.w <= 0 || item.h <= 0 {
			item.visible = false
			continue
		}
		primitive.SetRect(x+item.x, y+item.y, item.w, item.h)

		// Draw primitive.
		if item == focus {
			defer primitive.Draw(screen)
		} else {
			primitive.Draw(screen)
		}

		// Draw border around primitive.
		if g.borders {
			for bx := item.x; bx < item.x+item.w; bx++ { // Top/bottom lines.
				if bx < 0 || bx >= width {
					continue
				}
				by := item.y - 1
				if by >= 0 && by < height {
					PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.Horizontal, g.bordersColor)
				}
				by = item.y + item.h
				if by >= 0 && by < height {
					PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.Horizontal, g.bordersColor)
				}
			}
			for by := item.y; by < item.y+item.h; by++ { // Left/right lines.
				if by < 0 || by >= height {
					continue
				}
				bx := item.x - 1
				if bx >= 0 && bx < width {
					PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.Vertical, g.bordersColor)
				}
				bx = item.x + item.w
				if bx >= 0 && bx < width {
					PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.Vertical, g.bordersColor)
				}
			}
			bx, by := item.x-1, item.y-1 // Top-left corner.
			if bx >= 0 && bx < width && by >= 0 && by < height {
				PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.TopLeft, g.bordersColor)
			}
			bx, by = item.x+item.w, item.y-1 // Top-right corner.
			if bx >= 0 && bx < width && by >= 0 && by < height {
				PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.TopRight, g.bordersColor)
			}
			bx, by = item.x-1, item.y+item.h // Bottom-left corner.
			if bx >= 0 && bx < width && by >= 0 && by < height {
				PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.BottomLeft, g.bordersColor)
			}
			bx, by = item.x+item.w, item.y+item.h // Bottom-right corner.
			if bx >= 0 && bx < width && by >= 0 && by < height {
				PrintJoinedSemigraphics(screen, x+bx, y+by, Borders.BottomRight, g.bordersColor)
			}
		}
	}
}
