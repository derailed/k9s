package tview

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	colorful "github.com/lucasb-eyer/go-colorful"
	runewidth "github.com/mattn/go-runewidth"
)

// TabSize is the number of spaces with which a tab character will be replaced.
var TabSize = 4

// textViewIndex contains information about each line displayed in the text
// view.
type textViewIndex struct {
	Line            int    // The index into the "buffer" variable.
	Pos             int    // The index into the "buffer" string (byte position).
	NextPos         int    // The (byte) index of the next character in this buffer line.
	Width           int    // The screen width of this line.
	ForegroundColor string // The starting foreground color ("" = don't change, "-" = reset).
	BackgroundColor string // The starting background color ("" = don't change, "-" = reset).
	Attributes      string // The starting attributes ("" = don't change, "-" = reset).
	Region          string // The starting region ID.
}

// TextView is a box which displays text. It implements the io.Writer interface
// so you can stream text to it. This does not trigger a redraw automatically
// but if a handler is installed via SetChangedFunc(), you can cause it to be
// redrawn. (See SetChangedFunc() for more details.)
//
// Navigation
//
// If the text view is scrollable (the default), text is kept in a buffer which
// may be larger than the screen and can be navigated similarly to Vim:
//
//   - h, left arrow: Move left.
//   - l, right arrow: Move right.
//   - j, down arrow: Move down.
//   - k, up arrow: Move up.
//   - g, home: Move to the top.
//   - G, end: Move to the bottom.
//   - Ctrl-F, page down: Move down by one page.
//   - Ctrl-B, page up: Move up by one page.
//
// If the text is not scrollable, any text above the top visible line is
// discarded.
//
// Use SetInputCapture() to override or modify keyboard input.
//
// Colors
//
// If dynamic colors are enabled via SetDynamicColors(), text color can be
// changed dynamically by embedding color strings in square brackets. This works
// the same way as anywhere else. Please see the package documentation for more
// information.
//
// Regions and Highlights
//
// If regions are enabled via SetRegions(), you can define text regions within
// the text and assign region IDs to them. Text regions start with region tags.
// Region tags are square brackets that contain a region ID in double quotes,
// for example:
//
//   We define a ["rg"]region[""] here.
//
// A text region ends with the next region tag. Tags with no region ID ([""])
// don't start new regions. They can therefore be used to mark the end of a
// region. Region IDs must satisfy the following regular expression:
//
//   [a-zA-Z0-9_,;: \-\.]+
//
// Regions can be highlighted by calling the Highlight() function with one or
// more region IDs. This can be used to display search results, for example.
//
// The ScrollToHighlight() function can be used to jump to the currently
// highlighted region once when the text view is drawn the next time.
//
// See https://github.com/rivo/tview/wiki/TextView for an example.
type TextView struct {
	sync.Mutex
	*Box

	// The text buffer.
	buffer []string

	// The last bytes that have been received but are not part of the buffer yet.
	recentBytes []byte

	// The processed line index. This is nil if the buffer has changed and needs
	// to be re-indexed.
	index []*textViewIndex

	// The text alignment, one of AlignLeft, AlignCenter, or AlignRight.
	align int

	// Indices into the "index" slice which correspond to the first line of the
	// first highlight and the last line of the last highlight. This is calculated
	// during re-indexing. Set to -1 if there is no current highlight.
	fromHighlight, toHighlight int

	// The screen space column of the highlight in its first line. Set to -1 if
	// there is no current highlight.
	posHighlight int

	// A set of region IDs that are currently highlighted.
	highlights map[string]struct{}

	// The last width for which the current table is drawn.
	lastWidth int

	// The screen width of the longest line in the index (not the buffer).
	longestLine int

	// The index of the first line shown in the text view.
	lineOffset int

	// If set to true, the text view will always remain at the end of the content.
	trackEnd bool

	// The number of characters to be skipped on each line (not in wrap mode).
	columnOffset int

	// The height of the content the last time the text view was drawn.
	pageSize int

	// If set to true, the text view will keep a buffer of text which can be
	// navigated when the text is longer than what fits into the box.
	scrollable bool

	// If set to true, lines that are longer than the available width are wrapped
	// onto the next line. If set to false, any characters beyond the available
	// width are discarded.
	wrap bool

	// If set to true and if wrap is also true, lines are split at spaces or
	// after punctuation characters.
	wordWrap bool

	// The (starting) color of the text.
	textColor tcell.Color

	// If set to true, the text color can be changed dynamically by piping color
	// strings in square brackets to the text view.
	dynamicColors bool

	// If set to true, region tags can be used to define regions.
	regions bool

	// A temporary flag which, when true, will automatically bring the current
	// highlight(s) into the visible screen.
	scrollToHighlights bool

	// An optional function which is called when the content of the text view has
	// changed.
	changed func()

	// An optional function which is called when the user presses one of the
	// following keys: Escape, Enter, Tab, Backtab.
	done func(tcell.Key)
}

