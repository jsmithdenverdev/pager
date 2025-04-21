package app

func Authorize() func() error {
	return func() error {
		return nil
	}
}
