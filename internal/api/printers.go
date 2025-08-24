package api

import (
	"bytes"
	"fmt"
	"pa11y-go-wrapper/internal/analysis"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"

	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// GenerateHTML generates an HTML document from a list of analyses.
func GenerateHTML(analyses []*analysis.Analysis) (string, error) {
	var builder bytes.Buffer

	builder.WriteString("<html><head><title>Completed Analyses</title></head><body>")
	builder.WriteString("<h1>Completed Analyses</h1>")
	builder.WriteString("<table border='1'><tr><th>ID</th><th>URL</th><th>Status</th><th>Created At</th><th>Updated At</th></tr>")

	for _, a := range analyses {
		builder.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			a.ID, a.URL, a.Status, a.CreatedAt.Format("2006-01-02 15:04:05"), a.UpdatedAt.Format("2006-01-02 15:04:05")))
	}

	builder.WriteString("</table></body></html>")

	return builder.String(), nil
}

// GeneratePDF generates a PDF document from a list of analyses.
func GeneratePDF(analyses []*analysis.Analysis) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	m.AddRows(text.NewRow(10, "Completed Analyses", props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	m.AddRows(getTransactions(analyses)...)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document.GetBytes(), nil
}

func getTransactions(analyses []*analysis.Analysis) []core.Row {
	rows := []core.Row{
		row.New(5).Add(
			text.NewCol(2, "ID", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
			text.NewCol(4, "URL", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
			text.NewCol(2, "Status", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
			text.NewCol(2, "Created At", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
			text.NewCol(2, "Updated At", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		),
	}

	var contentsRow []core.Row
	for i, a := range analyses {
		r := row.New(4).Add(
			text.NewCol(2, a.ID, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(4, a.URL, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, string(a.Status), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, a.CreatedAt.Format("2006-01-02 15:04:05"), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, a.UpdatedAt.Format("2006-01-02 15:04:05"), props.Text{Size: 8, Align: align.Center}),
		)
		if i%2 == 0 {
			gray := getGrayColor()
			r.WithStyle(&props.Cell{BackgroundColor: gray})
		}

		contentsRow = append(contentsRow, r)
	}

	rows = append(rows, contentsRow...)

	return rows
}
func getGrayColor() *props.Color {
	return &props.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}
