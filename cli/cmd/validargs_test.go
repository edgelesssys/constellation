package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIsIntArg(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"valid int 1":                {[]string{"1"}, false},
		"valid int 2":                {[]string{"42"}, false},
		"valid int 3":                {[]string{"987987498"}, false},
		"valid int and other args":   {[]string{"3", "hello"}, false},
		"valid int and other args 2": {[]string{"3", "4"}, false},
		"invalid 1":                  {[]string{"hello world"}, true},
		"invalid 2":                  {[]string{"98798d749f8"}, true},
		"invalid 3":                  {[]string{"three"}, true},
		"invalid 4":                  {[]string{"0.3"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isIntArg(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsIntGreaterArg(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"valid int 1":                  {[]string{"13"}, false},
		"valid int 2":                  {[]string{"42"}, false},
		"valid int 3":                  {[]string{"987987498"}, false},
		"invalid int 1":                {[]string{"1"}, true},
		"invalid int and other args":   {[]string{"3", "hello"}, true},
		"invalid int and other args 2": {[]string{"-14", "4"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isIntGreaterArg(0, 12)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsIntGreaterZeroArg(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"valid int 1":                {[]string{"13"}, false},
		"valid int 2":                {[]string{"42"}, false},
		"valid int 3":                {[]string{"987987498"}, false},
		"invalid":                    {[]string{"0"}, true},
		"invalid int 1":              {[]string{"-42", "hello"}, true},
		"invalid int and other args": {[]string{"-9487239847", "4"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isIntGreaterZeroArg(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsPort(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"valid port 1":   {[]string{"80"}, false},
		"valid port 2":   {[]string{"8080"}, false},
		"valid port 3":   {[]string{"65535"}, false},
		"invalid port 1": {[]string{"foo"}, true},
		"invalid port 2": {[]string{"65536"}, true},
		"invalid port 3": {[]string{"-1"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isPort(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsIP(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"valid ip 1":   {[]string{"192.168.0.2"}, false},
		"valid ip 2":   {[]string{"127.0.0.1"}, false},
		"valid ip 3":   {[]string{"8.8.8.8"}, false},
		"invalid ip 1": {[]string{"foo"}, true},
		"invalid ip 2": {[]string{"foo.bar.baz.1"}, true},
		"invalid ip 3": {[]string{"800.800.800.800"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isIP(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsEC2InstanceType(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"is instance type 1":    {[]string{"4xl"}, false},
		"is instance type 2":    {[]string{"12xlarge", "something else"}, false},
		"isn't instance type 1": {[]string{"notAnInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isEC2InstanceType(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsGCPInstanceType(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"is instance type 1":    {[]string{"n2d-standard-4"}, false},
		"is instance type 2":    {[]string{"n2d-standard-16", "something else"}, false},
		"isn't instance type 1": {[]string{"notAnInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isGCPInstanceType(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsAzureInstanceType(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"is instance type 1":    {[]string{"Standard_DC2as_v5"}, false},
		"is instance type 2":    {[]string{"Standard_DC8as_v5", "something else"}, false},
		"isn't instance type 1": {[]string{"notAnInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isAzureInstanceType(0)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsCloudProvider(t *testing.T) {
	testCases := map[string]struct {
		pos     int
		args    []string
		wantErr bool
	}{
		"gcp":     {0, []string{"gcp"}, false},
		"azure":   {1, []string{"foo", "azure"}, false},
		"foo":     {0, []string{"foo"}, true},
		"empty":   {0, []string{""}, true},
		"unknown": {0, []string{"unknown"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isCloudProvider(tc.pos)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsInstanceTypeForProvider(t *testing.T) {
	testCases := map[string]struct {
		typePos     int
		providerPos int
		args        []string
		wantErr     bool
	}{
		"valid gcp type 1":          {1, 0, []string{"gcp", "n2d-standard-4"}, false},
		"valid gcp type 2":          {1, 0, []string{"gcp", "n2d-standard-16", "foo"}, false},
		"valid azure type 1":        {1, 0, []string{"azure", "Standard_DC2as_v5"}, false},
		"valid azure type 2":        {1, 0, []string{"azure", "Standard_DC8as_v5", "foo"}, false},
		"mixed order 1":             {0, 3, []string{"n2d-standard-4", "", "foo", "gcp"}, false},
		"mixed order 2":             {2, 1, []string{"", "gcp", "n2d-standard-4", "foo", "bar"}, false},
		"invalid gcp type":          {1, 0, []string{"gcp", "foo"}, true},
		"invalid azure type":        {1, 0, []string{"azure", "foo"}, true},
		"args to short":             {2, 0, []string{"foo"}, true},
		"provider position invalid": {1, 2, []string{"gcp", "n2d-standard-4"}, true},
		"type position invalid":     {2, 0, []string{"gcp", "n2d-standard-4"}, true},
		"unknown provider":          {1, 0, []string{"foo", "n2d-standard-4"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isInstanceTypeForProvider(tc.typePos, tc.providerPos)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
