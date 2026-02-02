// Copyright (c) 2024 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package collection

// Set contains a collection of unique items.
// See https://en.wikipedia.org/wiki/Set_(mathematics) for set theory.
type Set[T comparable] struct {
	items map[T]struct{}
}

// Create a new set that can store items of type T.
func NewSet[T comparable]() Set[T] {
	return Set[T]{
		items: make(map[T]struct{}),
	}
}

// Create a new set that can store items of type T with the capacity pre-allocated.
func NewSetWithCapacity[T comparable](capacity int) Set[T] {
	return Set[T]{
		items: make(map[T]struct{}, capacity),
	}
}

// Create a new set that contains only the unique items from a number of slices.
func NewSetFrom[T comparable](args ...[]T) Set[T] {
	capacity := 0
	for _, arg := range args {
		capacity += len(arg)
	}

	s := NewSetWithCapacity[T](capacity)
	for _, arg := range args {
		for _, i := range arg {
			s.Insert(i)
		}
	}
	return s
}

// Return the number of items stored in the set.
func (s Set[T]) Len() int {
	return len(s.items)
}

// The items stored in the set.
func (s Set[T]) Items() []T {
	result := make([]T, 0, len(s.items))
	for k := range s.items {
		result = append(result, k)
	}
	return result
}

// Insert a new item into the set.
// Returns true if the item could be inserted and false if the item is already in the set.
func (s Set[T]) Insert(item T) bool {
	if s.Contains(item) {
		return false
	}

	s.items[item] = struct{}{}
	return true
}

// Insert a slice of items into the set.
func (s Set[T]) InsertSlice(items []T) {
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// Remove the item from the set.
// Returns true if the item was in the set before removing.
func (s Set[T]) Remove(item T) bool {
	found := s.Contains(item)
	delete(s.items, item)
	return found
}

// Remove a slice of items from the set.
func (s Set[T]) RemoveSlice(items []T) {
	for _, item := range items {
		delete(s.items, item)
	}
}

// Returns true if the item is in the set.
func (s Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

// ContainsSlice returns true if all items in the slice is present in the set.
func (s Set[T]) ContainsSlice(items []T) bool {
	for _, item := range items {
		_, ok := s.items[item]
		if !ok {
			return false
		}
	}
	return true
}

// Return a new set that is the union of this set and another.
func (a Set[T]) Union(b Set[T]) Set[T] {
	c := Set[T]{
		items: MapUnion(a.items, b.items),
	}
	return c
}

// Return a new set that contains only the items that are present in both sets.
func (a Set[T]) Intersection(b Set[T]) Set[T] {
	c := Set[T]{
		items: MapIntersection(a.items, b.items),
	}
	return c
}

// Return a new set that contains only the items that are present in this set but not in b.
func (a Set[T]) Difference(b Set[T]) Set[T] {
	c := Set[T]{
		items: MapDifference(a.items, b.items),
	}
	return c
}

// Return a new set that contains only the items that are present in one or the other set but not the items that appear in both sets.
func (a Set[T]) SymmetricDifference(b Set[T]) Set[T] {
	c := Set[T]{
		items: MapSymmetricDifference(a.items, b.items),
	}
	return c
}
