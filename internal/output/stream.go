package output

import (
	"fmt"
	"io"
	"os"
)

// StreamWriter writes tokens to a destination as they arrive.
// It handles the transition from streaming to the final cost line.
type StreamWriter struct {
	w   io.Writer
	tty bool
}

// NewStreamWriter returns a StreamWriter targeting the given writer.
func NewStreamWriter(w io.Writer, tty bool) *StreamWriter {
	return &StreamWriter{w: w, tty: tty}
}

// StdoutStreamWriter returns a StreamWriter that writes to stdout.
func StdoutStreamWriter() *StreamWriter {
	return &StreamWriter{w: os.Stdout, tty: isTTY}
}

// WriteToken writes a single token chunk to the output.
func (s *StreamWriter) WriteToken(token string) {
	fmt.Fprint(s.w, token)
}

// Flush ensures a trailing newline after streaming is complete.
func (s *StreamWriter) Flush() {
	fmt.Fprintln(s.w)
}
