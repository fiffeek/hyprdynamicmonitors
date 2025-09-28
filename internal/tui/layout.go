package tui

import "github.com/sirupsen/logrus"

type Layout struct {
	visibleWidth  int
	visibleHeight int
	reserved      int
	reservedTop   int
}

func NewLayout() *Layout {
	return &Layout{
		visibleWidth:  0,
		visibleHeight: 0,
		reserved:      2,
		reservedTop:   0,
	}
}

func (l *Layout) SetReservedTop(r int) {
	l.reservedTop = r
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
	logrus.Debugf("Reserved width %d", l.reserved)
	return 2*l.visibleWidth/3 - l.reserved
}

func (l *Layout) AvailableHeight() int {
	return l.visibleHeight - l.reservedTop - l.reserved
}

func (l *Layout) AvailableWidth() int {
	return l.visibleWidth - l.reserved
}

func (l *Layout) LeftMonitorsHeight() int {
	return 3 * l.AvailableHeight() / 4
}

func (l *Layout) LeftSubpaneHeight() int {
	return l.AvailableHeight() / 4
}

func (l *Layout) RightPreviewHeight() int {
	return 6 * l.AvailableHeight() / 7
}

func (l *Layout) RightHyprHeight() int {
	return l.AvailableHeight() / 7
}
