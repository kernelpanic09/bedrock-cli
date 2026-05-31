package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var isTTY = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

// Styles used across the CLI. All muted - this isn't a rainbow.
var (
	styleHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)
)

// Writer wraps an io.Writer with TTY-aware formatting helpers.
type Writer struct {
	w       io.Writer
	isTTY   bool
	noColor bool
}

// Stdout returns a Writer targeting os.Stdout.
func Stdout() *Writer {
	return &Writer{
		w:     os.Stdout,
		isTTY: isTTY,
	}
}

// New returns a Writer targeting the given writer. Color is disabled unless
// the caller explicitly sets it.
func New(w io.Writer) *Writer {
	return &Writer{w: w, noColor: true}
}

// SetNoColor disables color output regardless of TTY detection.
func (w *Writer) SetNoColor(v bool) {
	w.noColor = v
}

func (w *Writer) color() bool {
	return w.isTTY && !w.noColor
}

// Header prints a section header.
func (w *Writer) Header(text string) {
	if w.color() {
		fmt.Fprintln(w.w, styleHeader.Render(text))
	} else {
		fmt.Fprintln(w.w, text)
	}
}

// Dim prints muted/secondary text.
func (w *Writer) Dim(text string) {
	if w.color() {
		fmt.Fprintln(w.w, styleDim.Render(text))
	} else {
		fmt.Fprintln(w.w, text)
	}
}

// Success prints a success message.
func (w *Writer) Success(text string) {
	if w.color() {
		fmt.Fprintln(w.w, styleGreen.Render(text))
	} else {
		fmt.Fprintln(w.w, text)
	}
}

// Warn prints a warning.
func (w *Writer) Warn(text string) {
	if w.color() {
		fmt.Fprintln(w.w, styleYellow.Render(text))
	} else {
		fmt.Fprintf(w.w, "warning: %s\n", text)
	}
}

// Error prints an error message to stderr.
func (w *Writer) Error(text string) {
	if w.color() {
		fmt.Fprintln(os.Stderr, styleRed.Render(text))
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", text)
	}
}

// Bold prints bold text.
func (w *Writer) Bold(text string) {
	if w.color() {
		fmt.Fprintln(w.w, styleBold.Render(text))
	} else {
		fmt.Fprintln(w.w, text)
	}
}

// Println writes a plain line.
func (w *Writer) Println(text string) {
	fmt.Fprintln(w.w, text)
}

// Printf writes formatted output.
func (w *Writer) Printf(format string, args ...any) {
	fmt.Fprintf(w.w, format, args...)
}

// Writer returns the underlying io.Writer.
func (w *Writer) Writer() io.Writer {
	return w.w
}

// CostLine prints the per-invocation cost summary line.
// Format: [$0.0012, 234 tokens, 1.2s]
func (w *Writer) CostLine(costUSD float64, inputTokens, outputTokens int, durationMs int64, cached bool) {
	totalTokens := inputTokens + outputTokens
	seconds := float64(durationMs) / 1000.0

	var parts []string
	if cached {
		parts = append(parts, "cached")
	}
	parts = append(parts,
		fmt.Sprintf("$%.4f", costUSD),
		fmt.Sprintf("%d tokens", totalTokens),
		fmt.Sprintf("%.1fs", seconds),
	)

	line := "[" + strings.Join(parts, ", ") + "]"
	if w.color() {
		fmt.Fprintln(w.w, styleDim.Render(line))
	} else {
		fmt.Fprintln(w.w, line)
	}
}

// Table renders a simple ASCII table. headers and rows must have the same
// number of columns.
func (w *Writer) Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Compute column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build format string.
	var sb strings.Builder

	// Header row.
	if w.color() {
		fmt.Fprint(w.w, styleHeader.Render(formatRow(headers, widths))+"\n")
	} else {
		fmt.Fprintln(w.w, formatRow(headers, widths))
	}

	// Separator.
	sep := buildSep(widths)
	if w.color() {
		fmt.Fprintln(w.w, styleDim.Render(sep))
	} else {
		fmt.Fprintln(w.w, sep)
	}

	// Data rows.
	for _, row := range rows {
		sb.Reset()
		fmt.Fprintln(w.w, formatRow(row, widths))
	}
}

// BoxedResponse renders a model's response inside a labeled rounded box.
// Used in compare mode.
func (w *Writer) BoxedResponse(label, content string) {
	if !w.color() {
		fmt.Fprintf(w.w, "=== %s ===\n%s\n", label, content)
		return
	}
	boxStyle := styleBorder.Copy().BorderForeground(lipgloss.Color("12"))
	labeled := styleHeader.Render(label) + "\n" + content
	fmt.Fprintln(w.w, boxStyle.Render(labeled))
}

func formatRow(cells []string, widths []int) string {
	var sb strings.Builder
	for i, cell := range cells {
		if i < len(widths) {
			sb.WriteString(fmt.Sprintf("%-*s", widths[i], cell))
		} else {
			sb.WriteString(cell)
		}
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	return sb.String()
}

func buildSep(widths []int) string {
	var sb strings.Builder
	for i, w := range widths {
		sb.WriteString(strings.Repeat("-", w))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	return sb.String()
}
