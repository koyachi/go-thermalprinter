package thermalprinter

import (
	"fmt"
	"github.com/tarm/serial"
	"strings"
	"time"
	"unicode/utf8"
)

// Character commands
const (
	InverseMask      = 1 << 1
	UpdownMask       = 1 << 2
	BoldMask         = 1 << 3
	DoubleHeightMask = 1 << 4
	DoubleWidthMask  = 1 << 5
	StrikeMask       = 1 << 6
)

const (
	UPC_A = iota
	UPC_E
	EAN13
	EAN8
	CODE39
	I25
	CODEBAR
	CODE93
	CODE128
	CODE11
	MSI
)

type Printer struct {
	port            *serial.Port
	timeout         int
	resumeTime      time.Time
	byteTime        float64
	dotPrintTime    float64
	dotFeedTime     float64
	prevByte        byte
	column          int
	maxColumn       int
	charHeight      int
	lineSpacing     int
	barcodeHeight   int
	printMode       byte
	defaultHeatTime int
}

func charToByte(c string) byte {
	r, _ := utf8.DecodeLastRuneInString(c)
	return byte(r)
}

func newlineByte() byte {
	return charToByte("\n")
}

func NewPrinter(name string, baud int, timeout int) (*Printer, error) {
	c := &serial.Config{Name: name, Baud: baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}

	p := &Printer{
		port:            s,
		timeout:         timeout,
		dotPrintTime:    0.033,
		dotFeedTime:     0.0025,
		prevByte:        newlineByte(),
		column:          0,
		maxColumn:       32,
		charHeight:      24,
		lineSpacing:     8,
		barcodeHeight:   50,
		printMode:       0,
		defaultHeatTime: 60,
	}

	// Calculate time to issue one byte to the printer.
	// 11 bits (not 8) to accomodate idle, start and stop bits.
	// Idle time might be unnecessary, but erring on side of
	// caution here.
	p.byteTime = 11.0 / float64(baud)

	// The printer can't start receiving data immediately upon
	// power up -- it needs a moment to cold boot and initialize.
	// Allow at least 1/2 sec of uptime before printer can
	// receive data.
	p.timeoutSet(0.5)

	p.Wake()
	p.Reset()

	//これがあると薄い？ => 関係無かった
	/*
		// Description of print settings from page 23 of the manual:
		// ESC 7 n1 n2 n3 Setting Control Parameter Command
		// Decimal: 27 55 n1 n2 n3
		// Set "max heating dots", "heating time", "heating interval"
		// n1 = 0-255 Max heat dots, Unit (8dots), Default: 7 (64 dots)
		// n2 = 3-255 Heating time, Unit (10us), Default: 80 (800us)
		// n3 = 0-255 Heating interval, Unit (10us), Default: 2 (20us)
		// The more max heating dots, the more peak current will cost
		// when printing, the faster printing speed. The max heating
		// dots is 8*(n1+1). The more heating time, the more density,
		// but the slower printing speed. If heating time is too short,
		// blank page may occur. The more heating interval, the more
		// clear, but the slower printing speed.

		// TODO: heatTime from method arguments.
			heatTime := p.defaultHeatTime
			heatDots := 20
			heatInterval := 250
			p.writeBytes([]byte{
				27,                 // Esc
				55,                 // 7 (print settings)
				byte(heatDots),     // Heat dots (20 = balance darkness w/no jams)
				byte(heatTime),     // Lib deefault = 45
				byte(heatInterval), // Heat interval(500us = slower but darker)
			})
	*/

	//これがあると薄い？ => 関係ない？
	/*
		// Description of print density from page 23 of the manual:
		// DC2 # n Set printing density
		// Decimal: 18 35 n
		// D4..D0 of n is used to set the printing density.
		// Density is 50% + 5% * n(D4-D0) printing density.
		// D7..D5 of n is used to set the printing break time.
		// Break time is n(D7-D5)*250us.
		// (Unsure of the default value for either -- not documented)
		printDensity := 14  // 120% (can go higher, but text gets fuzzy)
		printBreakTime := 4 // 500us
		p.writeBytes([]byte{
			18, // DC2
			35, // Print density
			byte(printBreakTime<<5 | printDensity),
		})
	*/

	p.dotPrintTime = 0.03
	//p.dotPrintTime = 0.05
	//p.dotPrintTime = 0.1
	p.dotFeedTime = 0.0021

	return p, nil
}

func (p *Printer) timeoutSet(second float64) {
	d := time.Duration(float64(time.Second) * second)
	p.resumeTime = time.Now().Add(d)
}

