package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEnvironmentVars(t *testing.T) {
	vars, err := NewEnvironmentVars()

	assert.Nil(t, err)
	assert.Equal(t, EnvironmentVars{
		Port:            "1234",
		WebDir:          "web",
		SiriusURL:       "http://host.docker.internal:8080",
		SiriusPublicURL: "",
		Prefix:          "",
	}, vars)
}
