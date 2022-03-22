package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIsIntArg(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isIntArg(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
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
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsIntGreaterArg(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isIntGreaterArg(0, 12),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
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
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsIntGreaterZeroArg(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isIntGreaterZeroArg(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
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
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsEC2InstanceType(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isEC2InstanceType(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"is instance type 1":    {[]string{"4xl"}, false},
		"is instance type 2":    {[]string{"12xlarge", "something else"}, false},
		"isn't instance type 1": {[]string{"notanInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsGCPInstanceType(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isGCPInstanceType(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"is instance type 1":    {[]string{"n2d-standard-4"}, false},
		"is instance type 2":    {[]string{"n2d-standard-16", "something else"}, false},
		"isn't instance type 1": {[]string{"notanInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestIsAzureInstanceType(t *testing.T) {
	testCmd := &cobra.Command{
		Use:  "test",
		Args: isAzureInstanceType(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"is instance type 1":    {[]string{"Standard_DC2as_v5"}, false},
		"is instance type 2":    {[]string{"Standard_DC8as_v5", "something else"}, false},
		"isn't instance type 1": {[]string{"notanInstanceType"}, true},
		"isn't instance type 2": {[]string{"Hello World!"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := testCmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
