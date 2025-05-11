package dataflow

import (
	"fmt"
)

type DataFlow[T any] interface {
	Next() bool
	Value() T
	Reset()
}

func Pipe[T, U any](source DataFlow[T], adapter func(DataFlow[T]) DataFlow[U]) DataFlow[U] {
	return adapter(source)
}

type Pipeline[T any] struct {
	dataflow DataFlow[T]
}

func New[T any](dataflow DataFlow[T]) *Pipeline[T] {
	return &Pipeline[T]{dataflow: dataflow}
}

func (p *Pipeline[T]) GetFlow() DataFlow[T] {
	return p.dataflow
}

func (p *Pipeline[T]) Apply(adapter interface{}) interface{} {
	switch typedAdapter := adapter.(type) {
	case func(DataFlow[T]) DataFlow[string]:
		return &Pipeline[string]{dataflow: typedAdapter(p.dataflow)}
	case func(DataFlow[T]) DataFlow[[]byte]:
		return &Pipeline[[]byte]{dataflow: typedAdapter(p.dataflow)}
	case func(DataFlow[T]) DataFlow[int]:
		return &Pipeline[int]{dataflow: typedAdapter(p.dataflow)}
	case func(DataFlow[T]) DataFlow[bool]:
		return &Pipeline[bool]{dataflow: typedAdapter(p.dataflow)}
	case func(DataFlow[T]) DataFlow[float64]:
		return &Pipeline[float64]{dataflow: typedAdapter(p.dataflow)}
	default:
		panic(fmt.Sprintf("Unsupported adapter type: %T", adapter))
	}
}

func (p *Pipeline[T]) Run() []T {
	var results []T
	for p.dataflow.Next() {
		results = append(results, p.dataflow.Value())
	}
	return results
}

type KV[K, V any] struct {
	Key   K
	Value V
}
