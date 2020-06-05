package xo

type abort struct {
	err error
}

// Abort will abort with the supplied error.
func Abort(err error) {
	panic(abort{&Err{
		Err:    err,
		Caller: GetCaller(1),
	}})
}

// AbortIf will only abort with the supplied error if present.
func AbortIf(err error) {
	if err != nil {
		panic(abort{&Err{
			Err:    err,
			Caller: GetCaller(1),
		}})
	}
}

// Resume will recover an earlier call to Abort or AbortIf and call fn if an
// error has been recovered. It will not recover direct calls to the built-in
// panic function.
//
// Note: If the built-in panic function has been called with nil, a call to
// Resume will discard that panic and continue execution.
func Resume(fn func(error)) {
	val := recover()
	if cause, ok := val.(abort); ok {
		fn(cause.err)
		return
	} else if val != nil {
		panic(val)
	}
}

// Panic will panic with the provided error.
func Panic(err error) {
	panic(W(err))
}

// Recover will recover any panic and call fn if an errors has been recovered.
//
// Note: If the built-in panic function has been called with nil, a call to
// Recover will discard that panic and continue execution.
func Recover(fn func(error)) {
	val := recover()
	if val != nil {
		switch val := val.(type) {
		case error:
			fn(WF(val, "PANIC"))
		case string:
			fn(F("PANIC: %s", val))
		default:
			fn(F("PANIC: %v", val))
		}
	}
}
