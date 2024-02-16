package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewFinanceVars(t *testing.T) {
	r, _ := http.NewRequest("GET", "/path", nil)

	envVars := EnvironmentVars{}
	vars, err := NewAppVars(r, envVars)

	assert.Nil(t, err)
	assert.Equal(t, AppVars{
		Path:            "/path",
		XSRFToken:       "",
		EnvironmentVars: envVars,
	}, vars)
}
