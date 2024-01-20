package components

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/erroneousboat/termui"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/erroneousboat/slack-term/config"
	// "github.com/adrg/xdg"
)

// Chat is the definition of a Chat component
type Chat struct {
	List           *termui.List
	Messages       map[string]Message
	MessagesSorted []Message
	Oldest         time.Time
	OldestTs       string
	Offset         int
	UserNameStyle  string
	TempMsgCnt     int
}

type Line struct {
	cells []termui.Cell
}

// CreateChatComponent is the constructor for the Chat struct
func CreateChatComponent(inputHeight int) *Chat {
	chat := &Chat{
		List:     termui.NewList(),
		Messages: make(map[string]Message),
		Offset:   0,
	}

	chat.List.Height = termui.TermHeight() - inputHeight
	chat.List.Overflow = "wrap"

	return chat
}

func (c *Chat) GetOldest() (time.Time, string) {
	ts := ""
	time := time.Now()
	for _, msg := range c.Messages {
		if time.Compare(msg.Time) > 0 {
			time = msg.Time
			ts = msg.RawTs
		}
	}
	return time, ts
}

// Buffer implements interface termui.Bufferer
func (c *Chat) Buffer() termui.Buffer {
	// Convert Messages into termui.Cell
	// // TODO
	// // Limit to the messages on screen
	// cells := c.MessagesToCells(c.MessagesSorted)
	//
	// // We will create an array of Line structs, this allows us
	// // to more easily render the items in a list. We will range
	// // over the cells we've created and create a Line within
	// // the bounds of the Chat pane
	//
	// lines := []Line{}
	// line := Line{}
	//
	// // When we encounter a newline or, are at the bounds of the chat view we
	// // stop iterating over the cells and add the line to the line array
	// x := 0
	// for _, cell := range cells {
	//
	// 	// When we encounter a newline we add the line to the array
	// 	if cell.Ch == '\n' {
	// 		lines = append(lines, line)
	//
	// 		// Reset for new line
	// 		line = Line{}
	// 		x = 0
	// 		continue
	// 	}
	//
	// 	if x+cell.Width() > c.List.InnerBounds().Dx() {
	// 		lines = append(lines, line)
	//
	// 		// Reset for new line
	// 		line = Line{}
	// 		x = 0
	// 	}
	//
	// 	line.cells = append(line.cells, cell)
	// 	x += cell.Width()
	// }
	//
	// // Append the last line to the array when we didn't encounter any
	// // newlines or were at the bounds of the chat view
	// lines = append(lines, line)

	// We will print lines bottom up, it will loop over the lines
	// backwards and for every line it'll set the cell in that line.
	// Offset is the number which allows us to begin printing the
	// line above the last line.

	buf := c.List.Buffer()
	paneMinY := c.List.InnerBounds().Min.Y
	paneMaxY := c.List.InnerBounds().Max.Y
	currentY := paneMaxY - 1
	offset := c.Offset
	msgsLen := len(c.MessagesSorted)

	for j := (msgsLen - 1); j >= 0; j-- {
		m := c.MessagesSorted[j]
		linesHeight := len(m.Lines)
		for i := (linesHeight - 1) - offset; i >= 0; i-- {
			if currentY < paneMinY {
				break
			}
			x := c.List.InnerBounds().Min.X
			for _, cell := range m.Lines[i].cells {
				buf.Set(x, currentY, cell)
				x += cell.Width()
			}

			// When we're not at the end of the pane, fill it up
			// with empty characters
			for x < c.List.InnerBounds().Max.X {
				buf.Set(
					x, currentY,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemFgColor,
						Bg: c.List.ItemBgColor,
					},
				)
				x += runewidth.RuneWidth(' ')
			}
			currentY--
		}
		if currentY < paneMinY {
			break
		}
		offset -= linesHeight
		if offset < 0 {
			offset = 0
		}
	}

	// If the space above currentY is empty we need to fill
	// it up with blank lines, otherwise the List object will
	// render the items top down, and the result will mix.
	for currentY >= paneMinY {
		x := c.List.InnerBounds().Min.X
		for x < c.List.InnerBounds().Max.X {
			buf.Set(
				x, currentY,
				termui.Cell{
					Ch: ' ',
					Fg: c.List.ItemFgColor,
					Bg: c.List.ItemBgColor,
				},
			)
			x += runewidth.RuneWidth(' ')
		}
		currentY--
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Chat) GetHeight() int {
	return c.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Chat) SetWidth(w int) {
	c.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Chat) SetX(x int) {
	c.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Chat) SetY(y int) {
	c.List.SetY(y)
}

