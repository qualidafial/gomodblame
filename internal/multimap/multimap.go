package multimap

import "github.com/qualidafial/gomodblame/internal/set"

type Multimap[K, V comparable] map[K]set.Set[V]

func New[K, V comparable]() Multimap[K, V] {
	return make(Multimap[K, V])
}

func (m Multimap[K, V]) Size() int {
	size := 0
	for _, values := range m {
		size += len(values)
	}
	return size
}

func (m Multimap[K, V]) Add(key K, value V) {
	s, ok := m[key]
	if !ok {
		s = make(set.Set[V])
		m[key] = s
	}

	s.Add(value)
}

func (m Multimap[K, V]) Remove(key K, value V) {
	set, ok := m[key]
	if !ok {
		return
	}

	set.Remove(value)
	if len(set) == 0 {
		delete(m, key)
	}
}

func (m Multimap[K, V]) Contains(key K, value V) bool {
	_, ok := m[key][value]
	return ok
}

func (m Multimap[K, V]) ContainsKey(key K) bool {
	return len(m[key]) > 0
}

func (m Multimap[K, V]) All(f func(key K, value V) bool) bool {
	for key, values := range m {
		for value := range values {
			if !f(key, value) {
				return false
			}
		}
	}
	return true
}

func (m Multimap[K, V]) Inverse() Multimap[V, K] {
	inv := Multimap[V, K]{}
	for key, values := range m {
		for value := range values {
			inv.Add(value, key)
		}
	}
	return inv
}

func (m Multimap[K, V]) Clone() Multimap[K, V] {
	clone := Multimap[K, V]{}
	for key, values := range m {
		for value := range values {
			clone.Add(key, value)
		}
	}
	return clone
}
