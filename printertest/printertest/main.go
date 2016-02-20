package main

import (
	"github.com/koyachi/go-thermalprinter"
	"log"
)

func main() {
	printer, err := thermalprinter.NewPrinter("/dev/ttyAMA0", 19200, 5)
	if err != nil {
		log.Fatal(err)
	}

	// Test inverse on & off
	printer.InverseOn()
	printer.Println("Inverse ON")
	printer.InverseOff()

	// Test character double-height on & off
	printer.DoubleHeightOn()
	printer.Println("Double Height ON")
	printer.DoubleHeightOff()

	// Set justification (right, center, left) -- accepts "L", "C", "R"
	printer.Justify("R")
	printer.Println("Right justified")
	printer.Justify("C")
	printer.Println("Center justified")
	printer.Justify("L")
	printer.Println("Left justified")

	// Test more styles
	printer.BoldOn()
	printer.Println("Bold text")
	printer.BoldOff()

	printer.UnderlineOn()
	printer.Println("Underlined text")
	printer.UnderlineOff()

	printer.SetSize("L") // Set type size, accepts "S", "M", "L"
	printer.Println("Large")
	printer.SetSize("M")
	printer.Println("Medium")
	printer.SetSize("S")
	printer.Println("Small")

	printer.Justify("C")
	printer.Println("normal\nline\nspacing")
	printer.SetLineHeight(50)
	printer.Println("Taller\nline\nspacing")
	printer.SetLineHeight() // Reset to default
	printer.Justify("L")

	// Barcode examples
	printer.Feed(1)
	// CODE39 is the most common alphanumeric barcode
	printer.PrintBarcode("ADAFRUT", thermalprinter.CODE39)
	printer.SetBarcodeHeight(100)
	// Print UPC line on product barcodes
	printer.PrintBarcode("123456789123", thermalprinter.UPC_A)

	// TODO: PrintBitmap

	printer.Sleep()      // Tell printer to sleep
	printer.Wake()       // Call Wake() before printing again, even if reset
	printer.SetDefault() // Restore printer to defaults
}
