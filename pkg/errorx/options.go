package errorx

type Option func(*Error)

func WithCode(code Code) Option {
	return func(err *Error) {
		err.code = code
	}
}

func WithType(t string) Option {
	return func(err *Error) {
		err.errType = t
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
