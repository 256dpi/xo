package xo

type cause struct {
	err error
}

// Abort will abort with the supplied error.
func Abort(err error) {
	panic(cause{&Err{
		Err:    err,
		Caller: GetCaller(1),
	}})
}

// AbortIf will only abort with the supplied error if present.
func AbortIf(err error) {
	if err != nil {
		panic(cause{&Err{
			Err:    err,
			Caller: GetCaller(1),
		}})
	}
}

// Resume will try to recover an earlier call to Abort or AbortIf and call fn
// if an error has been recovered. It will not recover direct calls to the
// built-in panic function.
//
// Note: If the built-in panic function has been called with nil, a call to
// Resume will discard that panic and continue execution.
func Resume(fn func(error)) {
	val := recover()
	if cause, ok := val.(cause); ok {
		fn(cause.err)
		return
	} else if val != nil {
		panic(val)
	}
}
