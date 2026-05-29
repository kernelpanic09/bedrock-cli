package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestCostLine(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	w.CostLine(0.0012, 100, 134, 1200, false)

	got := buf.String()
	if !strings.Contains(got, "$0.0012") {
		t.Errorf("CostLine() output %q missing cost", got)
	}
	if !strings.Contains(got, "234 tokens") {
		t.Errorf("CostLine() output %q missing token count", got)
	}
	if !strings.Contains(got, "1.2s") {
		t.Errorf("CostLine() output %q missing duration", got)
	}
}

func TestCostLineCached(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	w.CostLine(0.0, 100, 50, 50, true)

	got := buf.String()
	if !strings.Contains(got, "cached") {
		t.Errorf("CostLine(cached) output %q missing 'cached' label", got)
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"alpha", "1"},
		{"beta", "2"},
	}

	w.Table(headers, rows)

	got := buf.String()
	for _, h := range headers {
		if !strings.Contains(got, h) {
			t.Errorf("Table() output missing header %q", h)
		}
	}
	for _, row := range rows {
		for _, cell := range row {
			if !strings.Contains(got, cell) {
				t.Errorf("Table() output missing cell %q", cell)
			}
		}
	}
}

func TestTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)
	// Should not panic on empty input.
	w.Table([]string{}, [][]string{})
}

func TestBoxedResponseNoColor(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	w.BoxedResponse("haiku", "some response")

	got := buf.String()
	if !strings.Contains(got, "haiku") {
		t.Errorf("BoxedResponse() output %q missing label", got)
	}
	if !strings.Contains(got, "some response") {
		t.Errorf("BoxedResponse() output %q missing content", got)
	}
}

func TestPrintln(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)
	w.Println("hello world")

	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("Println() output %q missing text", buf.String())
	}
}
