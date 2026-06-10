package components

import "time"

const dblClickThreshold = 500 * time.Millisecond

// DblClickDetector tracks clicks to detect double-clicks on the same item
type DblClickDetector struct {
	lastIdx  int
	lastTime time.Time
	now      func() time.Time
}

func (d *DblClickDetector) clock() time.Time {
	if d.now != nil {
		return d.now()
	}
	return time.Now()
}

// Click registers a click on idx and returns true if it is a double-click
// (same index clicked twice within 500ms)
func (d *DblClickDetector) Click(idx int) bool {
	now := d.clock()
	dbl := idx == d.lastIdx && now.Sub(d.lastTime) < dblClickThreshold
	if dbl {
		d.lastTime = time.Time{} // reset to prevent triple-click
	} else {
		d.lastIdx = idx
		d.lastTime = now
	}
	return dbl
}
