package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var envTests = []struct {
	name     string
	json     string
	expected WebAssetServerSetting
}{
	{
		"empty",
		`{"JWT_SECRET": "sekret"}`,
		WebAssetServerSetting{Host: "127.0.0.1", Port: 3000},
	},
	{
		"full",
		`{
			"JWT_SECRET": "sekret",
			"WEB_ASSET_SERVER": { "port": 9090, "host": "0.0.0.0" }
		 }`,
		WebAssetServerSetting{Host: "0.0.0.0", Port: 9090},
	},
}

func TestNewWebAssetServer(t *testing.T) {
	for _, tt := range envTests {
		t.Run(tt.name, func(t *testing.T) {
			actual := new(ServerConfig)
			dec := json.NewDecoder(strings.NewReader(tt.json))
			err := dec.Decode(actual)
			require.NoError(t, err)
			initFromConfig(actual)
			require.Equal(t, tt.expected, actual.WEB_ASSET_SERVER)
		})
	}
}
