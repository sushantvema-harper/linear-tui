package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sushantvema-harper/linear-tui/internal/linearapi"
)

// AttachmentsBrowser provides a fullscreen modal for browsing issue attachments one at a time.
type AttachmentsBrowser struct {
	app          *App
	modal        *tview.Flex
	contentView  *tview.TextView
	headerView   *tview.TextView
	helpView     *tview.TextView
	searchInput  *tview.InputField
	attachments  []linearapi.Attachment
	filtered     []linearapi.Attachment
	currentIndex int
	searchActive bool
	searchQuery  string
}

// NewAttachmentsBrowser creates a new attachments browser.
func NewAttachmentsBrowser(app *App) *AttachmentsBrowser {
	ab := &AttachmentsBrowser{app: app}

	// Header: 1-line showing index
	ab.headerView = tview.NewTextView()
	ab.headerView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(app.theme.HeaderBg)

	// Content: attachment metadata display
	ab.contentView = tview.NewTextView()
	ab.contentView.SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true).
		SetBorder(false).
		SetBackgroundColor(app.theme.Background)
	ab.contentView.SetBorderPadding(1, 1, 2, 2)

	// Help: 1-line key hints
	ab.helpView = tview.NewTextView()
	ab.helpView.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(app.theme.HeaderBg)
	ab.helpView.SetText(fmt.Sprintf("%sn/p: next/prev | Ctrl+N/P: scroll | /: search | y: copy URL | o: open | Esc: close[-]", app.themeTags.SecondaryText))

	// Search input (hidden by default)
	ab.searchInput = tview.NewInputField()
	ab.searchInput.SetLabel(" Filter: ")
	ab.searchInput.SetFieldBackgroundColor(app.theme.InputBg)
	ab.searchInput.SetFieldTextColor(app.theme.Foreground)
	ab.searchInput.SetLabelColor(app.theme.Accent)
	ab.searchInput.SetBackgroundColor(app.theme.HeaderBg)
	ab.searchInput.SetChangedFunc(func(text string) {
		ab.searchQuery = text
		ab.applyFilter()
		ab.render()
	})
	ab.searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			ab.toggleSearch()
		}
	})

	// Build layout: header + content + help
	ab.modal = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ab.headerView, 1, 0, false).
		AddItem(ab.contentView, 0, 1, true).
		AddItem(ab.helpView, 1, 0, false)
	ab.modal.SetBackgroundColor(app.theme.Background)

	return ab
}

// Show displays the attachments browser with the given attachments.
func (ab *AttachmentsBrowser) Show(attachments []linearapi.Attachment) {
	ab.attachments = attachments
	ab.filtered = attachments
	ab.currentIndex = 0
	ab.searchActive = false
	ab.searchQuery = ""

	// Reset search input
	ab.searchInput.SetText("")

	// Ensure search row is hidden
	ab.rebuildLayout(false)

	ab.render()

	ab.app.attachmentsBrowserActive = true
	ab.app.pages.AddPage("attachments_browser", ab.modal, true, true)
	ab.app.app.SetFocus(ab.contentView)
}

// Hide closes the attachments browser.
func (ab *AttachmentsBrowser) Hide() {
	ab.app.attachmentsBrowserActive = false
	ab.app.pages.RemovePage("attachments_browser")
	ab.app.updateFocus()
}

// render updates the header and content for the current attachment.
func (ab *AttachmentsBrowser) render() {
	if len(ab.filtered) == 0 {
		ab.headerView.SetText(fmt.Sprintf("%sNo attachments%s[-]",
			ab.app.themeTags.SecondaryText,
			func() string {
				if ab.searchQuery != "" {
					return " matching filter"
				}
				return ""
			}()))
		ab.contentView.Clear()
		ab.contentView.SetText(fmt.Sprintf("%sNo attachments to display.[-]", ab.app.themeTags.SecondaryText))
		return
	}

	attach := ab.filtered[ab.currentIndex]

	// Header: "Attachment 3/15"
	headerText := fmt.Sprintf("%sAttachment %d/%d[-]",
		ab.app.themeTags.Accent,
		ab.currentIndex+1, len(ab.filtered))
	ab.headerView.SetText(headerText)

	// Content: metadata fields
	keyColor := ab.app.themeTags.SecondaryText
	valColor := ab.app.themeTags.Foreground
	accentColor := ab.app.themeTags.Accent
	dividerColor := ab.app.themeTags.Border

	var lines []string

	if attach.Title != "" {
		lines = append(lines, fmt.Sprintf("[b]%s%s[-]", valColor, attach.Title))
		lines = append(lines, "")
	}

	if attach.Subtitle != "" {
		lines = append(lines, fmt.Sprintf("%sSubtitle:[-]    %s%s[-]", keyColor, valColor, attach.Subtitle))
	}

	if attach.URL != "" {
		lines = append(lines, fmt.Sprintf("%sURL:[-]         %s%s[-]", keyColor, accentColor, attach.URL))
	}

	if attach.SourceType != "" {
		lines = append(lines, fmt.Sprintf("%sSource:[-]      %s%s[-]", keyColor, valColor, attach.SourceType))
	}

	creatorDisplay := attach.Creator.DisplayName
	if creatorDisplay == "" {
		creatorDisplay = attach.Creator.Name
	}
	if creatorDisplay != "" {
		lines = append(lines, fmt.Sprintf("%sCreator:[-]     %s%s[-]", keyColor, valColor, creatorDisplay))
	}

	if !attach.CreatedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("%sCreated:[-]     %s%s[-]", keyColor, valColor, attach.CreatedAt.Format("Jan 2, 2006 3:04 PM")))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s────────────────────────────────────────[-]", dividerColor))

	ab.contentView.Clear()
	ab.contentView.SetText(strings.Join(lines, "\n"))
	ab.contentView.ScrollToBeginning()
}

