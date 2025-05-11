package dataflow

type FilterFlow[T any] struct {
	source    DataFlow[T]
	predicate func(T) bool
	current   T
}

func Filter[T any](predicate func(T) bool) func(DataFlow[T]) DataFlow[T] {
	return func(source DataFlow[T]) DataFlow[T] {
		return &FilterFlow[T]{
			source:    source,
			predicate: predicate,
		}
	}
}

func (f *FilterFlow[T]) Next() bool {
	for f.source.Next() {
		value := f.source.Value()
		if f.predicate(value) {
			f.current = value
			return true
		}
	}
	return false
}

func (f *FilterFlow[T]) Value() T {
	return f.current
}

func (f *FilterFlow[T]) Reset() {
	f.source.Reset()
}
