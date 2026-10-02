package chans

func TryRecv[T any](ch <-chan T) (*T, bool) {
	select {
	case t := <-ch:
		return &t, true
	default:
		return nil, false
	}
}

func Collect[T any](ch <-chan T) []T {
	ts := make([]T, 0, len(ch))
	for t := range ch {
		ts = append(ts, t)
	}
	return ts
}
