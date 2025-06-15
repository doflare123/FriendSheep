package utils

func ValidateInput(input interface{}) error {
	err := validate.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
