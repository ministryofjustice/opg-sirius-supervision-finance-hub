package shared

type NillableString struct {
	Value string
	Valid bool
}

func TransformNillableString(stringPointer *string) NillableString {
	var transformedNillableString NillableString
	if stringPointer != nil {
		transformedNillableString = NillableString{
			*stringPointer,
			true,
		}
	}
	return transformedNillableString
}
