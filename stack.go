package xo

type abort struct {
	err error
}

// Abort will abort with the supplied error.
func Abort(err error) {
	panic(abort{
		err: WS(err, 1),
	})
}

// AbortIf will only abort with the supplied error if present.
func AbortIf(err error) {
	if err != nil {
		panic(abort{
			err: WS(err, 1),
		})
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
	panic(WS(err, 1))
}

// PanicIf will only panic if the supplied error is present.
func PanicIf(err error) {
	if err != nil {
		panic(WS(err, 1))
	}
}

// Recover will recover any panic and call fn if an error has been recovered.
//
// Note: If the built-in panic function has been called with nil, a call to
// Recover will discard that panic and continue execution.
func Recover(fn func(error)) {
	val := recover()
	if val != nil {
		var err error
		switch val := val.(type) {
		case error:
			err = WF(val, "PANIC")
		case string:
			err = F("PANIC: %s", val)
		default:
			err = F("PANIC: %v", val)
		}

		// yield
		fn(Drop(err, 1))
	}
}

// Catch will call fn and recover any panic and return an error.
//
// Note: If the built-in panic function has been called with nil, a call to
// Recover will discard that panic and continue execution.
func Catch(fn func() error) (err error) {
	// recover panics
	defer Recover(func(e error) {
		err = Drop(e, 1)
	})

	// call fn
	err = WS(fn(), 1)

	return
}
