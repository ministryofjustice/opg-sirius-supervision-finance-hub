package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownloadRequest_Decode(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected DownloadRequest
		fail     bool
	}{
		{
			name: "success",
			in:   "eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=",
			expected: DownloadRequest{
				Key:       "test.csv",
				VersionId: "vprAxsYtLVsb5P9H_qHeNUiU9MBnPNcz",
			},
			fail: false,
		},
		{
			name:     "fail",
			in:       "notavalidstring=",
			expected: DownloadRequest{},
			fail:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d DownloadRequest
			err := d.Decode(tt.in)
			if tt.fail {
				assert.Error(t, err)
			} else {
				assert.EqualValuesf(t, tt.expected, d, "Decode()")
			}
		})
	}
}

func TestDownloadRequest_Encode(t *testing.T) {
	d := &DownloadRequest{
		Key:       "test.csv",
		VersionId: "vprAxsYtLVsb5P9H_qHeNUiU9MBnPNcz",
	}
	got, _ := d.Encode()
	assert.Equalf(t, "eyJLZXkiOiJ0ZXN0LmNzdiIsIlZlcnNpb25JZCI6InZwckF4c1l0TFZzYjVQOUhfcUhlTlVpVTlNQm5QTmN6In0=", got, "Encode()")
}