// NewTextView returns a new text view.
func NewTextView() *TextView {
	return &TextView{
		Box:           NewBox(),
		highlights:    make(map[string]struct{}),
		lineOffset:    -1,
		scrollable:    true,
		align:         AlignLeft,
		wrap:          true,
		textColor:     Styles.PrimaryTextColor,
		regions:       false,
		dynamicColors: false,
	}
}

// SetScrollable sets the flag that decides whether or not the text view is
// scrollable. If true, text is kept in a buffer and can be navigated.
func (t *TextView) SetScrollable(scrollable bool) *TextView {
	t.scrollable = scrollable
	if !scrollable {
		t.trackEnd = true
	}
	return t
}

// SetWrap sets the flag that, if true, leads to lines that are longer than the
// available width being wrapped onto the next line. If false, any characters
// beyond the available width are not displayed.
func (t *TextView) SetWrap(wrap bool) *TextView {
	if t.wrap != wrap {
		t.index = nil
	}
	t.wrap = wrap
	return t
}

// SetWordWrap sets the flag that, if true and if the "wrap" flag is also true
// (see SetWrap()), wraps the line at spaces or after punctuation marks. Note
// that trailing spaces will not be printed.
//
// This flag is ignored if the "wrap" flag is false.
func (t *TextView) SetWordWrap(wrapOnWords bool) *TextView {
	if t.wordWrap != wrapOnWords {
		t.index = nil
	}
	t.wordWrap = wrapOnWords
	return t
}

// SetTextAlign sets the text alignment within the text view. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (t *TextView) SetTextAlign(align int) *TextView {
	if t.align != align {
		t.index = nil
	}
	t.align = align
	return t
}

// SetTextColor sets the initial color of the text (which can be changed
// dynamically by sending color strings in square brackets to the text view if
// dynamic colors are enabled).
func (t *TextView) SetTextColor(color tcell.Color) *TextView {
	t.textColor = color
	return t
}

// SetText sets the text of this text view to the provided string. Previously
// contained text will be removed.
func (t *TextView) SetText(text string) *TextView {
	t.Clear()
	fmt.Fprint(t, text)
	return t
}

// SetDynamicColors sets the flag that allows the text color to be changed
// dynamically. See class description for details.
func (t *TextView) SetDynamicColors(dynamic bool) *TextView {
	if t.dynamicColors != dynamic {
		t.index = nil
	}
	t.dynamicColors = dynamic
	return t
}

// SetRegions sets the flag that allows to define regions in the text. See class
// description for details.
func (t *TextView) SetRegions(regions bool) *TextView {
	if t.regions != regions {
		t.index = nil
	}
	t.regions = regions
	return t
}

// SetChangedFunc sets a handler function which is called when the text of the
// text view has changed. This is useful when text is written to this io.Writer
// in a separate goroutine. This does not automatically cause the screen to be
// refreshed so you may want to use the "changed" handler to redraw the screen.
//
// Note that to avoid race conditions or deadlocks, there are a few rules you
// should follow:
//
//   - You can call Application.Draw() from this handler.
//   - You can call TextView.HasFocus() from this handler.
//   - During the execution of this handler, access to any other variables from
//     this primitive or any other primitive should be queued using
//     Application.QueueUpdate().
//
// See package description for details on dealing with concurrency.
func (t *TextView) SetChangedFunc(handler func()) *TextView {
	t.changed = handler
	return t
}

// SetDoneFunc sets a handler which is called when the user presses on the
// following keys: Escape, Enter, Tab, Backtab. The key is passed to the
// handler.
func (t *TextView) SetDoneFunc(handler func(key tcell.Key)) *TextView {
	t.done = handler
	return t
}

// ScrollTo scrolls to the specified row and column (both starting with 0).
func (t *TextView) ScrollTo(row, column int) *TextView {
	if !t.scrollable {
		return t
	}
	t.lineOffset = row
	t.columnOffset = column
	return t
}

