package shared

type NillableDate struct {
	Value Date
	Valid bool
}

func TransformNillableDate(stringPointer *string) NillableDate {
	var transformedNillableDate NillableDate
	if stringPointer != nil {
		transformedNillableDate = NillableDate{
			NewDate(*stringPointer),
			true,
		}
	}
	return transformedNillableDate
}
