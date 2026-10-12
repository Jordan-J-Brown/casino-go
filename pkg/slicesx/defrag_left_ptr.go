package slicesx

// Defragments non-nil elements to the left of the slice, shifting nil elements to the right.
// This does not remove elements and maintains the original capacity.
//
// [indexLast] will be -1 if all elements are nil.
func DefragLeftPtr[T any, S ~[]*T](slice S) (indexLast int) {
	write := 0
	for read := range slice {
		if slice[read] != nil {
			slice[write] = slice[read]
			write++
		}
	}

	for i := write; i < len(slice); i++ {
		slice[i] = nil
	}

	return write - 1
}
