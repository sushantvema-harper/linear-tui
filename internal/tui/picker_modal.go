package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PickerItem represents an item in a picker.
type PickerItem struct {
	ID    string
	Label string
}

// PickerModal manages a picker overlay for selecting from a list of items.
type PickerModal struct {
	app          *App
	modal        *tview.Flex
	modalContent *tview.Flex
	input        *tview.InputField
	list         *tview.List
	items        []PickerItem
	filtered     []PickerItem
	query        string
	onSelect     func(item PickerItem)
}

// NewPickerModal creates a new picker modal.
func NewPickerModal(app *App) *PickerModal {
	pm := &PickerModal{
		app: app,
	}

	// Create search input
	pm.input = tview.NewInputField()
	pm.input.
		SetLabel("> ").
		SetLabelColor(app.theme.Accent).
		SetFieldWidth(0).
		SetPlaceholder("Type to filter...").
		SetPlaceholderTextColor(app.theme.SecondaryText).
		SetFieldBackgroundColor(app.theme.InputBg).
		SetFieldTextColor(app.theme.Foreground).
		SetBackgroundColor(app.theme.HeaderBg)

	// Create list
	pm.list = tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(app.theme.Foreground).
		SetSelectedBackgroundColor(app.theme.Accent).
		SetSelectedTextColor(app.theme.SelectionText).
		SetHighlightFullLine(true)
	pm.list.SetBackgroundColor(app.theme.HeaderBg)

	// Create help text
	helpText := tview.NewTextView()
	helpText.SetText("↑↓: navigate | Enter: select | Esc: cancel")
	helpText.SetTextColor(app.theme.SecondaryText)
	helpText.SetBackgroundColor(app.theme.HeaderBg)

	// Build modal content
	pm.modalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pm.input, 1, 0, true).
		AddItem(pm.list, 0, 1, false).
		AddItem(helpText, 1, 0, false)
	pm.modalContent.Box = tview.NewBox().SetBackgroundColor(app.theme.HeaderBg)
	pm.modalContent.SetBackgroundColor(app.theme.HeaderBg).
		SetBorder(true).
		SetBorderColor(app.theme.Accent).
		SetTitleColor(app.theme.Foreground)
	padding := app.density.ModalPadding
	pm.modalContent.SetBorderPadding(padding.Top, padding.Bottom, padding.Left, padding.Right)

	// Center the modal on screen
	pm.modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(pm.modalContent, 15, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
	pm.modal.SetBackgroundColor(app.theme.Background)

	return pm
}

// Show displays the picker modal with the given title and items.
func (pm *PickerModal) Show(title string, items []PickerItem, onSelect func(item PickerItem)) {
	pm.items = items
	pm.filtered = items
	pm.query = ""
	pm.onSelect = onSelect

	pm.modalContent.SetTitle(" " + title + " ")
	pm.input.SetText("")
	pm.input.SetChangedFunc(func(text string) {
		pm.query = text
		pm.filterItems()
		pm.updateList()
	})

	pm.updateList()

	pm.app.pages.AddPage("picker", pm.modal, true, true)
	pm.app.pages.SendToFront("picker")
	pm.app.app.SetFocus(pm.input)
}

// Hide hides the picker modal.
func (pm *PickerModal) Hide() {
	pm.app.pickerActive = false
	pm.input.SetChangedFunc(nil)
	pm.app.pages.RemovePage("picker")

	// Check if create issue modal is visible and restore focus to it
	if pm.app.pages.HasPage("create_issue") {
		pm.app.pages.SendToFront("create_issue")
		if pm.app.createIssueModal != nil {
			pm.app.app.SetFocus(pm.app.createIssueModal.form)
		}
		return
	}

	// Check if edit title modal is visible and restore focus to it
	if pm.app.pages.HasPage("edit_title") {
		pm.app.pages.SendToFront("edit_title")
		if pm.app.editTitleModal != nil {
			pm.app.app.SetFocus(pm.app.editTitleModal.form)
		}
		return
	}

	pm.app.updateFocus()
}

// HandleKey handles keyboard input for the picker.
func (pm *PickerModal) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		pm.Hide()
		return nil
	case tcell.KeyEnter:
		idx := pm.list.GetCurrentItem()
		if idx >= 0 && idx < len(pm.filtered) {
			item := pm.filtered[idx]
			pm.Hide()
			if pm.onSelect != nil {
				pm.onSelect(item)
			}
		}
		return nil
	case tcell.KeyUp, tcell.KeyCtrlP:
		idx := pm.list.GetCurrentItem()
		if idx > 0 {
			pm.list.SetCurrentItem(idx - 1)
		}
		return nil
	case tcell.KeyDown, tcell.KeyCtrlN:
		idx := pm.list.GetCurrentItem()
		if idx < pm.list.GetItemCount()-1 {
			pm.list.SetCurrentItem(idx + 1)
		}
		return nil
	}
	return event
}

// filterItems filters the items based on the current query.
func (pm *PickerModal) filterItems() {
	if pm.query == "" {
		pm.filtered = pm.items
		return
	}

	query := strings.ToLower(pm.query)
	filtered := make([]PickerItem, 0)
	for _, item := range pm.items {
		if strings.Contains(strings.ToLower(item.Label), query) {
			filtered = append(filtered, item)
		}
	}
	pm.filtered = filtered
}

// updateList refreshes the list display with filtered items.
func (pm *PickerModal) updateList() {
	pm.list.Clear()
	for _, item := range pm.filtered {
		pm.list.AddItem(item.Label, "", 0, nil)
	}
	if len(pm.filtered) > 0 {
		pm.list.SetCurrentItem(0)
	}
}

// GetModal returns the modal flex for adding to pages.
func (pm *PickerModal) GetModal() *tview.Flex {
	return pm.modal
}
