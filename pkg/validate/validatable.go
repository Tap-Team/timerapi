package validate

type Validatable interface {
	Validate() error
}

type ValidatableStruct interface {
	ValidatableVariables() []Validatable
}

func ValidateFields(v ValidatableStruct) error {
	for _, v := range v.ValidatableVariables() {
		err := v.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// validate list of validatable return index of failed or -1 and error
func ValidateMany[T Validatable](list []T) (int, error) {
	for i, v := range list {
		err := v.Validate()
		if err != nil {
			return i, err
		}
	}
	return -1, nil
}

func ValidateManyStructs[T ValidatableStruct](list []T) (int, error) {
	for i, v := range list {
		err := ValidateFields(v)
		if err != nil {
			return i, err
		}
	}
	return -1, nil
}
