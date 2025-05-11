package dataflow

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func Some[T any](value T) Optional[T] {
	return Optional[T]{
		Value:    value,
		HasValue: true,
	}
}

func None[T any]() Optional[T] {
	return Optional[T]{
		HasValue: false,
	}
}

type DropNulloptFlow[T any] struct {
	source  DataFlow[Optional[T]]
	current T
}

func DropNullopt[T any]() func(DataFlow[Optional[T]]) DataFlow[T] {
	return func(source DataFlow[Optional[T]]) DataFlow[T] {
		return &DropNulloptFlow[T]{
			source: source,
		}
	}
}

func (d *DropNulloptFlow[T]) Next() bool {
	for d.source.Next() {
		optional := d.source.Value()
		if optional.HasValue {
			d.current = optional.Value
			return true
		}
	}
	return false
}

func (d *DropNulloptFlow[T]) Value() T {
	return d.current
}

func (d *DropNulloptFlow[T]) Reset() {
	d.source.Reset()
}

type Result[T, E any] struct {
	Value    T
	Error    E
	HasError bool
}

func Success[T, E any](value T) Result[T, E] {
	return Result[T, E]{
		Value:    value,
		HasError: false,
	}
}

func Failure[T, E any](err E) Result[T, E] {
	return Result[T, E]{
		Error:    err,
		HasError: true,
	}
}

type SplitExpectedResult[T, E, ST, SE any] struct {
	Success DataFlow[ST]
	Failure DataFlow[SE]
}

func SplitExpected[T, E, ST, SE any](
	successTransform func(T) ST,
	errorTransform func(E) SE,
) func(DataFlow[Result[T, E]]) SplitExpectedResult[T, E, ST, SE] {
	return func(source DataFlow[Result[T, E]]) SplitExpectedResult[T, E, ST, SE] {

		source.Reset()

		successFlow := &FilterTransformFlow[Result[T, E], ST]{
			source: source,
			filter: func(result Result[T, E]) bool {
				return !result.HasError
			},
			transform: func(result Result[T, E]) ST {
				return successTransform(result.Value)
			},
		}

		source.Reset()
		failureFlow := &FilterTransformFlow[Result[T, E], SE]{
			source: source,
			filter: func(result Result[T, E]) bool {
				return result.HasError
			},
			transform: func(result Result[T, E]) SE {
				return errorTransform(result.Error)
			},
		}

		return SplitExpectedResult[T, E, ST, SE]{
			Success: successFlow,
			Failure: failureFlow,
		}
	}
}

type FilterTransformFlow[T, U any] struct {
	source    DataFlow[T]
	filter    func(T) bool
	transform func(T) U
	current   U
}

func (f *FilterTransformFlow[T, U]) Next() bool {
	for f.source.Next() {
		value := f.source.Value()
		if f.filter(value) {
			f.current = f.transform(value)
			return true
		}
	}
	return false
}

func (f *FilterTransformFlow[T, U]) Value() U {
	return f.current
}

func (f *FilterTransformFlow[T, U]) Reset() {
	f.source.Reset()
}
