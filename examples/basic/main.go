package main

import (
	"log"
	"strconv"

	labels "github.com/potibm/gofpdf-labels"
)

func main() {
	const fontsize = 12

	doc, err := labels.NewPdfLabelDocument("90x54", 0, 0)
	doc.SetFont("Arial", "", fontsize)

	if err != nil {
		log.Fatal(err)
	}

	// Add 10 simple labels
	for i := 0; i < 10; i++ {
		doc.AddLabel("Label #" + strconv.Itoa(i+1))
	}

	// Write to file
	err = doc.OutputFileAndClose("labels.pdf")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Generated labels.pdf")
}
