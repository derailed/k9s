package tview

import (
	"sort"

	"github.com/gdamore/tcell"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// TableCell represents one cell inside a Table. You can instantiate this type
// directly but all colors (background and text) will be set to their default
// which is black.
type TableCell struct {
	// The text to be displayed in the table cell.
	Text string

	// The alignment of the cell text. One of AlignLeft (default), AlignCenter,
	// or AlignRight.
	Align int

	// The maximum width of the cell in screen space. This is used to give a
	// column a maximum width. Any cell text whose screen width exceeds this width
	// is cut off. Set to 0 if there is no maximum width.
	MaxWidth int

	// If the total table width is less than the available width, this value is
	// used to add extra width to a column. See SetExpansion() for details.
	Expansion int

	// The color of the cell text.
	Color tcell.Color

	// The background color of the cell.
	BackgroundColor tcell.Color

	// The style attributes of the cell.
	Attributes tcell.AttrMask

	// If set to true, this cell cannot be selected.
	NotSelectable bool

	// The position and width of the cell the last time table was drawn.
	x, y, width int
}

// NewTableCell returns a new table cell with sensible defaults. That is, left
// aligned text with the primary text color (see Styles) and a transparent
// background (using the background of the Table).
func NewTableCell(text string) *TableCell {
	return &TableCell{
		Text:            text,
		Align:           AlignLeft,
		Color:           Styles.PrimaryTextColor,
		BackgroundColor: tcell.ColorDefault,
	}
}

// SetText sets the cell's text.
func (c *TableCell) SetText(text string) *TableCell {
	c.Text = text
	return c
}

// SetAlign sets the cell's text alignment, one of AlignLeft, AlignCenter, or
// AlignRight.
func (c *TableCell) SetAlign(align int) *TableCell {
	c.Align = align
	return c
}

// SetMaxWidth sets maximum width of the cell in screen space. This is used to
// give a column a maximum width. Any cell text whose screen width exceeds this
// width is cut off. Set to 0 if there is no maximum width.
func (c *TableCell) SetMaxWidth(maxWidth int) *TableCell {
	c.MaxWidth = maxWidth
	return c
}

// SetExpansion sets the value by which the column of this cell expands if the
// available width for the table is more than the table width (prior to applying
// this expansion value). This is a proportional value. The amount of unused
// horizontal space is divided into widths to be added to each column. How much
// extra width a column receives depends on the expansion value: A value of 0
// (the default) will not cause the column to increase in width. Other values
// are proportional, e.g. a value of 2 will cause a column to grow by twice
// the amount of a column with a value of 1.
//
// Since this value affects an entire column, the maximum over all visible cells
// in that column is used.
//
// This function panics if a negative value is provided.
func (c *TableCell) SetExpansion(expansion int) *TableCell {
	if expansion < 0 {
		panic("Table cell expansion values may not be negative")
	}
	c.Expansion = expansion
	return c
}

// SetTextColor sets the cell's text color.
func (c *TableCell) SetTextColor(color tcell.Color) *TableCell {
	c.Color = color
	return c
}

// SetBackgroundColor sets the cell's background color. Set to
// tcell.ColorDefault to use the table's background color.
func (c *TableCell) SetBackgroundColor(color tcell.Color) *TableCell {
	c.BackgroundColor = color
	return c
}

// SetAttributes sets the cell's text attributes. You can combine different
// attributes using bitmask operations:
//
//   cell.SetAttributes(tcell.AttrUnderline | tcell.AttrBold)
func (c *TableCell) SetAttributes(attr tcell.AttrMask) *TableCell {
	c.Attributes = attr
	return c
}

// SetStyle sets the cell's style (foreground color, background color, and
// attributes) all at once.
func (c *TableCell) SetStyle(style tcell.Style) *TableCell {
	c.Color, c.BackgroundColor, c.Attributes = style.Decompose()
	return c
}

// SetSelectable sets whether or not this cell can be selected by the user.
func (c *TableCell) SetSelectable(selectable bool) *TableCell {
	c.NotSelectable = !selectable
	return c
}

// GetLastPosition returns the position of the table cell the last time it was
// drawn on screen. If the cell is not on screen, the return values are
// undefined.
//
// Because the Table class will attempt to keep selected cells on screen, this
// function is most useful in response to a "selected" event (see
// SetSelectedFunc()) or a "selectionChanged" event (see
// SetSelectionChangedFunc()).
func (c *TableCell) GetLastPosition() (x, y, width int) {
	return c.x, c.y, c.width
}

// Table visualizes two-dimensional data consisting of rows and columns. Each
// Table cell is defined via SetCell() by the TableCell type. They can be added
// dynamically to the table and changed any time.
//
// The most compact display of a table is without borders. Each row will then
// occupy one row on screen and columns are separated by the rune defined via
// SetSeparator() (a space character by default).
//
// When borders are turned on (via SetBorders()), each table cell is surrounded
// by lines. Therefore one table row will require two rows on screen.
//
// Columns will use as much horizontal space as they need. You can constrain
// their size with the MaxWidth parameter of the TableCell type.
//
// Fixed Columns
//
// You can define fixed rows and rolumns via SetFixed(). They will always stay
// in their place, even when the table is scrolled. Fixed rows are always the
// top rows. Fixed columns are always the leftmost columns.
//
// Selections
//
// You can call SetSelectable() to set columns and/or rows to "selectable". If
// the flag is set only for columns, entire columns can be selected by the user.
// If it is set only for rows, entire rows can be selected. If both flags are
// set, individual cells can be selected. The "selected" handler set via
// SetSelectedFunc() is invoked when the user presses Enter on a selection.
//
// Navigation
//
// If the table extends beyond the available space, it can be navigated with
// key bindings similar to Vim:
//
//   - h, left arrow: Move left by one column.
//   - l, right arrow: Move right by one column.
//   - j, down arrow: Move down by one row.
//   - k, up arrow: Move up by one row.
//   - g, home: Move to the top.
//   - G, end: Move to the bottom.
//   - Ctrl-F, page down: Move down by one page.
//   - Ctrl-B, page up: Move up by one page.
//
// When there is no selection, this affects the entire table (except for fixed
// rows and columns). When there is a selection, the user moves the selection.
// The class will attempt to keep the selection from moving out of the screen.
//
// Use SetInputCapture() to override or modify keyboard input.
//
// See https://github.com/rivo/tview/wiki/Table for an example.
type Table struct {
	*Box

	// Whether or not this table has borders around each cell.
	borders bool

	// The color of the borders or the separator.
	bordersColor tcell.Color

	// If there are no borders, the column separator.
	separator rune

	// The cells of the table. Rows first, then columns.
	cells [][]*TableCell

	// The rightmost column in the data set.
	lastColumn int

	// The number of fixed rows / columns.
	fixedRows, fixedColumns int

	// Whether or not rows or columns can be selected. If both are set to true,
	// cells can be selected.
	rowsSelectable, columnsSelectable bool

	// The currently selected row and column.
	selectedRow, selectedColumn int

	// The number of rows/columns by which the table is scrolled down/to the
	// right.
	rowOffset, columnOffset int

	// If set to true, the table's last row will always be visible.
	trackEnd bool

	// The number of visible rows the last time the table was drawn.
	visibleRows int

	// The style of the selected rows. If this value is 0, selected rows are
	// simply inverted.
	selectedStyle tcell.Style

	// An optional function which gets called when the user presses Enter on a
	// selected cell. If entire rows selected, the column value is undefined.
	// Likewise for entire columns.
	selected func(row, column int)

	// An optional function which gets called when the user changes the selection.
	// If entire rows selected, the column value is undefined.
	// Likewise for entire columns.
	selectionChanged func(row, column int)

	// An optional function which gets called when the user presses Escape, Tab,
	// or Backtab. Also when the user presses Enter if nothing is selectable.
	done func(key tcell.Key)
}

// NewTable returns a new table.
func NewTable() *Table {
	return &Table{
		Box:          NewBox(),
		bordersColor: Styles.GraphicsColor,
		separator:    ' ',
		lastColumn:   -1,
	}
}

// Clear removes all table data.
func (t *Table) Clear() *Table {
	t.cells = nil
	t.lastColumn = -1
	return t
}

// SetBorders sets whether or not each cell in the table is surrounded by a
// border.
func (t *Table) SetBorders(show bool) *Table {
	t.borders = show
	return t
}

// SetBordersColor sets the color of the cell borders.
func (t *Table) SetBordersColor(color tcell.Color) *Table {
	t.bordersColor = color
	return t
}

// SetSelectedStyle sets a specific style for selected cells. If no such style
// is set, per default, selected cells are inverted (i.e. their foreground and
// background colors are swapped).
//
// To reset a previous setting to its default, make the following call:
//
//   table.SetSelectedStyle(tcell.ColorDefault, tcell.ColorDefault, 0)
func (t *Table) SetSelectedStyle(foregroundColor, backgroundColor tcell.Color, attributes tcell.AttrMask) *Table {
	t.selectedStyle = tcell.StyleDefault.Foreground(foregroundColor).Background(backgroundColor) | tcell.Style(attributes)
	return t
}

// SetSeparator sets the character used to fill the space between two
// neighboring cells. This is a space character ' ' per default but you may
// want to set it to Borders.Vertical (or any other rune) if the column
// separation should be more visible. If cell borders are activated, this is
// ignored.
//
// Separators have the same color as borders.
func (t *Table) SetSeparator(separator rune) *Table {
	t.separator = separator
	return t
}

// SetFixed sets the number of fixed rows and columns which are always visible
// even when the rest of the cells are scrolled out of view. Rows are always the
// top-most ones. Columns are always the left-most ones.
func (t *Table) SetFixed(rows, columns int) *Table {
	t.fixedRows, t.fixedColumns = rows, columns
	return t
}

// SetSelectable sets the flags which determine what can be selected in a table.
// There are three selection modi:
//
//   - rows = false, columns = false: Nothing can be selected.
//   - rows = true, columns = false: Rows can be selected.
//   - rows = false, columns = true: Columns can be selected.
//   - rows = true, columns = true: Individual cells can be selected.
func (t *Table) SetSelectable(rows, columns bool) *Table {
	t.rowsSelectable, t.columnsSelectable = rows, columns
	return t
}

// GetSelectable returns what can be selected in a table. Refer to
// SetSelectable() for details.
func (t *Table) GetSelectable() (rows, columns bool) {
	return t.rowsSelectable, t.columnsSelectable
}

// GetSelection returns the position of the current selection.
// If entire rows are selected, the column index is undefined.
// Likewise for entire columns.
func (t *Table) GetSelection() (row, column int) {
	return t.selectedRow, t.selectedColumn
}

// Select sets the selected cell. Depending on the selection settings
// specified via SetSelectable(), this may be an entire row or column, or even
// ignored completely.
func (t *Table) Select(row, column int) *Table {
	t.selectedRow, t.selectedColumn = row, column
	return t
}

// SetOffset sets how many rows and columns should be skipped when drawing the
// table. This is useful for large tables that do not fit on the screen.
// Navigating a selection can change these values.
//
// Fixed rows and columns are never skipped.
func (t *Table) SetOffset(row, column int) *Table {
	t.rowOffset, t.columnOffset = row, column
	return t
}

// GetOffset returns the current row and column offset. This indicates how many
// rows and columns the table is scrolled down and to the right.
func (t *Table) GetOffset() (row, column int) {
	return t.rowOffset, t.columnOffset
}

// SetSelectedFunc sets a handler which is called whenever the user presses the
// Enter key on a selected cell/row/column. The handler receives the position of
// the selection and its cell contents. If entire rows are selected, the column
// index is undefined. Likewise for entire columns.
func (t *Table) SetSelectedFunc(handler func(row, column int)) *Table {
	t.selected = handler
	return t
}

// SetSelectionChangedFunc sets a handler which is called whenever the user
// navigates to a new selection. The handler receives the position of the new
// selection. If entire rows are selected, the column index is undefined.
// Likewise for entire columns.
func (t *Table) SetSelectionChangedFunc(handler func(row, column int)) *Table {
	t.selectionChanged = handler
	return t
}

// SetDoneFunc sets a handler which is called whenever the user presses the
// Escape, Tab, or Backtab key. If nothing is selected, it is also called when
// user presses the Enter key (because pressing Enter on a selection triggers
// the "selected" handler set via SetSelectedFunc()).
func (t *Table) SetDoneFunc(handler func(key tcell.Key)) *Table {
	t.done = handler
	return t
}

// SetCell sets the content of a cell the specified position. It is ok to
// directly instantiate a TableCell object. If the cell has content, at least
// the Text and Color fields should be set.
//
// Note that setting cells in previously unknown rows and columns will
// automatically extend the internal table representation, e.g. starting with
// a row of 100,000 will immediately create 100,000 empty rows.
//
// To avoid unnecessary garbage collection, fill columns from left to right.
func (t *Table) SetCell(row, column int, cell *TableCell) *Table {
	if row >= len(t.cells) {
		t.cells = append(t.cells, make([][]*TableCell, row-len(t.cells)+1)...)
	}
	rowLen := len(t.cells[row])
	if column >= rowLen {
		t.cells[row] = append(t.cells[row], make([]*TableCell, column-rowLen+1)...)
		for c := rowLen; c < column; c++ {
			t.cells[row][c] = &TableCell{}
		}
	}
	t.cells[row][column] = cell
	if column > t.lastColumn {
		t.lastColumn = column
	}
	return t
}

// SetCellSimple calls SetCell() with the given text, left-aligned, in white.
func (t *Table) SetCellSimple(row, column int, text string) *Table {
	t.SetCell(row, column, NewTableCell(text))
	return t
}

// GetCell returns the contents of the cell at the specified position. A valid
// TableCell object is always returned but it will be uninitialized if the cell
// was not previously set.
func (t *Table) GetCell(row, column int) *TableCell {
	if row >= len(t.cells) || column >= len(t.cells[row]) {
		return &TableCell{}
	}
	return t.cells[row][column]
}

// GetRowCount returns the number of rows in the table.
func (t *Table) GetRowCount() int {
	return len(t.cells)
}

// GetColumnCount returns the (maximum) number of columns in the table.
func (t *Table) GetColumnCount() int {
	if len(t.cells) == 0 {
		return 0
	}
	return t.lastColumn + 1
}

// ScrollToBeginning scrolls the table to the beginning to that the top left
// corner of the table is shown. Note that this position may be corrected if
// there is a selection.
func (t *Table) ScrollToBeginning() *Table {
	t.trackEnd = false
	t.columnOffset = 0
	t.rowOffset = 0
	return t
}

// ScrollToEnd scrolls the table to the beginning to that the bottom left corner
// of the table is shown. Adding more rows to the table will cause it to
// automatically scroll with the new data. Note that this position may be
// corrected if there is a selection.
func (t *Table) ScrollToEnd() *Table {
	t.trackEnd = true
	t.columnOffset = 0
	t.rowOffset = len(t.cells)
	return t
}

// Draw draws this primitive onto the screen.
func (t *Table) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)

	// What's our available screen space?
	x, y, width, height := t.GetInnerRect()
	if t.borders {
		t.visibleRows = height / 2
	} else {
		t.visibleRows = height
	}

	// Return the cell at the specified position (nil if it doesn't exist).
	getCell := func(row, column int) *TableCell {
		if row < 0 || column < 0 || row >= len(t.cells) || column >= len(t.cells[row]) {
			return nil
		}
		return t.cells[row][column]
	}

	// If this cell is not selectable, find the next one.
	if t.rowsSelectable || t.columnsSelectable {
		if t.selectedColumn < 0 {
			t.selectedColumn = 0
		}
		if t.selectedRow < 0 {
			t.selectedRow = 0
		}
		for t.selectedRow < len(t.cells) {
			cell := getCell(t.selectedRow, t.selectedColumn)
			if cell == nil || !cell.NotSelectable {
				break
			}
			t.selectedColumn++
			if t.selectedColumn > t.lastColumn {
				t.selectedColumn = 0
				t.selectedRow++
			}
		}
	}

	// Clamp row offsets.
	if t.rowsSelectable {
		if t.selectedRow >= t.fixedRows && t.selectedRow < t.fixedRows+t.rowOffset {
			t.rowOffset = t.selectedRow - t.fixedRows
			t.trackEnd = false
		}
		if t.borders {
			if 2*(t.selectedRow+1-t.rowOffset) >= height {
				t.rowOffset = t.selectedRow + 1 - height/2
				t.trackEnd = false
			}
		} else {
			if t.selectedRow+1-t.rowOffset >= height {
				t.rowOffset = t.selectedRow + 1 - height
				t.trackEnd = false
			}
		}
	}
	if t.borders {
		if 2*(len(t.cells)-t.rowOffset) < height {
			t.trackEnd = true
		}
	} else {
		if len(t.cells)-t.rowOffset < height {
			t.trackEnd = true
		}
	}
	if t.trackEnd {
		if t.borders {
			t.rowOffset = len(t.cells) - height/2
		} else {
			t.rowOffset = len(t.cells) - height
		}
	}
	if t.rowOffset < 0 {
		t.rowOffset = 0
	}

	// Clamp column offset. (Only left side here. The right side is more
	// difficult and we'll do it below.)
	if t.columnsSelectable && t.selectedColumn >= t.fixedColumns && t.selectedColumn < t.fixedColumns+t.columnOffset {
		t.columnOffset = t.selectedColumn - t.fixedColumns
	}
	if t.columnOffset < 0 {
		t.columnOffset = 0
	}
	if t.selectedColumn < 0 {
		t.selectedColumn = 0
	}

	// Determine the indices and widths of the columns and rows which fit on the
	// screen.
	var (
		columns, rows, widths   []int
		tableHeight, tableWidth int
	)
	rowStep := 1
	if t.borders {
		rowStep = 2    // With borders, every table row takes two screen rows.
		tableWidth = 1 // We start at the second character because of the left table border.
	}
	indexRow := func(row int) bool { // Determine if this row is visible, store its index.
		if tableHeight >= height {
			return false
		}
		rows = append(rows, row)
		tableHeight += rowStep
		return true
	}
	for row := 0; row < t.fixedRows && row < len(t.cells); row++ { // Do the fixed rows first.
		if !indexRow(row) {
			break
		}
	}
	for row := t.fixedRows + t.rowOffset; row < len(t.cells); row++ { // Then the remaining rows.
		if !indexRow(row) {
			break
		}
	}
	var (
		skipped, lastTableWidth, expansionTotal int
		expansions                              []int
	)
