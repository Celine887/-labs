package dataflow

type SliceFlow[T any] struct {
	data       []T
	currentIdx int
}

func AsDataFlow[T any](data []T) *Pipeline[T] {
	return New(&SliceFlow[T]{
		data:       data,
		currentIdx: -1,
	})
}

func (s *SliceFlow[T]) Next() bool {
	s.currentIdx++
	return s.currentIdx < len(s.data)
}

func (s *SliceFlow[T]) Value() T {
	if s.currentIdx < 0 || s.currentIdx >= len(s.data) {
		var zero T
		return zero
	}
	return s.data[s.currentIdx]
}

func (s *SliceFlow[T]) Reset() {
	s.currentIdx = -1
}

type AsVectorFlow[T any] struct {
	source   DataFlow[T]
	result   []T
	consumed bool
}

func AsVector[T any]() func(DataFlow[T]) DataFlow[[]T] {
	return func(source DataFlow[T]) DataFlow[[]T] {
		return &AsVectorFlow[T]{
			source: source,
		}
	}
}

func (a *AsVectorFlow[T]) Next() bool {
	if a.consumed {
		return false
	}

	a.result = make([]T, 0)
	for a.source.Next() {
		a.result = append(a.result, a.source.Value())
	}

	a.consumed = true
	return true
}

func (a *AsVectorFlow[T]) Value() []T {
	return a.result
}

func (a *AsVectorFlow[T]) Reset() {
	a.source.Reset()
	a.result = nil
	a.consumed = false
}
