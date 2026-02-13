package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sushantvema-harper/linear-tui/internal/linearapi"
)

// CommentsBrowser provides a fullscreen modal for browsing issue comments one at a time.
type CommentsBrowser struct {
	app          *App
	modal        *tview.Flex
	contentView  *tview.TextView
	headerView   *tview.TextView
	helpView     *tview.TextView
	searchInput  *tview.InputField
	comments     []linearapi.Comment
	filtered     []linearapi.Comment
	currentIndex int
	searchActive bool
	searchQuery  string
	issueURL     string
	issueID      string
}

// NewCommentsBrowser creates a new comments browser.
func NewCommentsBrowser(app *App) *CommentsBrowser {
	cb := &CommentsBrowser{app: app}

	// Header: 1-line showing index, author, date
	cb.headerView = tview.NewTextView()
	cb.headerView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(app.theme.HeaderBg)

	// Content: scrollable markdown-rendered comment body
	cb.contentView = tview.NewTextView()
	cb.contentView.SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true).
		SetBorder(false).
		SetBackgroundColor(app.theme.Background)
	cb.contentView.SetBorderPadding(1, 1, 2, 2)

	// Help: 1-line key hints
	cb.helpView = tview.NewTextView()
	cb.helpView.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(app.theme.HeaderBg)
	cb.helpView.SetText(fmt.Sprintf("%sn/p: next/prev | Ctrl+N/P: scroll | Ctrl+D/U: half-page | /: search | y: copy link | o: open | t: comment | Esc: close[-]", app.themeTags.SecondaryText))

	// Search input (hidden by default)
	cb.searchInput = tview.NewInputField()
	cb.searchInput.SetLabel(" Filter: ")
	cb.searchInput.SetFieldBackgroundColor(app.theme.InputBg)
	cb.searchInput.SetFieldTextColor(app.theme.Foreground)
	cb.searchInput.SetLabelColor(app.theme.Accent)
	cb.searchInput.SetBackgroundColor(app.theme.HeaderBg)
	cb.searchInput.SetChangedFunc(func(text string) {
		cb.searchQuery = text
		cb.applyFilter()
		cb.render()
	})
	cb.searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			cb.toggleSearch()
		}
	})

	// Build layout: header + content + help
	cb.modal = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(cb.headerView, 1, 0, false).
		AddItem(cb.contentView, 0, 1, true).
		AddItem(cb.helpView, 1, 0, false)
	cb.modal.SetBackgroundColor(app.theme.Background)

	return cb
}

// Show displays the comments browser with the given comments.
func (cb *CommentsBrowser) Show(comments []linearapi.Comment, issueURL string, issueID string) {
	cb.comments = comments
	cb.filtered = comments
	cb.currentIndex = 0
	cb.searchActive = false
	cb.searchQuery = ""
	cb.issueURL = issueURL
	cb.issueID = issueID

	// Reset search input
	cb.searchInput.SetText("")

	// Ensure search row is hidden
	cb.rebuildLayout(false)

	cb.render()

	cb.app.commentsBrowserActive = true
	cb.app.pages.AddPage("comments_browser", cb.modal, true, true)
	cb.app.app.SetFocus(cb.contentView)
}

// Hide closes the comments browser.
func (cb *CommentsBrowser) Hide() {
	cb.app.commentsBrowserActive = false
	cb.app.pages.RemovePage("comments_browser")
	cb.app.updateFocus()
}

// render updates the header and content for the current comment.
func (cb *CommentsBrowser) render() {
	if len(cb.filtered) == 0 {
		cb.headerView.SetText(fmt.Sprintf("%sNo comments%s[-]",
			cb.app.themeTags.SecondaryText,
			func() string {
				if cb.searchQuery != "" {
					return " matching filter"
				}
				return ""
			}()))
		cb.contentView.Clear()
		cb.contentView.SetText(fmt.Sprintf("%sNo comments to display.[-]", cb.app.themeTags.SecondaryText))
		return
	}

	comment := cb.filtered[cb.currentIndex]

	// Header: "Comment 3/15 | Author Name | Jan 15, 2024 (edited)"
	authorDisplay := comment.Author.DisplayName
	if authorDisplay == "" {
		authorDisplay = comment.Author.Name
	}
	if comment.Author.IsMe {
		authorDisplay += " (me)"
	}
	timeStr := comment.CreatedAt.Format("Jan 2, 2006")
	if !comment.UpdatedAt.Equal(comment.CreatedAt) {
		timeStr += " (edited)"
	}
	headerText := fmt.Sprintf("%sComment %d/%d[-] %s| %s%s[-] %s| %s%s[-]",
		cb.app.themeTags.Accent,
		cb.currentIndex+1, len(cb.filtered),
		cb.app.themeTags.Border,
		cb.app.themeTags.Foreground, authorDisplay,
		cb.app.themeTags.Border,
		cb.app.themeTags.SecondaryText, timeStr)
	cb.headerView.SetText(headerText)

	// Content: rendered markdown
	cb.contentView.Clear()
	writer := tview.ANSIWriter(cb.contentView)
	rendered := renderMarkdown(comment.Body)
	_, _ = fmt.Fprint(writer, rendered)
	cb.contentView.ScrollToBeginning()
}

