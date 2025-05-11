package dataflow

import (
	"fmt"
	"io"
)

type WriteFlow[T any] struct {
	source    DataFlow[T]
	writer    io.Writer
	separator string
	processed bool
}

func Write[T any](writer io.Writer, separator string) func(DataFlow[T]) DataFlow[T] {
	return func(source DataFlow[T]) DataFlow[T] {
		return &WriteFlow[T]{
			source:    source,
			writer:    writer,
			separator: separator,
			processed: false,
		}
	}
}

func (w *WriteFlow[T]) Next() bool {
	if w.processed {
		return false
	}

	isFirst := true
	for w.source.Next() {
		if !isFirst {
			fmt.Fprint(w.writer, w.separator)
		}
		fmt.Fprint(w.writer, w.source.Value())
		isFirst = false
	}

	if !isFirst {
		fmt.Fprint(w.writer, w.separator)
	}

	w.processed = true
	return false
}

func (w *WriteFlow[T]) Value() T {
	var zero T
	return zero
}

func (w *WriteFlow[T]) Reset() {
	w.source.Reset()
	w.processed = false
}
