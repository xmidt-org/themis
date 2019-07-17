package xerror

import "time"

// TryOptionalDuration returns a closure of the form expected by Do.  Useful
// when stringing together a bunch of parsing calls.  The returned closure returns
// a nil error if source is blank.
func TryOptionalDuration(source string, target *time.Duration) func() error {
	return func() (err error) {
		if len(source) > 0 {
			*target, err = time.ParseDuration(source)
		}

		return
	}
}