// ScrollToBeginning scrolls to the top left corner of the text if the text view
// is scrollable.
func (t *TextView) ScrollToBeginning() *TextView {
	if !t.scrollable {
		return t
	}
	t.trackEnd = false
	t.lineOffset = 0
	t.columnOffset = 0
	return t
}

// ScrollToEnd scrolls to the bottom left corner of the text if the text view
// is scrollable. Adding new rows to the end of the text view will cause it to
// scroll with the new data.
func (t *TextView) ScrollToEnd() *TextView {
	if !t.scrollable {
		return t
	}
	t.trackEnd = true
	t.columnOffset = 0
	return t
}

// GetScrollOffset returns the number of rows and columns that are skipped at
// the top left corner when the text view has been scrolled.
func (t *TextView) GetScrollOffset() (row, column int) {
	return t.lineOffset, t.columnOffset
}

// Clear removes all text from the buffer.
func (t *TextView) Clear() *TextView {
	t.buffer = nil
	t.recentBytes = nil
	t.index = nil
	return t
}

// Highlight specifies which regions should be highlighted. See class
// description for details on regions. Empty region strings are ignored.
//
// Text in highlighted regions will be drawn inverted, i.e. with their
// background and foreground colors swapped.
//
// Calling this function will remove any previous highlights. To remove all
// highlights, call this function without any arguments.
func (t *TextView) Highlight(regionIDs ...string) *TextView {
	t.highlights = make(map[string]struct{})
	for _, id := range regionIDs {
		if id == "" {
			continue
		}
		t.highlights[id] = struct{}{}
	}
	t.index = nil
	return t
}

// GetHighlights returns the IDs of all currently highlighted regions.
func (t *TextView) GetHighlights() (regionIDs []string) {
	for id := range t.highlights {
		regionIDs = append(regionIDs, id)
	}
	return
}

// ScrollToHighlight will cause the visible area to be scrolled so that the
// highlighted regions appear in the visible area of the text view. This
// repositioning happens the next time the text view is drawn. It happens only
// once so you will need to call this function repeatedly to always keep
// highlighted regions in view.
//
// Nothing happens if there are no highlighted regions or if the text view is
// not scrollable.
func (t *TextView) ScrollToHighlight() *TextView {
	if len(t.highlights) == 0 || !t.scrollable || !t.regions {
		return t
	}
	t.index = nil
	t.scrollToHighlights = true
	t.trackEnd = false
	return t
}

// GetRegionText returns the text of the region with the given ID. If dynamic
// colors are enabled, color tags are stripped from the text. Newlines are
// always returned as '\n' runes.
//
// If the region does not exist or if regions are turned off, an empty string
// is returned.
func (t *TextView) GetRegionText(regionID string) string {
	if !t.regions || regionID == "" {
		return ""
	}

	var (
		buffer          bytes.Buffer
		currentRegionID string
	)

	for _, str := range t.buffer {
		// Find all color tags in this line.
		var colorTagIndices [][]int
		if t.dynamicColors {
			colorTagIndices = colorPattern.FindAllStringIndex(str, -1)
		}

		// Find all regions in this line.
		var (
			regionIndices [][]int
			regions       [][]string
		)
		if t.regions {
			regionIndices = regionPattern.FindAllStringIndex(str, -1)
			regions = regionPattern.FindAllStringSubmatch(str, -1)
		}

		// Analyze this line.
		var currentTag, currentRegion int
		for pos, ch := range str {
			// Skip any color tags.
			if currentTag < len(colorTagIndices) && pos >= colorTagIndices[currentTag][0] && pos < colorTagIndices[currentTag][1] {
				if pos == colorTagIndices[currentTag][1]-1 {
					currentTag++
				}
				continue
			}

			// Skip any regions.
			if currentRegion < len(regionIndices) && pos >= regionIndices[currentRegion][0] && pos < regionIndices[currentRegion][1] {
				if pos == regionIndices[currentRegion][1]-1 {
					if currentRegionID == regionID {
						// This is the end of the requested region. We're done.
						return buffer.String()
					}
					currentRegionID = regions[currentRegion][1]
					currentRegion++
				}
				continue
			}

			// Add this rune.
			if currentRegionID == regionID {
				buffer.WriteRune(ch)
			}
		}

		// Add newline.
		if currentRegionID == regionID {
			buffer.WriteRune('\n')
		}
	}

	return escapePattern.ReplaceAllString(buffer.String(), `[$1$2]`)
}