// HandleKey handles keyboard input for the attachments browser.
func (ab *AttachmentsBrowser) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	// When search is active, only intercept specific keys
	if ab.searchActive {
		switch event.Key() {
		case tcell.KeyEscape:
			ab.toggleSearch()
			return nil
		case tcell.KeyEnter:
			ab.toggleSearch()
			return nil
		case tcell.KeyCtrlC:
			ab.app.app.Stop()
			return nil
		}
		// Let all other keys flow to the focused search input
		return event
	}

	switch event.Key() {
	case tcell.KeyEscape:
		ab.Hide()
		return nil
	case tcell.KeyCtrlN:
		row, col := ab.contentView.GetScrollOffset()
		ab.contentView.ScrollTo(row+1, col)
		return nil
	case tcell.KeyCtrlP:
		row, col := ab.contentView.GetScrollOffset()
		if row > 0 {
			ab.contentView.ScrollTo(row-1, col)
		}
		return nil
	case tcell.KeyCtrlD:
		_, _, _, h := ab.contentView.GetInnerRect()
		delta := max(1, h/2)
		row, col := ab.contentView.GetScrollOffset()
		ab.contentView.ScrollTo(row+delta, col)
		return nil
	case tcell.KeyCtrlU:
		_, _, _, h := ab.contentView.GetInnerRect()
		delta := max(1, h/2)
		row, col := ab.contentView.GetScrollOffset()
		targetRow := row - delta
		if targetRow < 0 {
			targetRow = 0
		}
		ab.contentView.ScrollTo(targetRow, col)
		return nil
	case tcell.KeyCtrlC:
		ab.app.app.Stop()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'n', 'j':
			if ab.currentIndex < len(ab.filtered)-1 {
				ab.currentIndex++
				ab.render()
			}
			return nil
		case 'p', 'k':
			if ab.currentIndex > 0 {
				ab.currentIndex--
				ab.render()
			}
			return nil
		case 'q':
			ab.Hide()
			return nil
		case '/':
			ab.toggleSearch()
			return nil
		case 'y':
			if len(ab.filtered) > 0 {
				attach := ab.filtered[ab.currentIndex]
				if attach.URL != "" {
					_ = copyToClipboard(attach.URL)
					ab.app.showToast("Copied attachment URL")
				}
			}
			return nil
		case 'o':
			if len(ab.filtered) > 0 {
				attach := ab.filtered[ab.currentIndex]
				if attach.URL != "" {
					_ = openURL(attach.URL)
				}
			}
			return nil
		}
	}

	// Consume all other keys
	return nil
}

// applyFilter filters attachments by title, subtitle, URL, and source type.
func (ab *AttachmentsBrowser) applyFilter() {
	if ab.searchQuery == "" {
		ab.filtered = ab.attachments
		ab.currentIndex = 0
		return
	}

	query := strings.ToLower(ab.searchQuery)
	filtered := make([]linearapi.Attachment, 0)
	for _, a := range ab.attachments {
		if strings.Contains(strings.ToLower(a.Title), query) ||
			strings.Contains(strings.ToLower(a.Subtitle), query) ||
			strings.Contains(strings.ToLower(a.URL), query) ||
			strings.Contains(strings.ToLower(a.SourceType), query) {
			filtered = append(filtered, a)
		}
	}
	ab.filtered = filtered
	ab.currentIndex = 0
}

// toggleSearch shows/hides the search input.
func (ab *AttachmentsBrowser) toggleSearch() {
	ab.searchActive = !ab.searchActive

	if ab.searchActive {
		ab.rebuildLayout(true)
		ab.app.app.SetFocus(ab.searchInput)
	} else {
		// Clear search and restore full list
		ab.searchQuery = ""
		ab.searchInput.SetText("")
		ab.filtered = ab.attachments
		ab.currentIndex = 0
		ab.rebuildLayout(false)
		ab.app.app.SetFocus(ab.contentView)
		ab.render()
	}
}

// rebuildLayout reconstructs the modal layout with or without the search row.
func (ab *AttachmentsBrowser) rebuildLayout(showSearch bool) {
	ab.modal.Clear()
	ab.modal.AddItem(ab.headerView, 1, 0, false)
	if showSearch {
		ab.modal.AddItem(ab.searchInput, 1, 0, false)
	}
	ab.modal.AddItem(ab.contentView, 0, 1, true)
	ab.modal.AddItem(ab.helpView, 1, 0, false)
}