// HandleKey handles keyboard input for the comments browser.
func (cb *CommentsBrowser) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	// When search is active, only intercept specific keys
	if cb.searchActive {
		switch event.Key() {
		case tcell.KeyEscape:
			cb.toggleSearch()
			return nil
		case tcell.KeyEnter:
			cb.toggleSearch()
			return nil
		case tcell.KeyCtrlC:
			cb.app.app.Stop()
			return nil
		}
		// Let all other keys flow to the focused search input
		return event
	}

	switch event.Key() {
	case tcell.KeyEscape:
		cb.Hide()
		return nil
	case tcell.KeyCtrlN:
		row, col := cb.contentView.GetScrollOffset()
		cb.contentView.ScrollTo(row+1, col)
		return nil
	case tcell.KeyCtrlP:
		row, col := cb.contentView.GetScrollOffset()
		if row > 0 {
			cb.contentView.ScrollTo(row-1, col)
		}
		return nil
	case tcell.KeyCtrlD:
		_, _, _, h := cb.contentView.GetInnerRect()
		delta := max(1, h/2)
		row, col := cb.contentView.GetScrollOffset()
		cb.contentView.ScrollTo(row+delta, col)
		return nil
	case tcell.KeyCtrlU:
		_, _, _, h := cb.contentView.GetInnerRect()
		delta := max(1, h/2)
		row, col := cb.contentView.GetScrollOffset()
		targetRow := row - delta
		if targetRow < 0 {
			targetRow = 0
		}
		cb.contentView.ScrollTo(targetRow, col)
		return nil
	case tcell.KeyCtrlC:
		cb.app.app.Stop()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'n', 'j':
			if cb.currentIndex < len(cb.filtered)-1 {
				cb.currentIndex++
				cb.render()
			}
			return nil
		case 'p', 'k':
			if cb.currentIndex > 0 {
				cb.currentIndex--
				cb.render()
			}
			return nil
		case 'q':
			cb.Hide()
			return nil
		case '/':
			cb.toggleSearch()
			return nil
		case 'y':
			if len(cb.filtered) > 0 {
				comment := cb.filtered[cb.currentIndex]
				url := comment.URL
				_ = copyToClipboard(url)
				cb.app.showToast("Copied comment link")
			}
			return nil
		case 'o':
			if len(cb.filtered) > 0 {
				comment := cb.filtered[cb.currentIndex]
				url := comment.URL
				_ = openURL(url)
			}
			return nil
		case 't':
			// Open create comment modal
			if cb.issueID != "" {
				cb.Hide()
				cb.app.createCommentModal.Show(cb.issueID, cb.app.handleCreateComment)
			}
			return nil
		}
	}

	// Consume all other keys
	return nil
}

// applyFilter filters comments by body text and author name.
func (cb *CommentsBrowser) applyFilter() {
	if cb.searchQuery == "" {
		cb.filtered = cb.comments
		cb.currentIndex = 0
		return
	}

	query := strings.ToLower(cb.searchQuery)
	filtered := make([]linearapi.Comment, 0)
	for _, c := range cb.comments {
		bodyMatch := strings.Contains(strings.ToLower(c.Body), query)
		authorMatch := strings.Contains(strings.ToLower(c.Author.Name), query) ||
			strings.Contains(strings.ToLower(c.Author.DisplayName), query)
		if bodyMatch || authorMatch {
			filtered = append(filtered, c)
		}
	}
	cb.filtered = filtered
	cb.currentIndex = 0
}

// toggleSearch shows/hides the search input.
func (cb *CommentsBrowser) toggleSearch() {
	cb.searchActive = !cb.searchActive

	if cb.searchActive {
		cb.rebuildLayout(true)
		cb.app.app.SetFocus(cb.searchInput)
	} else {
		// Clear search and restore full list
		cb.searchQuery = ""
		cb.searchInput.SetText("")
		cb.filtered = cb.comments
		cb.currentIndex = 0
		cb.rebuildLayout(false)
		cb.app.app.SetFocus(cb.contentView)
		cb.render()
	}
}

// rebuildLayout reconstructs the modal layout with or without the search row.
func (cb *CommentsBrowser) rebuildLayout(showSearch bool) {
	cb.modal.Clear()
	cb.modal.AddItem(cb.headerView, 1, 0, false)
	if showSearch {
		cb.modal.AddItem(cb.searchInput, 1, 0, false)
	}
	cb.modal.AddItem(cb.contentView, 0, 1, true)
	cb.modal.AddItem(cb.helpView, 1, 0, false)
}
