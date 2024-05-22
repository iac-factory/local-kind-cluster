package pointers

func Boolean(value bool) *bool {
	return &value
}

func String(value string) *string {
	return &value
}

func Error(value error) *string {
	if value == nil {
		return nil
	}

	v := value.Error()

	return &v
}
