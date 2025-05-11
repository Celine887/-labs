package dataflow

type AggregateByKeyFlow[K comparable, V, T any] struct {
	source     DataFlow[T]
	aggregator func(T, V) V
	keyMapper  func(T) K
	initialVal func() V
	result     map[K]V
	keys       []K
	currentIdx int
}

func AggregateByKey[K comparable, V, T any](
	initialVal V,
	aggregator func(T, V) V,
	keyMapper func(T) K,
) func(DataFlow[T]) DataFlow[KV[K, V]] {
	return func(source DataFlow[T]) DataFlow[KV[K, V]] {
		return &AggregateByKeyFlow[K, V, T]{
			source:     source,
			aggregator: aggregator,
			keyMapper:  keyMapper,
			initialVal: func() V { return initialVal },
			currentIdx: -1,
		}
	}
}

func (a *AggregateByKeyFlow[K, V, T]) Next() bool {

	if a.result == nil {
		a.aggregate()
	}

	a.currentIdx++
	return a.currentIdx < len(a.keys)
}

func (a *AggregateByKeyFlow[K, V, T]) Value() KV[K, V] {
	if a.currentIdx < 0 || a.currentIdx >= len(a.keys) {
		var zero KV[K, V]
		return zero
	}

	key := a.keys[a.currentIdx]
	return KV[K, V]{
		Key:   key,
		Value: a.result[key],
	}
}

func (a *AggregateByKeyFlow[K, V, T]) Reset() {
	a.source.Reset()
	a.result = nil
	a.keys = nil
	a.currentIdx = -1
}

func (a *AggregateByKeyFlow[K, V, T]) aggregate() {
	a.result = make(map[K]V)
	keysMap := make(map[K]bool)

	for a.source.Next() {
		item := a.source.Value()
		key := a.keyMapper(item)

		if val, ok := a.result[key]; ok {
			a.result[key] = a.aggregator(item, val)
		} else {
			a.result[key] = a.aggregator(item, a.initialVal())
			keysMap[key] = true
		}
	}

	a.keys = make([]K, 0, len(keysMap))
	for key := range keysMap {
		a.keys = append(a.keys, key)
	}
}
