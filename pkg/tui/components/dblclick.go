package components

import "time"

const dblClickThreshold = 500 * time.Millisecond

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

func (d *DblClickDetector) Click(idx int) bool {
	now := d.clock()
	dbl := idx == d.lastIdx && now.Sub(d.lastTime) < dblClickThreshold
	if dbl {
		d.lastTime = time.Time{}
	} else {
		d.lastIdx = idx
		d.lastTime = now
	}
	return dbl
}