// Focus is called when this primitive receives focus.
func (t *TextView) Focus(delegate func(p Primitive)) {
	// Implemented here with locking because this is used by layout primitives.
	t.Lock()
	defer t.Unlock()
	t.hasFocus = true
}

// HasFocus returns whether or not this primitive has focus.
func (t *TextView) HasFocus() bool {
	// Implemented here with locking because this may be used in the "changed"
	// callback.
	t.Lock()
	defer t.Unlock()
	return t.hasFocus
}

// Write lets us implement the io.Writer interface. Tab characters will be
// replaced with TabSize space characters. A "\n" or "\r\n" will be interpreted
// as a new line.
func (t *TextView) Write(p []byte) (n int, err error) {
	// Notify at the end.
	t.Lock()
	changed := t.changed
	t.Unlock()
	if changed != nil {
		defer changed() // Deadlocks may occur if we lock here.
	}

	t.Lock()
	defer t.Unlock()

	// Copy data over.
	newBytes := append(t.recentBytes, p...)
	t.recentBytes = nil

	// If we have a trailing invalid UTF-8 byte, we'll wait.
	if r, _ := utf8.DecodeLastRune(p); r == utf8.RuneError {
		t.recentBytes = newBytes
		return len(p), nil
	}

	// If we have a trailing open dynamic color, exclude it.
	if t.dynamicColors {
		openColor := regexp.MustCompile(`\[([a-zA-Z]*|#[0-9a-zA-Z]*)$`)
		location := openColor.FindIndex(newBytes)
		if location != nil {
			t.recentBytes = newBytes[location[0]:]
			newBytes = newBytes[:location[0]]
		}
	}

	// If we have a trailing open region, exclude it.
	if t.regions {
		openRegion := regexp.MustCompile(`\["[a-zA-Z0-9_,;: \-\.]*"?$`)
		location := openRegion.FindIndex(newBytes)
		if location != nil {
			t.recentBytes = newBytes[location[0]:]
			newBytes = newBytes[:location[0]]
		}
	}

	// Transform the new bytes into strings.
	newLine := regexp.MustCompile(`\r?\n`)
	newBytes = bytes.Replace(newBytes, []byte{'\t'}, bytes.Repeat([]byte{' '}, TabSize), -1)
	for index, line := range newLine.Split(string(newBytes), -1) {
		if index == 0 {
			if len(t.buffer) == 0 {
				t.buffer = []string{line}
			} else {
				t.buffer[len(t.buffer)-1] += line
			}
		} else {
			t.buffer = append(t.buffer, line)
		}
	}

	// Reset the index.
	t.index = nil

	return len(p), nil
}

