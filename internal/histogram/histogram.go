package histogram

type Histogram[T comparable] map[T]int

func New[T comparable]() Histogram[T] {
	return make(Histogram[T])
}

func (h Histogram[T]) Add(value T) {
	h[value]++
}

func (h Histogram[T]) Remove(value T) {
	if h.Contains(value) {
		h[value]--
		if h[value] == 0 {
			delete(h, value)
		}
	}
}

func (h Histogram[T]) Contains(value T) bool {
	_, ok := h[value]
	return ok
}
