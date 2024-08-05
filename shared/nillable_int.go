package shared

type NillableInt struct {
	Value int
	Valid bool
}

func TransformNillableInt(stringPointer *string) NillableInt {
	var transformedNillableInt NillableInt
	if stringPointer != nil {
		transformedNillableInt = NillableInt{
			Value: DecimalStringToInt(*stringPointer),
			Valid: true,
		}
	}
	return transformedNillableInt
}
