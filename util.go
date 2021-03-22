package sqlx

func UnwrapAll(err error) error {
	for {
		if cause := UnwrapOnce(err); cause != nil {
			err = cause
			continue
		}
		break
	}
	return err
}

func UnwrapOnce(err error) (cause error) {
	switch e := err.(type) {
	case interface{ Cause() error }:
		return e.Cause()
	case interface{ Unwrap() error }:
		return e.Unwrap()
	}
	return nil
}