// reindexBuffer re-indexes the buffer such that we can use it to easily draw
// the buffer onto the screen. Each line in the index will contain a pointer
// into the buffer from which on we will print text. It will also contain the
// color with which the line starts.
func (t *TextView) reindexBuffer(width int) {
	if t.index != nil {
		return // Nothing has changed. We can still use the current index.
	}
	t.index = nil
	t.fromHighlight, t.toHighlight, t.posHighlight = -1, -1, -1

	// If there's no space, there's no index.
	if width < 1 {
		return
	}

	// Initial states.
	regionID := ""
	var highlighted bool

	// Go through each line in the buffer.
	for bufferIndex, str := range t.buffer {
		// Find all color tags in this line. Then remove them.
		var (
			colorTagIndices [][]int
			colorTags       [][]string
			escapeIndices   [][]int
		)
		strippedStr := str
		if t.dynamicColors {
			colorTagIndices, colorTags, escapeIndices, strippedStr, _ = decomposeString(str)
		}

		// Find all regions in this line. Then remove them.
		var (
			regionIndices [][]int
			regions       [][]string
		)
		if t.regions {
			regionIndices = regionPattern.FindAllStringIndex(str, -1)
			regions = regionPattern.FindAllStringSubmatch(str, -1)
			strippedStr = regionPattern.ReplaceAllString(strippedStr, "")
		}

		// We don't need the original string anymore for now.
		str = strippedStr

		// Split the line if required.
		var splitLines []string
		if t.wrap && len(str) > 0 {
			for len(str) > 0 {
				extract := runewidth.Truncate(str, width, "")
				if t.wordWrap && len(extract) < len(str) {
					// Add any spaces from the next line.
					if spaces := spacePattern.FindStringIndex(str[len(extract):]); spaces != nil && spaces[0] == 0 {
						extract = str[:len(extract)+spaces[1]]
					}

					// Can we split before the mandatory end?
					matches := boundaryPattern.FindAllStringIndex(extract, -1)
					if len(matches) > 0 {
						// Yes. Let's split there.
						extract = extract[:matches[len(matches)-1][1]]
					}
				}
				splitLines = append(splitLines, extract)
				str = str[len(extract):]
			}
		} else {
			// No need to split the line.
			splitLines = []string{str}
		}

		// Create index from split lines.
		var (
			originalPos, colorPos, regionPos, escapePos  int
			foregroundColor, backgroundColor, attributes string
		)
		for _, splitLine := range splitLines {
			line := &textViewIndex{
				Line:            bufferIndex,
				Pos:             originalPos,
				ForegroundColor: foregroundColor,
				BackgroundColor: backgroundColor,
				Attributes:      attributes,
				Region:          regionID,
			}

			// Shift original position with tags.
			lineLength := len(splitLine)
			remainingLength := lineLength
			tagEnd := originalPos
			totalTagLength := 0
			for {
				// Which tag comes next?
				nextTag := make([][3]int, 0, 3)
				if colorPos < len(colorTagIndices) {
					nextTag = append(nextTag, [3]int{colorTagIndices[colorPos][0], colorTagIndices[colorPos][1], 0}) // 0 = color tag.
				}
				if regionPos < len(regionIndices) {
					nextTag = append(nextTag, [3]int{regionIndices[regionPos][0], regionIndices[regionPos][1], 1}) // 1 = region tag.
				}
				if escapePos < len(escapeIndices) {
					nextTag = append(nextTag, [3]int{escapeIndices[escapePos][0], escapeIndices[escapePos][1], 2}) // 2 = escape tag.
				}
				minPos := -1
				tagIndex := -1
				for index, pair := range nextTag {
					if minPos < 0 || pair[0] < minPos {
						minPos = pair[0]
						tagIndex = index
					}
				}

				// Is the next tag in range?
				if tagIndex < 0 || minPos >= tagEnd+remainingLength {
					break // No. We're done with this line.
				}

				// Advance.
				strippedTagStart := nextTag[tagIndex][0] - originalPos - totalTagLength
				tagEnd = nextTag[tagIndex][1]
				tagLength := tagEnd - nextTag[tagIndex][0]
				if nextTag[tagIndex][2] == 2 {
					tagLength = 1
				}
				totalTagLength += tagLength
				remainingLength = lineLength - (tagEnd - originalPos - totalTagLength)

				// Process the tag.
				switch nextTag[tagIndex][2] {
				case 0:
					// Process color tags.
					foregroundColor, backgroundColor, attributes = styleFromTag(foregroundColor, backgroundColor, attributes, colorTags[colorPos])
					colorPos++
				case 1:
					// Process region tags.
					regionID = regions[regionPos][1]
					_, highlighted = t.highlights[regionID]

					// Update highlight range.
					if highlighted {
						line := len(t.index)
						if t.fromHighlight < 0 {
							t.fromHighlight, t.toHighlight = line, line
							t.posHighlight = runewidth.StringWidth(splitLine[:strippedTagStart])
						} else if line > t.toHighlight {
							t.toHighlight = line
						}
					}

					regionPos++
				case 2:
					// Process escape tags.
					escapePos++
				}
			}

			// Advance to next line.
			originalPos += lineLength + totalTagLength

			// Append this line.
			line.NextPos = originalPos
			line.Width = runewidth.StringWidth(splitLine)
			t.index = append(t.index, line)
		}

		// Word-wrapped lines may have trailing whitespace. Remove it.
		if t.wrap && t.wordWrap {
			for _, line := range t.index {
				str := t.buffer[line.Line][line.Pos:line.NextPos]
				spaces := spacePattern.FindAllStringIndex(str, -1)
				if spaces != nil && spaces[len(spaces)-1][1] == len(str) {
					oldNextPos := line.NextPos
					line.NextPos -= spaces[len(spaces)-1][1] - spaces[len(spaces)-1][0]
					line.Width -= runewidth.StringWidth(t.buffer[line.Line][line.NextPos:oldNextPos])
				}
			}
		}
	}

	// Calculate longest line.
	t.longestLine = 0
	for _, line := range t.index {
		if line.Width > t.longestLine {
			t.longestLine = line.Width
		}
	}
}

