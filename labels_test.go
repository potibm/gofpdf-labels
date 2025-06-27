package gofpdflabels

import (
	"math"
	"os"
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func TestGetLabelReturnsCorrectPositionAndSize(t *testing.T) {
	doc, err := NewPdfLabelDocument("90x54", 0, 0)

	if err != nil {
		t.Fatalf("Failed to create PdfLabelDoc: %v", err)
	}

	label := doc.placeLabel()

	if label.Position.X != doc.MarginLeft+doc.LabelPadding {
		t.Errorf("X position wrong: got %f, want %f", label.Position.X, doc.MarginLeft+doc.LabelPadding)
	}

	if label.Position.Y != doc.MarginTop+doc.LabelPadding {
		t.Errorf("Y position wrong: got %f, want %f", label.Position.Y, doc.MarginTop+doc.LabelPadding)
	}

	expectedWidth := doc.LabelSize.Width - 2*doc.LabelPadding
	if label.Size.Width != expectedWidth {
		t.Errorf("Width wrong: got %f, want %f", label.Size.Width, expectedWidth)
	}
}

func TestAdvanceLabelIncrementsCorrectly(t *testing.T) {
	doc, _ := NewPdfLabelDocument("90x54", 0, 0)
	doc.SetFont("Arial", "", 12)

	col := doc.ColPosition

	doc.advanceLabel()

	if doc.ColPosition != col+1 && col+1 < doc.Cols {
		t.Errorf("ColPosition not incremented correctly: got %d", doc.ColPosition)
	}

	// simulate reaching the end of a row
	doc.ColPosition = doc.Cols - 1
	doc.advanceLabel()

	if doc.ColPosition != 0 {
		t.Errorf("ColPosition should reset after Cols exceeded")
	}
}

func TestAddingPages(t *testing.T) {
	doc, _ := NewPdfLabelDocument("90x54", 0, 0)
	doc.SetFont("Arial", "", 12)

	totalPerPage := doc.Rows * doc.Cols

	// should start with page 1
	if doc.PageNo() != 0 {
		t.Errorf("Expected page number to start at 0, got %d", doc.PageNo())
	}

	// fill a full page
	for i := 0; i < totalPerPage; i++ {
		doc.AddLabel("Test")

		if doc.PageNo() != 1 {
			t.Errorf("Page number should be 1 while filling first page, got %d", doc.PageNo())
		}
	}

	// Add another label → should now be on page 2
	doc.AddLabel("Overflow")

	if doc.PageNo() != 2 {
		t.Errorf("Expected page number after overflow to be 2, got %d", doc.PageNo())
	}
}

func TestDifferentStartPosition(t *testing.T) {
	doc, _ := NewPdfLabelDocument("90x54", 4, 1)
	doc.SetFont("Arial", "", 12)

	doc.AddLabel("Test ")

	// last label on first page should be at position (4, 1)
	if doc.PageNo() != 1 {
		t.Errorf("Expected page number to start at 1, got %d", doc.PageNo())
	}

	doc.AddLabel("Test 2")

	// Now the next page comes
	if doc.PageNo() != 2 {
		t.Errorf("Expected page number to be 2 after overflow, got %d", doc.PageNo())
	}
}

func TestColumnsAndRows(t *testing.T) {
	doc, _ := NewPdfLabelDocument("90x54", 0, 0)
	doc.SetFont("Arial", "", 12)
	totalPerPage := doc.Rows * doc.Cols

	for i := 0; i < totalPerPage; i++ {
		if (i%2) == 0 && doc.ColPosition != 0 {
			t.Errorf("ColPosition should be reset to 0 at the start of a new row, got %d", doc.ColPosition)
		}

		if (i%2) == 1 && doc.ColPosition != 1 {
			t.Errorf("ColPosition should be 1 for odd labels in a 2-column layout, got %d", doc.ColPosition)
		}

		doc.AddLabel("text")
	}
}

func TestPdfOutputFile(t *testing.T) {
	doc, _ := NewPdfLabelDocument("90x54", 0, 0)
	doc.SetFont("Arial", "", 12)
	doc.AddCustomLabel(func(pdf *fpdf.Fpdf, label Label) {
		pdf.CellFormat(label.Size.Width, label.Size.Height, "Hello World", "", 0, "LT", false, 0, "")
	})
	doc.AddLabel("Hello world - Second label")

	err := doc.OutputFileAndClose("test.pdf")
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	info, err := os.Stat("test.pdf")
	if err != nil {
		t.Fatalf("PDF file not created")
	}

	if info.Size() == 0 {
		t.Errorf("PDF file is empty")
	}

	_ = os.Remove("test.pdf")
}

func TestConvertUnit(t *testing.T) {
	const epsilon = 1e-9

	// 1 inch == 25.4 mm
	mm := convertUnit(1.0, "in", "mm")
	if math.Abs(mm-25.4) > epsilon {
		t.Errorf("Expected 1 in to be 25.4 mm, got %f", mm)
	}

	// 25.4 mm == 1 inch
	in := convertUnit(25.4, "mm", "in")
	if math.Abs(in-1.0) > epsilon {
		t.Errorf("Expected 25.4 mm to be 1 in, got %f", in)
	}

	// Same unit should return same value
	val := convertUnit(10.0, "mm", "mm")
	if math.Abs(val-10.0) > epsilon {
		t.Errorf("Expected 10 mm to stay 10 mm, got %f", val)
	}
}

func TestUnknownLabelFormat(t *testing.T) {
	_, err := NewPdfLabelDocument("UnknownFormat", 0, 0)
	if err == nil {
		t.Error("Expected error for unknown label format, got nil")
	}

	expectedErr := "unknown label format: UnknownFormat"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message '%s', got '%s'", expectedErr, err.Error())
	}
}
