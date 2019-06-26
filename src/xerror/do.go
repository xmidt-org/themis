package xerror

// Do simply runs each of a sequence of functions in succession, stopping
// at any function that returns an error.  If no functions return errors,
// Do returns nil.
func Do(fs ...func() error) error {
	for _, f := range fs {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}
