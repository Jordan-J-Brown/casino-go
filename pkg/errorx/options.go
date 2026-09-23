package errorx

type Option func(*Error)

func WithType(errType Type) Option {
	return func(err *Error) {
		err.errType = errType
	}
}

func WithChildren(children ...error) Option {
	return func(err *Error) {
		err.children = children
	}
}

func WithOutputStack() Option {
	return func(err *Error) {
		err.outputStack = true
	}
}
