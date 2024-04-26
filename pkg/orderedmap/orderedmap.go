package orderedmap

type OrderedMap[T any] struct {
	container map[string]T
	keys      []string
}

func New[T any]() *OrderedMap[T] {
	return &OrderedMap[T]{
		container: make(map[string]T),
		keys:      make([]string, 0),
	}
}

func (om *OrderedMap[T]) Set(key string, value T) {
	if _, ok := om.container[key]; !ok {
		om.keys = append(om.keys, key)
	}

	om.container[key] = value
}

func (om *OrderedMap[T]) Get(key string) (T, bool) {
	value, ok := om.container[key]
	return value, ok
}

func (om *OrderedMap[T]) Iterate() <-chan struct {
	Key   string
	Value T
} {
	ch := make(chan struct {
		Key   string
		Value T
	})

	go func() {
		for _, key := range om.keys {
			ch <- struct {
				Key   string
				Value T
			}{
				Key:   key,
				Value: om.container[key],
			}
		}

		close(ch)
	}()

	return ch
}