// Draw draws this primitive onto the screen.
func (t *TextView) Draw(screen tcell.Screen) {
	t.Lock()
	defer t.Unlock()
	t.Box.Draw(screen)

	// Get the available size.
	x, y, width, height := t.GetInnerRect()
	t.pageSize = height

	// If the width has changed, we need to reindex.
	if width != t.lastWidth && t.wrap {
		t.index = nil
	}
	t.lastWidth = width

	// Re-index.
	t.reindexBuffer(width)

	// If we don't have an index, there's nothing to draw.
	if t.index == nil {
		return
	}

	// Move to highlighted regions.
	if t.regions && t.scrollToHighlights && t.fromHighlight >= 0 {
		// Do we fit the entire height?
		if t.toHighlight-t.fromHighlight+1 < height {
			// Yes, let's center the highlights.
			t.lineOffset = (t.fromHighlight + t.toHighlight - height) / 2
		} else {
			// No, let's move to the start of the highlights.
			t.lineOffset = t.fromHighlight
		}

		// If the highlight is too far to the right, move it to the middle.
		if t.posHighlight-t.columnOffset > 3*width/4 {
			t.columnOffset = t.posHighlight - width/2
		}

		// If the highlight is off-screen on the left, move it on-screen.
		if t.posHighlight-t.columnOffset < 0 {
			t.columnOffset = t.posHighlight - width/4
		}
	}
	t.scrollToHighlights = false

	// Adjust line offset.
	if t.lineOffset+height > len(t.index) {
		t.trackEnd = true
	}
	if t.trackEnd {
		t.lineOffset = len(t.index) - height
	}
	if t.lineOffset < 0 {
		t.lineOffset = 0
	}

	// Adjust column offset.
	if t.align == AlignLeft {
		if t.columnOffset+width > t.longestLine {
			t.columnOffset = t.longestLine - width
		}
		if t.columnOffset < 0 {
			t.columnOffset = 0
		}
	} else if t.align == AlignRight {
		if t.columnOffset-width < -t.longestLine {
			t.columnOffset = width - t.longestLine
		}
		if t.columnOffset > 0 {
			t.columnOffset = 0
		}
	} else { // AlignCenter.
		half := (t.longestLine - width) / 2
		if half > 0 {
			if t.columnOffset > half {
				t.columnOffset = half
			}
			if t.columnOffset < -half {
				t.columnOffset = -half
			}
		} else {
			t.columnOffset = 0
		}
	}

	// Draw the buffer.
	defaultStyle := tcell.StyleDefault.Foreground(t.textColor)
	for line := t.lineOffset; line < len(t.index); line++ {
		// Are we done?
		if line-t.lineOffset >= height {
			break
		}

		// Get the text for this line.
		index := t.index[line]
		text := t.buffer[index.Line][index.Pos:index.NextPos]
		foregroundColor := index.ForegroundColor
		backgroundColor := index.BackgroundColor
		attributes := index.Attributes
		regionID := index.Region

		// Get color tags.
		var (
			colorTagIndices [][]int
			colorTags       [][]string
			escapeIndices   [][]int
		)
		strippedText := text
		if t.dynamicColors {
			colorTagIndices, colorTags, escapeIndices, strippedText, _ = decomposeString(text)
		}

		// Get regions.
		var (
			regionIndices [][]int
			regions       [][]string
		)
		if t.regions {
			regionIndices = regionPattern.FindAllStringIndex(text, -1)
			regions = regionPattern.FindAllStringSubmatch(text, -1)
			strippedText = regionPattern.ReplaceAllString(strippedText, "")
			if !t.dynamicColors {
				escapeIndices = escapePattern.FindAllStringIndex(text, -1)
				strippedText = string(escapePattern.ReplaceAllString(strippedText, "[$1$2]"))
			}
		}

		// Calculate the position of the line.
		var skip, posX int
		if t.align == AlignLeft {
			posX = -t.columnOffset
		} else if t.align == AlignRight {
			posX = width - index.Width - t.columnOffset
		} else { // AlignCenter.
			posX = (width-index.Width)/2 - t.columnOffset
		}
		if posX < 0 {
			skip = -posX
			posX = 0
		}

		// Print the line.
		var colorPos, regionPos, escapePos, tagOffset, skipped int
		iterateString(strippedText, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
			// Process tags.
			for {
				if colorPos < len(colorTags) && textPos+tagOffset >= colorTagIndices[colorPos][0] && textPos+tagOffset < colorTagIndices[colorPos][1] {
					// Get the color.
					foregroundColor, backgroundColor, attributes = styleFromTag(foregroundColor, backgroundColor, attributes, colorTags[colorPos])
					tagOffset += colorTagIndices[colorPos][1] - colorTagIndices[colorPos][0]
					colorPos++
				} else if regionPos < len(regionIndices) && textPos+tagOffset >= regionIndices[regionPos][0] && textPos+tagOffset < regionIndices[regionPos][1] {
					// Get the region.
					regionID = regions[regionPos][1]
					tagOffset += regionIndices[regionPos][1] - regionIndices[regionPos][0]
					regionPos++
				} else {
					break
				}
			}

			// Skip the second-to-last character of an escape tag.
			if escapePos < len(escapeIndices) && textPos+tagOffset == escapeIndices[escapePos][1]-2 {
				tagOffset++
				escapePos++
			}

			// Mix the existing style with the new style.
			_, _, existingStyle, _ := screen.GetContent(x+posX, y+line-t.lineOffset)
			_, background, _ := existingStyle.Decompose()
			style := overlayStyle(background, defaultStyle, foregroundColor, backgroundColor, attributes)

			// Do we highlight this character?
			var highlighted bool
			if len(regionID) > 0 {
				if _, ok := t.highlights[regionID]; ok {
					highlighted = true
				}
			}
			if highlighted {
				fg, bg, _ := style.Decompose()
				if bg == tcell.ColorDefault {
					r, g, b := fg.RGB()
					c := colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}
					_, _, li := c.Hcl()
					if li < .5 {
						bg = tcell.ColorWhite
					} else {
						bg = tcell.ColorBlack
					}
				}
				style = style.Background(fg).Foreground(bg)
			}

			// Skip to the right.
			if !t.wrap && skipped < skip {
				skipped += screenWidth
				return false
			}

			// Stop at the right border.
			if posX+screenWidth > width {
				return true
			}

			// Draw the character.
			for offset := screenWidth - 1; offset >= 0; offset-- {
				if offset == 0 {
					screen.SetContent(x+posX+offset, y+line-t.lineOffset, main, comb, style)
				} else {
					screen.SetContent(x+posX+offset, y+line-t.lineOffset, ' ', nil, style)
				}
			}

			// Advance.
			posX += screenWidth
			return false
		})
	}

	// If this view is not scrollable, we'll purge the buffer of lines that have
	// scrolled out of view.
	if !t.scrollable && t.lineOffset > 0 {
		t.buffer = t.buffer[t.index[t.lineOffset].Line:]
		t.index = nil
	}
}

