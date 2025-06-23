package services

func ValidateInput(input interface{}) error {
	validator := GetValidator()
	err := validator.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
