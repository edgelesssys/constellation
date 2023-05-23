package snpversion

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestVersionMarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		sut  Version
		want string
	}{
		{
			name: "isLatest resolves to latest",
			sut: Version{
				Value:    1,
				IsLatest: true,
			},
			want: "latest\n",
		},
		{
			name: "value 5 resolves to 5",
			sut: Version{
				Value:    5,
				IsLatest: false,
			},
			want: "5\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt, err := yaml.Marshal(tt.sut)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(bt))
		})
	}
}
