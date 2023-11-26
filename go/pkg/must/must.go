package must

import "fmt"

func NoError(err error) {
	if err != nil {
		panic(fmt.Sprintf("must: non-nil error: %v", err))
	}
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("must: non-nil error: %v", err))
	}
	return value
}
