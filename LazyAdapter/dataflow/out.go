package dataflow

import (
	"fmt"
	"io"
)

type OutFlow[T any] struct {
	source DataFlow[T]
	writer io.Writer
}

func Out[T any](writer io.Writer) func(DataFlow[T]) DataFlow[T] {
	return func(source DataFlow[T]) DataFlow[T] {
		flow := &OutFlow[T]{
			source: source,
			writer: writer,
		}

		for source.Next() {
			value := source.Value()
			fmt.Fprintln(writer, value)
		}

		source.Reset()

		return flow
	}
}

func (o *OutFlow[T]) Next() bool {
	return false
}

func (o *OutFlow[T]) Value() T {
	var zero T
	return zero
}

func (o *OutFlow[T]) Reset() {
	o.source.Reset()
}