ColumnLoop:
	for column := 0; ; column++ {
		// If we've moved beyond the right border, we stop or skip a column.
		for tableWidth-1 >= width { // -1 because we include one extra column if the separator falls on the right end of the box.
			// We've moved beyond the available space.
			if column < t.fixedColumns {
				break ColumnLoop // We're in the fixed area. We're done.
			}
			if !t.columnsSelectable && skipped >= t.columnOffset {
				break ColumnLoop // There is no selection and we've already reached the offset.
			}
			if t.columnsSelectable && t.selectedColumn-skipped == t.fixedColumns {
				break ColumnLoop // The selected column reached the leftmost point before disappearing.
			}
			if t.columnsSelectable && skipped >= t.columnOffset &&
				(t.selectedColumn < column && lastTableWidth < width-1 && tableWidth < width-1 || t.selectedColumn < column-1) {
				break ColumnLoop // We've skipped as many as requested and the selection is visible.
			}
			if len(columns) <= t.fixedColumns {
				break // Nothing to skip.
			}

			// We need to skip a column.
			skipped++
			lastTableWidth -= widths[t.fixedColumns] + 1
			tableWidth -= widths[t.fixedColumns] + 1
			columns = append(columns[:t.fixedColumns], columns[t.fixedColumns+1:]...)
			widths = append(widths[:t.fixedColumns], widths[t.fixedColumns+1:]...)
			expansions = append(expansions[:t.fixedColumns], expansions[t.fixedColumns+1:]...)
		}

		// What's this column's width (without expansion)?
		maxWidth := -1
		expansion := 0
		for _, row := range rows {
			if cell := getCell(row, column); cell != nil {
				_, _, _, _, cellWidth := decomposeString(cell.Text)
				if cell.MaxWidth > 0 && cell.MaxWidth < cellWidth {
					cellWidth = cell.MaxWidth
				}
				if cellWidth > maxWidth {
					maxWidth = cellWidth
				}
				if cell.Expansion > expansion {
					expansion = cell.Expansion
				}
			}
		}
		if maxWidth < 0 {
			break // No more cells found in this column.
		}

		// Store new column info at the end.
		columns = append(columns, column)
		widths = append(widths, maxWidth)
		lastTableWidth = tableWidth
		tableWidth += maxWidth + 1
		expansions = append(expansions, expansion)
		expansionTotal += expansion
	}
	t.columnOffset = skipped

	// If we have space left, distribute it.
	if tableWidth < width {
		toDistribute := width - tableWidth
		for index, expansion := range expansions {
			if expansionTotal <= 0 {
				break
			}
			expWidth := toDistribute * expansion / expansionTotal
			widths[index] += expWidth
			toDistribute -= expWidth
			expansionTotal -= expansion
		}
	}

	// Helper function which draws border runes.
	borderStyle := tcell.StyleDefault.Background(t.backgroundColor).Foreground(t.bordersColor)
	drawBorder := func(colX, rowY int, ch rune) {
		screen.SetContent(x+colX, y+rowY, ch, nil, borderStyle)
	}

	// Draw the cells (and borders).
	var columnX int
	if !t.borders {
		columnX--
	}
	for columnIndex, column := range columns {
		columnWidth := widths[columnIndex]
		for rowY, row := range rows {
			if t.borders {
				// Draw borders.
				rowY *= 2
				for pos := 0; pos < columnWidth && columnX+1+pos < width; pos++ {
					drawBorder(columnX+pos+1, rowY, Borders.Horizontal)
				}
				ch := Borders.Cross
				if columnIndex == 0 {
					if rowY == 0 {
						ch = Borders.TopLeft
					} else {
						ch = Borders.LeftT
					}
				} else if rowY == 0 {
					ch = Borders.TopT
				}
				drawBorder(columnX, rowY, ch)
				rowY++
				if rowY >= height {
					break // No space for the text anymore.
				}
				drawBorder(columnX, rowY, Borders.Vertical)
			} else if columnIndex > 0 {
				// Draw separator.
				drawBorder(columnX, rowY, t.separator)
			}

			// Get the cell.
			cell := getCell(row, column)
			if cell == nil {
				continue
			}

			// Draw text.
			finalWidth := columnWidth
			if columnX+1+columnWidth >= width {
				finalWidth = width - columnX - 1
			}
			cell.x, cell.y, cell.width = x+columnX+1, y+rowY, finalWidth
			_, printed := printWithStyle(screen, cell.Text, x+columnX+1, y+rowY, finalWidth, cell.Align, tcell.StyleDefault.Foreground(cell.Color)|tcell.Style(cell.Attributes))
			if StringWidth(cell.Text)-printed > 0 && printed > 0 {
				_, _, style, _ := screen.GetContent(x+columnX+1+finalWidth-1, y+rowY)
				printWithStyle(screen, string(SemigraphicsHorizontalEllipsis), x+columnX+1+finalWidth-1, y+rowY, 1, AlignLeft, style)
			}
		}

		// Draw bottom border.
		if rowY := 2 * len(rows); t.borders && rowY < height {
			for pos := 0; pos < columnWidth && columnX+1+pos < width; pos++ {
				drawBorder(columnX+pos+1, rowY, Borders.Horizontal)
			}
			ch := Borders.BottomT
			if columnIndex == 0 {
				ch = Borders.BottomLeft
			}
			drawBorder(columnX, rowY, ch)
		}

		columnX += columnWidth + 1
	}

	// Draw right border.
	if t.borders && len(t.cells) > 0 && columnX < width {
		for rowY := range rows {
			rowY *= 2
			if rowY+1 < height {
				drawBorder(columnX, rowY+1, Borders.Vertical)
			}
			ch := Borders.RightT
			if rowY == 0 {
				ch = Borders.TopRight
			}
			drawBorder(columnX, rowY, ch)
		}
		if rowY := 2 * len(rows); rowY < height {
			drawBorder(columnX, rowY, Borders.BottomRight)
		}
	}

	// Helper function which colors the background of a box.
	// backgroundColor == tcell.ColorDefault => Don't color the background.
	// textColor == tcell.ColorDefault => Don't change the text color.
	// attr == 0 => Don't change attributes.
	// invert == true => Ignore attr, set text to backgroundColor or t.backgroundColor;
	//                   set background to textColor.
	colorBackground := func(fromX, fromY, w, h int, backgroundColor, textColor tcell.Color, attr tcell.AttrMask, invert bool) {
		for by := 0; by < h && fromY+by < y+height; by++ {
			for bx := 0; bx < w && fromX+bx < x+width; bx++ {
				m, c, style, _ := screen.GetContent(fromX+bx, fromY+by)
				fg, bg, a := style.Decompose()
				if invert {
					if fg == textColor || fg == t.bordersColor {
						fg = backgroundColor
					}
					if fg == tcell.ColorDefault {
						fg = t.backgroundColor
					}
					style = style.Background(textColor).Foreground(fg)
				} else {
					if backgroundColor != tcell.ColorDefault {
						bg = backgroundColor
					}
					if textColor != tcell.ColorDefault {
						fg = textColor
					}
					if attr != 0 {
						a = attr
					}
					style = style.Background(bg).Foreground(fg) | tcell.Style(a)
				}
				screen.SetContent(fromX+bx, fromY+by, m, c, style)
			}
		}
	}

	// Color the cell backgrounds. To avoid undesirable artefacts, we combine
	// the drawing of a cell by background color, selected cells last.
	type cellInfo struct {
		x, y, w, h int
		text       tcell.Color
		selected   bool
	}
	cellsByBackgroundColor := make(map[tcell.Color][]*cellInfo)
	var backgroundColors []tcell.Color
	for rowY, row := range rows {
		columnX := 0
		rowSelected := t.rowsSelectable && !t.columnsSelectable && row == t.selectedRow
		for columnIndex, column := range columns {
			columnWidth := widths[columnIndex]
			cell := getCell(row, column)
			if cell == nil {
				continue
			}
			bx, by, bw, bh := x+columnX, y+rowY, columnWidth+1, 1
			if t.borders {
				by = y + rowY*2
				bw++
				bh = 3
			}
			columnSelected := t.columnsSelectable && !t.rowsSelectable && column == t.selectedColumn
			cellSelected := !cell.NotSelectable && (columnSelected || rowSelected || t.rowsSelectable && t.columnsSelectable && column == t.selectedColumn && row == t.selectedRow)
			entries, ok := cellsByBackgroundColor[cell.BackgroundColor]
			cellsByBackgroundColor[cell.BackgroundColor] = append(entries, &cellInfo{
				x:        bx,
				y:        by,
				w:        bw,
				h:        bh,
				text:     cell.Color,
				selected: cellSelected,
			})
			if !ok {
				backgroundColors = append(backgroundColors, cell.BackgroundColor)
			}
			columnX += columnWidth + 1
		}
	}
	sort.Slice(backgroundColors, func(i int, j int) bool {
		// Draw brightest colors last (i.e. on top).
		r, g, b := backgroundColors[i].RGB()
		c := colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}
		_, _, li := c.Hcl()
		r, g, b = backgroundColors[j].RGB()
		c = colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}
		_, _, lj := c.Hcl()
		return li < lj
	})
	selFg, selBg, selAttr := t.selectedStyle.Decompose()
	for _, bgColor := range backgroundColors {
		entries := cellsByBackgroundColor[bgColor]
		for _, cell := range entries {
			if cell.selected {
				if t.selectedStyle != 0 {
					defer colorBackground(cell.x, cell.y, cell.w, cell.h, selBg, selFg, selAttr, false)
				} else {
					defer colorBackground(cell.x, cell.y, cell.w, cell.h, bgColor, cell.text, 0, true)
				}
			} else {
				colorBackground(cell.x, cell.y, cell.w, cell.h, bgColor, tcell.ColorDefault, 0, false)
			}
		}
	}
}

