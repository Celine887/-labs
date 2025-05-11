package dataflow

type TransformFlow[T, U any] struct {
	source      DataFlow[T]
	transformer func(T) U
	current     U
}

func Transform[T, U any](transformer func(T) U) func(DataFlow[T]) DataFlow[U] {
	return func(source DataFlow[T]) DataFlow[U] {
		return &TransformFlow[T, U]{
			source:      source,
			transformer: transformer,
		}
	}
}

func (t *TransformFlow[T, U]) Next() bool {
	if t.source.Next() {
		t.current = t.transformer(t.source.Value())
		return true
	}
	return false
}

func (t *TransformFlow[T, U]) Value() U {
	return t.current
}

func (t *TransformFlow[T, U]) Reset() {
	t.source.Reset()
}
