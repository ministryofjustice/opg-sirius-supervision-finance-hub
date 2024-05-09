package validation

import (
	"github.com/go-playground/validator/v10"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Validate
	}{
		{
			name: "can create a new validator",
			want: New(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New()
			if got.validator == nil {
				t.Errorf("New() = %v, validator is nil", got)
			}
		})
	}
}

func TestValidate_RegisterValidation(t *testing.T) {
	type fields struct {
		validator *validator.Validate
	}
	type args struct {
		tag string
		fn  validator.Func
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Register new validation",
			fields: fields{
				validator: validator.New(),
			},
			args: args{
				tag: "customTag",
				fn: func(fl validator.FieldLevel) bool {
					return fl.Field().String() == "valid"
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validate{
				validator: tt.fields.validator,
			}
			v.RegisterValidation(tt.args.tag, tt.args.fn)
		})
	}
}

//func TestValidate_ValidateDateInThePast(t *testing.T) {
//	type fields struct {
//		validator *validator.Validate
//	}
//	type args struct {
//		fl validator.FieldLevel
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   bool
//	}{
//		{
//			name:   "The date is in the past",
//			fields: fields{validator: validator.New()},
//			args:   args{fl: validator.},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &Validate{
//				validator: tt.fields.validator,
//			}
//			if got := v.ValidateDateInThePast(tt.args.fl); got != tt.want {
//				t.Errorf("ValidateDateInThePast() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestValidate_ValidateStruct(t *testing.T) {
//	type fields struct {
//		validator *validator.Validate
//	}
//	type args struct {
//		s interface{}
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   shared.ValidationError
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &Validate{
//				validator: tt.fields.validator,
//			}
//			if got := v.ValidateStruct(tt.args.s); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("ValidateStruct() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestValidate_ValidateThousandCharacterCount(t *testing.T) {
//	type fields struct {
//		validator *validator.Validate
//	}
//	type args struct {
//		fl validator.FieldLevel
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			v := &Validate{
//				validator: tt.fields.validator,
//			}
//			if got := v.ValidateThousandCharacterCount(tt.args.fl); got != tt.want {
//				t.Errorf("ValidateThousandCharacterCount() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
