package ui

import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SelectableEntry struct {
	widget.Entry
	isFocused bool
}

func (m *SelectableEntry) FocusGained() {
	m.Entry.FocusGained()
	m.isFocused = true
	fyne.Do(func() { m.Refresh() })
	fyne.Do(func() { m.TypedShortcut(&fyne.ShortcutSelectAll{}) })
}

type compactLink struct{ *widget.Hyperlink }

func (c *compactLink) MinSize() fyne.Size { return fyne.NewSize(c.Hyperlink.MinSize().Width, 25) }

type myTheme struct{ fyne.Theme }

func (m myTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground || n == theme.ColorNameHyperlink {
		return color.White
	}
	return m.Theme.Color(n, v)
}