func (p *Printer) timeoutWait() {
	for time.Now().Sub(p.resumeTime) < 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *Printer) writeBytes(data []byte) error {
	p.timeoutWait()
	p.timeoutSet(float64(len(data)) * p.byteTime)
	_, err := p.port.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Printer) write(data []byte) error {
	for _, c := range data {
		if c == 0x13 {
			continue
		}

		p.timeoutWait()
		_, err := p.port.Write([]byte{c})
		if err != nil {
			return err
		}
		d := p.byteTime
		if c == newlineByte() || p.column == p.maxColumn {
			// Newline or wrap
			if p.prevByte == newlineByte() {
				// Feed line (blank)
				d += float64(p.charHeight+p.lineSpacing) * p.dotFeedTime
			} else {
				// Text line
				d += (float64(p.charHeight) * p.dotPrintTime) + (float64(p.lineSpacing) * p.dotFeedTime)
				p.column = 0
				// Treat wrap as Newlineon next pass
				c = newlineByte()
			}
		} else {
			p.column += 1
		}
		p.timeoutSet(d)
		p.prevByte = c
	}
	/*
		// TODO:n
		_, err := p.port.Write(data)
		if err != nil {
			return err
		}
	*/
	return nil
}

func (p *Printer) Reset() {
	p.prevByte = newlineByte() // Treat as  if prior line is blank
	p.column = 0
	p.maxColumn = 32
	p.charHeight = 24
	p.lineSpacing = 8
	p.barcodeHeight = 50
	p.writeBytes([]byte{27, 64})
}

func (p *Printer) SetDefault() {
	p.Online()
	p.Justify("L")
	p.InverseOff()
	p.DoubleHeightOff()
	p.SetLineHeight(32)
	p.BoldOff()
	p.UnderlineOff()
	p.SetBarcodeHeight(50)
	p.SetSize("s")
}

func (p *Printer) PrintBarcode(text string, barcodeType int) {
	p.writeBytes([]byte{
		29, 72, 2, // Print label below barcodeType
		29, 119, 3, // Barcode width
		29, 107, byte(barcodeType), // Barcode type
	})
	// Print string
	p.timeoutWait()
	p.timeoutSet(float64(p.barcodeHeight+40) * p.dotPrintTime)
	p.port.Write([]byte(text))
	p.prevByte = newlineByte()
	p.Feed(2)
}

func (p *Printer) SetBarcodeHeight(val ...int) {
	_val := 50
	if val != nil && len(val) == 1 {
		_val = val[0]
	}
	if _val < 1 {
		_val = 1
	}
	p.barcodeHeight = _val
	p.writeBytes([]byte{29, 104, byte(_val)})
}

func (p *Printer) writePrintMode() {
	p.writeBytes([]byte{27, 33, p.printMode})
}

func (p *Printer) setPrintMode(mask byte) {
	p.printMode |= mask
	p.writePrintMode()
	if p.printMode&DoubleHeightMask != 0 {
		p.charHeight = 48
	} else {
		p.charHeight = 24
	}
	if p.printMode&DoubleWidthMask != 0 {
		p.maxColumn = 16
	} else {
		p.maxColumn = 32
	}
}

func (p *Printer) unsetPrintMode(mask byte) {
	p.printMode &= ^mask
	p.writePrintMode()
	if p.printMode&DoubleHeightMask != 0 {
		p.charHeight = 48
	} else {
		p.charHeight = 24
	}
	if p.printMode&DoubleWidthMask != 0 {
		p.maxColumn = 16
	} else {
		p.maxColumn = 32
	}
}

func (p *Printer) Normal() {
	p.printMode = 0
	p.writePrintMode()
}

func (p *Printer) InverseOn() {
	p.setPrintMode(InverseMask)
}

func (p *Printer) InverseOff() {
	p.unsetPrintMode(InverseMask)
}

func (p *Printer) UpsideDownOn() {
	p.setPrintMode(UpdownMask)
}

func (p *Printer) UpsideDownOff() {
	p.unsetPrintMode(UpdownMask)
}

func (p *Printer) DoubleHeightOn() {
	p.setPrintMode(DoubleHeightMask)
}

func (p *Printer) DoubleHeightOff() {
	p.unsetPrintMode(DoubleHeightMask)
}

func (p *Printer) DoubleWidthOn() {
	p.setPrintMode(DoubleWidthMask)
}

func (p *Printer) DoubleWidthOff() {
	p.unsetPrintMode(DoubleWidthMask)
}

func (p *Printer) StrikeOn() {
	p.setPrintMode(StrikeMask)
}

func (p *Printer) StrikeOff() {
	p.unsetPrintMode(StrikeMask)
}

func (p *Printer) BoldOn() {
	p.setPrintMode(BoldMask)
}

func (p *Printer) BoldOff() {
	p.unsetPrintMode(BoldMask)
}

func (p *Printer) Justify(value string) {
	var pos byte
	switch strings.ToUpper(value) {
	case "C":
		pos = 1
	case "R":
		pos = 2
	default:
		pos = 0
	}
	p.writeBytes([]byte{0x1B, 0x61, pos})
}

// Feeds by the specified number of lines
func (p *Printer) Feed(x ...int) {
	_x := 1
	if x != nil && len(x) == 1 {
		_x = x[0]
	}
	for _x > 0 {
		p.Print("\n")
		_x -= 1
	}
}

