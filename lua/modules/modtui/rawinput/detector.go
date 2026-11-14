package rawinput

// escape is the byte that terminals send before a key pressed with Alt (e.g. Alt+T sends "ESC t")
const escape = 0x1b

// altTDetector detects the Alt+T key (with or without Shift) in a stream of bytes read from the terminal
type altTDetector struct {
	afterEscape bool
}

// Feed processes the next byte. Returns true if it completes an Alt+T key press
func (d *altTDetector) Feed(b byte) bool {
	if d.afterEscape && (b == 't' || b == 'T') {
		d.afterEscape = false
		return true
	}

	d.afterEscape = b == escape
	return false
}
