package shared

type Nillable[T any] struct {
	Value T
	Valid bool
}

func TransformNillableDate(stringPointer *string) Nillable[Date] {
	var transformedNillable Nillable[Date]
	if stringPointer != nil {
		transformedNillable = Nillable[Date]{
			NewDate(*stringPointer),
			true,
		}
	}
	return transformedNillable
}

func TransformNillableString(stringPointer *string) Nillable[string] {
	var transformedNillable Nillable[string]
	if stringPointer != nil {
		transformedNillable = Nillable[string]{
			*stringPointer,
			true,
		}
	}
	return transformedNillable
}

func TransformNillableInt(stringPointer *string) Nillable[int32] {
	var transformedNillableInt Nillable[int32]
	if stringPointer != nil {
		transformedNillableInt = Nillable[int32]{
			Value: DecimalStringToInt(*stringPointer),
			Valid: true,
		}
	}
	return transformedNillableInt
}