// InputHandler returns the handler for this primitive.
func (t *Table) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		key := event.Key()

		if (!t.rowsSelectable && !t.columnsSelectable && key == tcell.KeyEnter) ||
			key == tcell.KeyEscape ||
			key == tcell.KeyTab ||
			key == tcell.KeyBacktab {
			if t.done != nil {
				t.done(key)
			}
			return
		}

		// Movement functions.
		previouslySelectedRow, previouslySelectedColumn := t.selectedRow, t.selectedColumn
		var (
			getCell = func(row, column int) *TableCell {
				if row < 0 || column < 0 || row >= len(t.cells) || column >= len(t.cells[row]) {
					return nil
				}
				return t.cells[row][column]
			}

			previous = func() {
				for t.selectedRow >= 0 {
					cell := getCell(t.selectedRow, t.selectedColumn)
					if cell == nil || !cell.NotSelectable {
						return
					}
					t.selectedColumn--
					if t.selectedColumn < 0 {
						t.selectedColumn = t.lastColumn
						t.selectedRow--
					}
				}
			}

			next = func() {
				if t.selectedColumn > t.lastColumn {
					t.selectedColumn = 0
					t.selectedRow++
					if t.selectedRow >= len(t.cells) {
						t.selectedRow = len(t.cells) - 1
					}
				}
				for t.selectedRow < len(t.cells) {
					cell := getCell(t.selectedRow, t.selectedColumn)
					if cell == nil || !cell.NotSelectable {
						return
					}
					t.selectedColumn++
					if t.selectedColumn > t.lastColumn {
						t.selectedColumn = 0
						t.selectedRow++
					}
				}
				t.selectedColumn = t.lastColumn
				t.selectedRow = len(t.cells) - 1
				previous()
			}

			home = func() {
				if t.rowsSelectable {
					t.selectedRow = 0
					t.selectedColumn = 0
					next()
				} else {
					t.trackEnd = false
					t.rowOffset = 0
					t.columnOffset = 0
				}
			}

			end = func() {
				if t.rowsSelectable {
					t.selectedRow = len(t.cells) - 1
					t.selectedColumn = t.lastColumn
					previous()
				} else {
					t.trackEnd = true
					t.columnOffset = 0
				}
			}

			down = func() {
				if t.rowsSelectable {
					t.selectedRow++
					if t.selectedRow >= len(t.cells) {
						t.selectedRow = len(t.cells) - 1
					}
					next()
				} else {
					t.rowOffset++
				}
			}

			up = func() {
				if t.rowsSelectable {
					t.selectedRow--
					if t.selectedRow < 0 {
						t.selectedRow = 0
					}
					previous()
				} else {
					t.trackEnd = false
					t.rowOffset--
				}
			}

			left = func() {
				if t.columnsSelectable {
					t.selectedColumn--
					if t.selectedColumn < 0 {
						t.selectedColumn = 0
					}
					previous()
				} else {
					t.columnOffset--
				}
			}

			right = func() {
				if t.columnsSelectable {
					t.selectedColumn++
					if t.selectedColumn > t.lastColumn {
						t.selectedColumn = t.lastColumn
					}
					next()
				} else {
					t.columnOffset++
				}
			}

			pageDown = func() {
				if t.rowsSelectable {
					t.selectedRow += t.visibleRows
					if t.selectedRow >= len(t.cells) {
						t.selectedRow = len(t.cells) - 1
					}
					next()
				} else {
					t.rowOffset += t.visibleRows
				}
			}

			pageUp = func() {
				if t.rowsSelectable {
					t.selectedRow -= t.visibleRows
					if t.selectedRow < 0 {
						t.selectedRow = 0
					}
					previous()
				} else {
					t.trackEnd = false
					t.rowOffset -= t.visibleRows
				}
			}
		)

		switch key {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g':
				home()
			case 'G':
				end()
			case 'j':
				down()
			case 'k':
				up()
			case 'h':
				left()
			case 'l':
				right()
			}
		case tcell.KeyHome:
			home()
		case tcell.KeyEnd:
			end()
		case tcell.KeyUp:
			up()
		case tcell.KeyDown:
			down()
		case tcell.KeyLeft:
			left()
		case tcell.KeyRight:
			right()
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			pageDown()
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			pageUp()
		case tcell.KeyEnter:
			if (t.rowsSelectable || t.columnsSelectable) && t.selected != nil {
				t.selected(t.selectedRow, t.selectedColumn)
			}
		}

		// If the selection has changed, notify the handler.
		if t.selectionChanged != nil &&
			(t.rowsSelectable && previouslySelectedRow != t.selectedRow ||
				t.columnsSelectable && previouslySelectedColumn != t.selectedColumn) {
			t.selectionChanged(t.selectedRow, t.selectedColumn)
		}
	})
}