func CalcLink(i int) string {
	base_links := []string{"a", "s", "d", "f", "g", "h", "j", "k", "l", "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "z", "x", "c", "v", "b", "n", "m"}
	base_len := len(base_links)
	link := ""
	for i >= 0 {
		ind := i % base_len
		link = link + base_links[ind]
		if i > base_len {
			i -= base_len
		} else {
			i = -1
		}
	}
	return link
}

func (c *Chat) ClearIdents() {
	for i, msg := range c.MessagesSorted {
		msg.Link = ""
		c.MessagesSorted[i] = msg
	}
}

type IdentRun struct {
	link_cnt  int
	last_link [2]string
	line_cnt  int
	j         int
	mode      string
	input     rune
}

func (c *Chat) RunIdent(msg Message, run IdentRun, i string, j string) (Message, IdentRun) {
	run.line_cnt += len(msg.Lines)
	// if run.mode == "download" && j == "" {
	if run.mode == "download" && (len(j) < 5 || !strings.Contains(msg.Content, "files.slack")) {
		return msg, run
	}
	log.Println("i: ", i, "j: ", j, "cont: ", msg.Content)
	if run.input == rune(0) {
		msg.Link = CalcLink(run.j)
		run.j += 1
	} else if rune(msg.Link[0]) == run.input {
		run.link_cnt += 1
		run.last_link[0] = i
		run.last_link[1] = j
		msg.Link = msg.Link[1:]
	} else {
		msg.Link = ""
	}
	msg.Lines = c.MessageToLines(msg)
	return msg, run
}

func (c *Chat) SetIdents(input rune, mode string) [2]string {
	var run IdentRun
	run.line_cnt = 0
	run.link_cnt = 0
	run.last_link = [2]string{"", ""}
	run.j = 0
	run.input = input
	run.mode = mode
	for i := len(c.MessagesSorted) - 1; i >= 0; i-- {
		msg := c.MessagesSorted[i]
		msg, new_run := c.RunIdent(msg, run, msg.ID, "")
		run = new_run
		ran_sub := false
		for key, sub := range msg.Messages {
			ran_sub = true
			new_sub, new_run := c.RunIdent(sub, run, msg.ID, key)
			run = new_run
			msg.Messages[key] = new_sub
		}
		if ran_sub {
			msg.Lines = c.MessageToLines(msg)
		}
		c.MessagesSorted[i] = msg
		if run.line_cnt >= c.GetMaxItems() {
			break
		}
	}
	if run.link_cnt == 1 {
		return run.last_link
	} else {
		return [2]string{"", ""}
	}
}

// GetMaxItems return the maximal amount of items can fit in the Chat
// component
func (c *Chat) GetMaxItems() int {
	return c.List.InnerBounds().Max.Y - c.List.InnerBounds().Min.Y
}

// func (c *Chat) SearchNext(message Message) {
//
//
// }

func (c *Chat) SetUserNameStyle(username string) {
	for _, msg := range c.MessagesSorted {
		if msg.Name == username {
			c.UserNameStyle = msg.StyleName
			return
		}
	}
}

// SetMessages will put the provided messages into the Messages field of the
// Chat view
func (c *Chat) SetMessages(messages []Message) {
	// Reset offset first, when scrolling in view and changing channels we
	// want the offset to be 0 when loading new messages
	c.Offset = 0
	c.TempMsgCnt = 0
	for _, msg := range messages {
		msg.Lines = c.MessageToLines(msg)
		c.Messages[msg.ID] = msg
		c.MessagesSorted = append(c.MessagesSorted, msg)
	}
	c.Oldest, c.OldestTs = c.GetOldest()
}

// AddMessage adds a single message to Messages
func (c *Chat) AddRawMessage(raw string, name string, thread string) {
	var message Message
	message.ID = fmt.Sprint(c.TempMsgCnt)
	c.TempMsgCnt += 1
	message.Thread = thread
	message.Content = raw
	message.StyleText = c.MessagesSorted[0].StyleText
	message.Time = time.Now()
	message.FormatTime = c.MessagesSorted[0].FormatTime
	message.StyleTime = c.MessagesSorted[0].StyleTime
	message.Name = name
	message.StyleName = c.UserNameStyle
	message.Lines = c.MessageToLines(message)
	c.Messages[message.ID] = message
	c.MessagesSorted = append(c.MessagesSorted, message)
	c.Oldest, c.OldestTs = c.GetOldest()
}

// AddMessage adds a single message to Messages
func (c *Chat) DelRawMessage() {
	delete(c.Messages, "0")
}

// AddMessage adds a single message to Messages
func (c *Chat) AddMessage(message Message) {
	message.Lines = c.MessageToLines(message)
	c.Messages[message.ID] = message
	t := c.MessagesSorted[len(c.MessagesSorted)-1].Time
	// prepend as it happened after
	if t.Compare(message.Time) < 0 {
		c.MessagesSorted = append(c.MessagesSorted, Message{})
		copy(c.MessagesSorted[1:], c.MessagesSorted[0:])
		c.MessagesSorted[0] = message
	} else {
		c.MessagesSorted = append(c.MessagesSorted, message)
	}
	// this shouldn't be necessary
	if c.Oldest.Compare(message.Time) > 0 {
		c.OldestTs = message.RawTs
	}
}

// AddMessage adds a single message to Messages
func (c *Chat) AddReplyMessage(message Message) {
	t := message.Time
	ind := -1
	for i, m := range c.MessagesSorted {
		if t.Compare(m.Time) > -1 {
			ind = i
			break
		}
	}
	if ind < 0 {
		return
	}

	c.MessagesSorted = append(c.MessagesSorted, Message{})
	copy(c.MessagesSorted[ind+1:], c.MessagesSorted[ind:])
	c.MessagesSorted[ind] = message
	c.Messages[message.ID] = message
	// this shouldn't be necessary
	if c.Oldest.Compare(message.Time) > 0 {
		c.OldestTs = message.RawTs
	}
}

// AddReply adds a single reply to a parent thread, it also sets
// the thread separator
func (c *Chat) AddReply(parentID string, message Message) {
	// It is possible that a message is received but the parent is not
	// present in the chat view
	if _, ok := c.Messages[parentID]; ok {
		message.Thread = "  "
		message.Lines = c.MessageToLines(message)
		c.Messages[parentID].Messages[message.ID] = message
		c.AddReplyMessage(message)
	} else {
		c.AddMessage(message)
	}
}

// IsNewThread check whether a message that is going to be added as
// a child to a parent message, is the first one or not
func (c *Chat) IsNewThread(parentID string) bool {
	if parent, ok := c.Messages[parentID]; ok {
		if len(parent.Messages) > 0 {
			return true
		}
	}
	return false
}

// ClearMessages clear the c.Messages
func (c *Chat) ClearMessages() {
	c.Messages = make(map[string]Message)
	c.MessagesSorted = make([]Message, 0)
}

// ScrollUp will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in substracting the offset
// from the len(Chat.Messages).
func (c *Chat) ScrollUp(n int) (bool, int) {
	c.Offset = c.Offset + n
	l := len(c.Messages)
	// Protect overscrolling
	if c.Offset > l {
		c.Offset = len(c.Messages)
	}
	if c.Offset >= 2*l/3 {
		return true, l
	}
	return false, l
}

// ScrollDown will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in substracting the offset
// from the len(Chat.Messages).
func (c *Chat) ScrollDown(n int) {
	c.Offset = c.Offset - n

	// Protect overscrolling
	if c.Offset < 0 {
		c.Offset = 0
	}
}

func (c *Chat) MoveCursorTop() {
	c.Offset = len(c.Messages)
}

func (c *Chat) MoveCursorBottom() {
	c.Offset = 0
}

// SetBorderLabel will set Label of the Chat pane to the specified string
func (c *Chat) SetBorderLabel(channelName string) {
	c.List.BorderLabel = channelName
}

// MessagesToCellsUnsorted is a wrapper around MessagesToCells for unsorted messages
func (c *Chat) MessagesToCellsUnsorted(msgs map[string]Message) []termui.Cell {
	sortedMessages := SortMessages(msgs)
	cells := c.MessagesToCells(sortedMessages)
	return cells
}

func (c *Chat) CellsToLines(cells []termui.Cell) []Line {
	lines := make([]Line, 0)
	x := 0
	line := Line{}
	for _, cell := range cells {
		// When we encounter a newline we add the line to the array
		if cell.Ch == '\n' {
			lines = append(lines, line)
			// Reset for new line
			line = Line{}
			x = 0
			continue
		}
		if x+cell.Width() > c.List.InnerBounds().Dx() {
			lines = append(lines, line)
			// Reset for new line
			line = Line{}
			x = 0
		}
		line.cells = append(line.cells, cell)
		x += cell.Width()
	}
	if len(line.cells) > 0 {
		lines = append(lines, line)
	}
	return lines
}

func (c *Chat) MessageToLines(msg Message) []Line {
	lines := make([]Line, 0)

	cells := c.MessageToCells(msg)
	lines = append(lines, c.CellsToLines(cells)...)

	for _, m := range msg.Messages {
		lines = append(lines, c.MessageToLines(m)...)
	}

	return lines
}

// MessagesToCells is a wrapper around MessageToCells to use for a slice of
// of type Message
func (c *Chat) MessagesToCells(msgs []Message) []termui.Cell {
	cells := make([]termui.Cell, 0)

	for i, msg := range msgs {
		cells = append(cells, c.MessageToCells(msg)...)

		if len(msg.Messages) > 0 {
			cells = append(cells, termui.Cell{Ch: '\n'})
			cells = append(cells, c.MessagesToCellsUnsorted(msg.Messages)...)
		}

		// Add a newline after every message
		if i < len(msgs)-1 {
			cells = append(cells, termui.Cell{Ch: '\n'})
		}
	}

	return cells
}

// MessageToCells will convert a Message struct to termui.Cell
//
// We're building parts of the message individually, or else DefaultTxBuilder
// will interpret potential markdown usage in a message as well.
func (c *Chat) MessageToCells(msg Message) []termui.Cell {
	cells := make([]termui.Cell, 0)

	if msg.Link != "" {
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetLink(),
			termui.ColorDefault, termui.ColorDefault)...,
		)
	}

	// When msg.Time and msg.Name are empty (in the case of attachments)
	// don't add the time and name parts.
	if (msg.Time != time.Time{} && msg.Name != "") {
		// Time
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetTime(),
			termui.ColorDefault, termui.ColorDefault)...,
		)

		// Thread
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetThread(),
			termui.ColorDefault, termui.ColorDefault)...,
		)

		// Name
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetName(),
			termui.ColorDefault, termui.ColorDefault)...,
		)
	}

	// Hack, in order to get the correct fg and bg attributes. This is
	// because the readAttr function in termui is unexported.
	txCells := termui.DefaultTxBuilder.Build(
		msg.GetContent(),
		termui.ColorDefault, termui.ColorDefault,
	)

	// Text
	for _, r := range msg.Content {
		cells = append(
			cells,
			termui.Cell{
				Ch: r,
				Fg: txCells[0].Fg,
				Bg: txCells[0].Bg,
			},
		)
	}

	return cells
}

// Help shows the usage and key bindings in the chat pane
func (c *Chat) Help(usage string, cfg *config.Config) {
	msgUsage := Message{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Content: usage,
	}

	c.Messages[msgUsage.ID] = msgUsage

	for mode, mapping := range cfg.KeyMap {
		msgMode := Message{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Content: fmt.Sprintf("%s", strings.ToUpper(mode)),
		}
		c.Messages[msgMode.ID] = msgMode

		msgNewline := Message{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Content: "",
		}
		c.Messages[msgNewline.ID] = msgNewline

		var keys []string
		for k := range mapping {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			msgKey := Message{
				ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
				Content: fmt.Sprintf("    %-12s%-15s", k, mapping[k]),
			}
			c.Messages[msgKey.ID] = msgKey
		}

		msgNewline.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		c.Messages[msgNewline.ID] = msgNewline
	}
}
