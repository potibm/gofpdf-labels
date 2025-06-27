package gofpdflabels

import (
	"errors"
	"math"

	"codeberg.org/go-pdf/fpdf"
)

type Point struct {
	X float64
	Y float64
}

type Size struct {
	Width  float64
	Height float64
}

type Label struct {
	Position Point
	Size     Size
}

type PdfLabelDoc struct {
	*fpdf.Fpdf
	MarginLeft          float64
	MarginTop           float64
	XSpace              float64
	YSpace              float64
	Rows                int
	Cols                int
	LabelSize           Size
	LabelPadding        float64
	SheetUnit           string
	RowPosition         int
	ColPosition         int
	CutLines            bool
	Format              string
	PendingPageCreation bool // Indicates if a new page needs to be created
}

type LabelFormat struct {
	PaperSize  string
	Unit       string
	MarginLeft float64
	MarginTop  float64
	Cols       int
	Rows       int
	SpaceX     float64
	SpaceY     float64
	Width      float64
	Height     float64
	CutLines   bool
}

type LabelCallback func(pdf *fpdf.Fpdf, label Label)

func NewPdfLabelDocument(formatName string, row, col int) (*PdfLabelDoc, error) {
	const defaultLabelPadding = 3

	format, ok := labels[formatName]

	if !ok {
		return nil, errors.New("unknown label format: " + formatName)
	}

	unit := format.Unit
	pdf := fpdf.New("P", "mm", format.PaperSize, "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	label := &PdfLabelDoc{
		Fpdf:       pdf,
		MarginLeft: convertUnit(format.MarginLeft, unit, "mm"),
		MarginTop:  convertUnit(format.MarginTop, unit, "mm"),
		XSpace:     convertUnit(format.SpaceX, unit, "mm"),
		YSpace:     convertUnit(format.SpaceY, unit, "mm"),
		Cols:       format.Cols,
		Rows:       format.Rows,
		LabelSize: Size{
			Width:  convertUnit(format.Width, unit, "mm"),
			Height: convertUnit(format.Height, unit, "mm"),
		},
		LabelPadding:        convertUnit(defaultLabelPadding, "mm", "mm"),
		CutLines:            format.CutLines,
		RowPosition:         int(math.Abs(float64(row % format.Rows))),
		ColPosition:         int(math.Abs(float64(col % format.Cols))),
		SheetUnit:           unit,
		Format:              formatName,
		PendingPageCreation: true, // Start with a new page
	}

	//label.addLabelPage()

	return label, nil
}

func (p *PdfLabelDoc) AddCustomLabel(callback LabelCallback) {
	if p.PendingPageCreation {
		p.addLabelPage()
		p.PendingPageCreation = false // Reset the flag after creating a page
	}

	label := p.placeLabel()

	callback(p.Fpdf, label)

	p.advanceLabel()
}

func (p *PdfLabelDoc) AddLabel(text string) {
	p.AddCustomLabel(func(pdf *fpdf.Fpdf, label Label) {
		pdf.CellFormat(label.Size.Width, label.Size.Height, text, "", 0, "LT", false, 0, "")
	})
}

func (p *PdfLabelDoc) addLabelPage() {
	p.AddPage()

	p.drawCutLines()
}

func (p *PdfLabelDoc) drawCutLines() {
	const Half = 2.0

	if !p.CutLines {
		return
	}

	pageWidth, pageHeight := p.GetPageSize()

	// vertical lines
	for col := 0; col <= p.Cols; col++ {
		x := p.MarginLeft + float64(col)*(p.LabelSize.Width+p.XSpace) - p.XSpace/Half
		p.Line(x, 0, x, pageHeight)
	}

	// horizontal lines
	for row := 0; row <= p.Rows; row++ {
		y := p.MarginTop + float64(row)*(p.LabelSize.Height+p.YSpace) - p.YSpace/Half
		p.Line(0, y, pageWidth, y)
	}
}

func convertUnit(value float64, fromUnit, toUnit string) float64 {
	const mmPerInch = 25.4

	// Normalize units to lowercase
	from := normalizeUnit(fromUnit)
	to := normalizeUnit(toUnit)

	if from == to {
		return value
	}

	// Conversion factors relative to 1 inch
	var unitsPerInch = map[string]float64{
		"in": 1.0,
		"mm": mmPerInch,
	}

	fromFactor, fromOk := unitsPerInch[from]
	toFactor, toOk := unitsPerInch[to]

	if !fromOk || !toOk {
		panic("unsupported unit conversion from " + from + " to " + to)
	}

	// Convert value to inches, then to destination unit
	return value * (toFactor / fromFactor)
}

func normalizeUnit(unit string) string {
	switch unit {
	case "in", "inch", "inches":
		return "in"
	case "mm", "millimeter", "millimeters":
		return "mm"
	default:
		return unit
	}
}

func (p *PdfLabelDoc) currentLabel() Label {
	return Label{
		Position: Point{
			X: p.MarginLeft + float64(p.ColPosition)*(p.LabelSize.Width+p.XSpace) + p.LabelPadding,
			Y: p.MarginTop + float64(p.RowPosition)*(p.LabelSize.Height+p.YSpace) + p.LabelPadding,
		},
		Size: Size{
			Width:  p.LabelSize.Width - 2*p.LabelPadding,
			Height: p.LabelSize.Height - 2*p.LabelPadding,
		},
	}
}

func (p *PdfLabelDoc) placeLabel() Label {
	label := p.currentLabel()
	p.SetXY(label.Position.X, label.Position.Y)

	return label
}

func (p *PdfLabelDoc) advanceLabel() {
	p.ColPosition++

	if p.ColPosition >= p.Cols {
		p.ColPosition = 0
		p.RowPosition++

		if p.RowPosition >= p.Rows {
			p.RowPosition = 0
			p.PendingPageCreation = true // Mark that a new page is needed
		}
	}
}
