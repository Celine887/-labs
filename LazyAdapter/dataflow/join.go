package dataflow

type JoinResult[K comparable, L, R any] struct {
	Key   K
	Left  L
	Right *R
}

type JoinFlow[K comparable, L, R any] struct {
	leftSource  DataFlow[L]
	rightSource DataFlow[R]
	leftKey     func(L) K
	rightKey    func(R) K
	rightMap    map[K]R
	leftItems   []L
	currentIdx  int
}

func Join[K comparable, L, R any](
	rightSource DataFlow[R],
	leftKey func(L) K,
	rightKey func(R) K,
) func(DataFlow[L]) DataFlow[JoinResult[K, L, R]] {
	return func(leftSource DataFlow[L]) DataFlow[JoinResult[K, L, R]] {
		return &JoinFlow[K, L, R]{
			leftSource:  leftSource,
			rightSource: rightSource,
			leftKey:     leftKey,
			rightKey:    rightKey,
			currentIdx:  -1,
		}
	}
}

func (j *JoinFlow[K, L, R]) Next() bool {

	if j.rightMap == nil {
		j.prepare()
	}

	j.currentIdx++
	return j.currentIdx < len(j.leftItems)
}

func (j *JoinFlow[K, L, R]) Value() JoinResult[K, L, R] {
	if j.currentIdx < 0 || j.currentIdx >= len(j.leftItems) {
		var zero JoinResult[K, L, R]
		return zero
	}

	leftItem := j.leftItems[j.currentIdx]
	key := j.leftKey(leftItem)

	if rightItem, ok := j.rightMap[key]; ok {
		return JoinResult[K, L, R]{
			Key:   key,
			Left:  leftItem,
			Right: &rightItem,
		}
	}

	return JoinResult[K, L, R]{
		Key:   key,
		Left:  leftItem,
		Right: nil,
	}
}

func (j *JoinFlow[K, L, R]) Reset() {
	j.leftSource.Reset()
	j.rightSource.Reset()
	j.rightMap = nil
	j.leftItems = nil
	j.currentIdx = -1
}

func (j *JoinFlow[K, L, R]) prepare() {

	j.rightMap = make(map[K]R)
	for j.rightSource.Next() {
		rightItem := j.rightSource.Value()
		key := j.rightKey(rightItem)
		j.rightMap[key] = rightItem
	}

	j.leftItems = make([]L, 0)
	for j.leftSource.Next() {
		j.leftItems = append(j.leftItems, j.leftSource.Value())
	}
}
