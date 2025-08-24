package api

import (
	"bytes"
	"fmt"
	"html"
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

	builder.WriteString("<html><head><title>Accessibility Analyses</title><meta charset='utf-8'></head><body>")
	builder.WriteString("<h1>Accessibility Analyses</h1>")

	if len(analyses) == 0 {
		builder.WriteString("<p>No analyses to display.</p>")
		builder.WriteString("</body></html>")
		return builder.String(), nil
	}

	for _, a := range analyses {
		builder.WriteString("<section style='margin-bottom:24px'>")
		builder.WriteString("<h2>" + html.EscapeString(a.URL) + "</h2>")
		builder.WriteString("<table border='1' cellpadding='4' cellspacing='0'>")
		builder.WriteString("<tr><th align='left'>ID</th><td>" + html.EscapeString(a.ID) + "</td></tr>")
		builder.WriteString("<tr><th align='left'>URL</th><td>" + html.EscapeString(a.URL) + "</td></tr>")
		builder.WriteString("<tr><th align='left'>Status</th><td>" + html.EscapeString(string(a.Status)) + "</td></tr>")
		if a.Runner != "" {
			builder.WriteString("<tr><th align='left'>Runner</th><td>" + html.EscapeString(a.Runner) + "</td></tr>")
		}
		if a.ErrorMessage != "" {
			builder.WriteString("<tr><th align='left'>Error</th><td>" + html.EscapeString(a.ErrorMessage) + "</td></tr>")
		}
		builder.WriteString("<tr><th align='left'>Created At</th><td>" + a.CreatedAt.Format("2006-01-02 15:04:05") + "</td></tr>")
		builder.WriteString("<tr><th align='left'>Updated At</th><td>" + a.UpdatedAt.Format("2006-01-02 15:04:05") + "</td></tr>")
		builder.WriteString("</table>")

		// Issues
		builder.WriteString("<h3>Issues (" + fmt.Sprintf("%d", len(a.Result)) + ")</h3>")
		if len(a.Result) == 0 {
			builder.WriteString("<p>No issues found.</p>")
		} else {
			builder.WriteString("<table border='1' cellpadding='4' cellspacing='0'>")
			builder.WriteString("<tr>" +
				"<th>#</th>" +
				"<th>Code</th>" +
				"<th>Message</th>" +
				"<th>Type</th>" +
				"<th>TypeCode</th>" +
				"<th>Selector</th>" +
				"<th>Context</th>" +
				"</tr>")
			for idx, issue := range a.Result {
				builder.WriteString("<tr>")
				builder.WriteString("<td>" + fmt.Sprintf("%d", idx+1) + "</td>")
				builder.WriteString("<td>" + html.EscapeString(issue.Code) + "</td>")
				builder.WriteString("<td>" + html.EscapeString(issue.Message) + "</td>")
				builder.WriteString("<td>" + html.EscapeString(issue.Type) + "</td>")
				builder.WriteString("<td>" + fmt.Sprintf("%d", issue.TypeCode) + "</td>")
				builder.WriteString("<td>" + html.EscapeString(issue.Selector) + "</td>")
				builder.WriteString("<td>" + html.EscapeString(issue.Context) + "</td>")
				builder.WriteString("</tr>")
			}
			builder.WriteString("</table>")
		}
		builder.WriteString("</section>")
	}

	builder.WriteString("</body></html>")

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

	m.AddRows(text.NewRow(10, "Accessibility Analyses", props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	// Add each analysis as a section
	for idx, a := range analyses {
		m.AddRows(getAnalysisSectionRows(a)...)
		// Add a spacer between analyses (except after the last one)
		if idx < len(analyses)-1 {
			m.AddRows(text.NewRow(5, " ", props.Text{}))
		}
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document.GetBytes(), nil
}

func getAnalysisSectionRows(a *analysis.Analysis) []core.Row {
	rows := []core.Row{}

	// Section title with URL
	rows = append(rows, text.NewRow(8, fmt.Sprintf("Analysis: %s", a.URL), props.Text{Style: fontstyle.Bold, Align: align.Left}))

	// Header key/value rows
	rows = append(rows, row.New(5).Add(
		text.NewCol(2, "ID:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
		text.NewCol(10, a.ID, props.Text{Size: 9, Align: align.Left}),
	))
	rows = append(rows, row.New(5).Add(
		text.NewCol(2, "URL:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
		text.NewCol(10, a.URL, props.Text{Size: 9, Align: align.Left}),
	))
	rows = append(rows, row.New(5).Add(
		text.NewCol(2, "Status:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
		text.NewCol(10, string(a.Status), props.Text{Size: 9, Align: align.Left}),
	))
	if a.Runner != "" {
		rows = append(rows, row.New(5).Add(
			text.NewCol(2, "Runner:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
			text.NewCol(10, a.Runner, props.Text{Size: 9, Align: align.Left}),
		))
	}
	if a.ErrorMessage != "" {
		rows = append(rows, row.New(5).Add(
			text.NewCol(2, "Error:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
			text.NewCol(10, a.ErrorMessage, props.Text{Size: 9, Align: align.Left}),
		))
	}
	rows = append(rows, row.New(5).Add(
		text.NewCol(2, "Created:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
		text.NewCol(10, a.CreatedAt.Format("2006-01-02 15:04:05"), props.Text{Size: 9, Align: align.Left}),
	))
	rows = append(rows, row.New(5).Add(
		text.NewCol(2, "Updated:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Left}),
		text.NewCol(10, a.UpdatedAt.Format("2006-01-02 15:04:05"), props.Text{Size: 9, Align: align.Left}),
	))

	// Spacer
	rows = append(rows, text.NewRow(4, " ", props.Text{}))

	// Issues header
	rows = append(rows, text.NewRow(7, fmt.Sprintf("Issues (%d)", len(a.Result)), props.Text{Style: fontstyle.Bold, Align: align.Left}))

	if len(a.Result) == 0 {
		rows = append(rows, text.NewRow(5, "No issues found.", props.Text{Align: align.Left}))
		return rows
	}

	// Issues table header
	headers := row.New(5).Add(
		text.NewCol(1, "#", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(2, "Code", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(3, "Message", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(2, "Type", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(1, "TypeCode", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(3, "Selector", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold}),
	)
	rows = append(rows, headers)

	// Issue rows
	for i, issue := range a.Result {
		ir := row.New(5).Add(
			text.NewCol(1, fmt.Sprintf("%d", i+1), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, issue.Code, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(3, issue.Message, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, issue.Type, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(1, fmt.Sprintf("%d", issue.TypeCode), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(3, issue.Selector, props.Text{Size: 8, Align: align.Left}),
		)
		if i%2 == 0 {
			ir.WithStyle(&props.Cell{BackgroundColor: getGrayColor()})
		}
		rows = append(rows, ir)
	}

	return rows
}

func getGrayColor() *props.Color {
	return &props.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}
