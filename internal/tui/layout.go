package tui

type Layout struct {
	visibleWidth  int
	visibleHeight int
	reserved      int
}

func NewLayout() *Layout {
	return &Layout{
		visibleWidth:  0,
		visibleHeight: 0,
		reserved:      2,
	}
}

func (l *Layout) SetWidth(width int) {
	l.visibleWidth = width
}

func (l *Layout) SetHeight(height int) {
	l.visibleHeight = height
}

func (l *Layout) LeftPanesWidth() int {
	return l.visibleWidth/3 - l.reserved
}

func (l *Layout) RightPanesWidth() int {
	return 2*l.visibleWidth/3 - l.reserved
}

func (l *Layout) LeftMonitorsHeight() int {
	return 3 * l.visibleHeight / 4
}

func (l *Layout) RightPreviewHeight() int {
	return 7 * l.visibleHeight / 8
}