// InputHandler returns the handler for this primitive.
func (t *TextView) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		key := event.Key()

		if key == tcell.KeyEscape || key == tcell.KeyEnter || key == tcell.KeyTab || key == tcell.KeyBacktab {
			if t.done != nil {
				t.done(key)
			}
			return
		}

		if !t.scrollable {
			return
		}

		switch key {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // Home.
				t.ScrollToBeginning()
			case 'G': // End.
				t.ScrollToEnd()
			case 'j': // Down.
				t.lineOffset++
			case 'k': // Up.
				t.trackEnd = false
				t.lineOffset--
			case 'h': // Left.
				t.columnOffset--
			case 'l': // Right.
				t.columnOffset++
			}
		case tcell.KeyHome:
			t.ScrollToBeginning()
		case tcell.KeyEnd:
			t.ScrollToEnd()
		case tcell.KeyUp:
			t.trackEnd = false
			t.lineOffset--
		case tcell.KeyDown:
			t.lineOffset++
		case tcell.KeyLeft:
			t.columnOffset--
		case tcell.KeyRight:
			t.columnOffset++
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			t.PageDown()
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			t.PageUp()
		}
	})
}

// PageUp move up a full page.
func (t *TextView) PageUp() {
	t.trackEnd = false
	t.lineOffset -= t.pageSize
}

// PageDown move down a full page.
func (t *TextView) PageDown() {
	t.lineOffset += t.pageSize
}