// Feed by the specified number of indivisual pixel rows
func (p *Printer) FeedRows(rows int) {
	p.writeBytes([]byte{27, 74, byte(rows)})
	p.timeoutSet(float64(rows) * p.dotFeedTime)
}

func (p *Printer) SetSize(value string) {
	var size byte
	switch strings.ToUpper(value) {
	case "L":
		// Large: double width and height
		size = 0x11
		p.charHeight = 48
		p.maxColumn = 16
	case "M":
		// Medium: double height
		size = 0x01
		p.charHeight = 48
		p.maxColumn = 32
	default:
		// Small: standard width and height
		size = 0x00
		p.charHeight = 24
		p.maxColumn = 32
	}
	p.writeBytes([]byte{29, 33, size, 10})
	p.prevByte = newlineByte() // Setting the size adds a linefeed
}

func (p *Printer) SetLineHeight(val ...int) {
	_val := 24
	if val != nil && len(val) == 1 {
		_val = val[0]
	}
	p.lineSpacing = _val - 24

	// The printer doesn't take into account the current text
	// height when setting line height, making this more skin
	// to inter-line spacing. Default line spacing is 32
	// (char height of 24, line spacing of 8).
	p.writeBytes([]byte{27, 51, byte(_val)})
}

// Underlines of different weights can be produced:
// 0 - no underline
// 1 - normal underline
// 2 - thick underline
func (p *Printer) UnderlineOn(weight ...int) {
	_weight := 1
	if weight != nil && len(weight) == 1 {
		_weight = weight[0]
	}
	p.writeBytes([]byte{27, 45, byte(_weight)})
}

func (p *Printer) UnderlineOff() {
	p.UnderlineOn(0)
}

func (p *Printer) PrintBitmap(w int, h int, bitmap []byte, lineAtATime bool) error {
	rowBytes := int(float64(w+7) / 8) // Round up to next byte boundary
	rowBytesClipped := 0
	if rowBytes >= 48 {
		rowBytesClipped = 48 // 384 pixels max width
	} else {
		rowBytesClipped = rowBytes
	}

	// if lineAtATime is true, print bitmaps
	// scanline-at-a-time (rather than in chunks).
	// This tends to make for much cleaner printin
	// (no feed gaps) on large images...but has the
	// oppsite effect on small images that would fit
	// in a single 'chunk', so use carefully!
	maxChunkHeight := 0
	if lineAtATime {
		maxChunkHeight = 1
	} else {
		// TODO:maxChunkHeightを使うロジックおかしそうなのでなおす => timeoutSetの時間調整？
		maxChunkHeight = 255
		//maxChunkHeight = 127
		//maxChunkHeight = h
	}

	i := 0
	for rowStart := 0; rowStart < h; rowStart += maxChunkHeight {
		chunkHeight := h - rowStart
		if chunkHeight > maxChunkHeight {
			chunkHeight = maxChunkHeight
		}
		fmt.Printf("rowBytes = %d, rowBytesClipped = %d\n", rowBytes, rowBytesClipped)
		fmt.Printf("h = %d, chunkHeight = %d, rowStart = %d\n", h, chunkHeight, rowStart)

		// Timeout wait happens here
		p.writeBytes([]byte{18, 42, byte(chunkHeight), byte(rowBytesClipped)})

		for y := 0; y < chunkHeight; y++ {
			fmt.Printf("  y = %d, i = %d\n", y, i)
			for x := 0; x < rowBytesClipped; x++ {
				_, err := p.port.Write([]byte{bitmap[i]})
				if err != nil {
					return err
				}
				i++
			}
			i += rowBytes - rowBytesClipped
		}
		//ためしにコメントアウト => 解決した。けどどうすべきか。
		//p.timeoutSet(float64(chunkHeight) * p.dotPrintTime)
	}
	p.prevByte = newlineByte()

	return nil
}

func (p *Printer) SetTimes(pt float64, ft float64) {
	p.dotPrintTime = pt
	p.dotFeedTime = ft
}

// Take the printer offline. Print commands sent after this
// will be ignored until 'online' is called.
func (p *Printer) Offline() {
	p.writeBytes([]byte{27, 61, 0})
}

// Take the printer online, Subsequent print commands will be obeyed.
func (p *Printer) Online() {
	p.writeBytes([]byte{27, 61, 1})
}

func (p *Printer) Sleep() {
	seconds := 1
	p.writeBytes([]byte{27, 56, byte(seconds)})
}

func (p *Printer) Wake() {
	p.timeoutSet(0)
	p.writeBytes([]byte{255})
	for i := 0; i < 10; i++ {
		p.writeBytes([]byte{27})
		p.timeoutSet(0.1)
	}
}

func (p *Printer) Flush() {
	p.writeBytes([]byte{12})
}

func (p *Printer) Print(s string) {
	p.write([]byte(s))
}

func (p *Printer) Println(s string) {
	p.Print(s + "\n")
}
