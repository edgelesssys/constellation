package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"gcp valid create 1":                {[]string{"gcp", "3", "3", "n2d-standard-2"}, false},
		"gcp valid create 2":                {[]string{"gcp", "3", "7", "n2d-standard-16"}, false},
		"gcp valid create 3":                {[]string{"gcp", "1", "2", "n2d-standard-96"}, false},
		"gcp invalid too many arguments":    {[]string{"gcp", "3", "2", "n2d-standard-2", "n2d-standard-2"}, true},
		"gcp invalid too many arguments 2":  {[]string{"gcp", "3", "2", "n2d-standard-2", "2"}, true},
		"gcp invalid no coordinators":       {[]string{"gcp", "0", "1", "n2d-standard-2"}, true},
		"gcp invalid no nodes":              {[]string{"gcp", "1", "0", "n2d-standard-2"}, true},
		"gcp invalid first is no int":       {[]string{"gcp", "n2d-standard-2", "1", "n2d-standard-2"}, true},
		"gcp invalid second is no int":      {[]string{"gcp", "3", "n2d-standard-2", "n2d-standard-2"}, true},
		"gcp invalid third is no size":      {[]string{"gcp", "2", "2", "2"}, true},
		"gcp invalid wrong order":           {[]string{"gcp", "n2d-standard-2", "2", "2"}, true},
		"azure valid create 1":              {[]string{"azure", "3", "3", "Standard_DC2as_v5"}, false},
		"azure valid create 2":              {[]string{"azure", "3", "7", "Standard_DC4as_v5"}, false},
		"azure valid create 3":              {[]string{"azure", "1", "2", "Standard_DC8as_v5"}, false},
		"azure invalid to many arguments":   {[]string{"azure", "3", "2", "Standard_DC2as_v5", "Standard_DC2as_v5"}, true},
		"azure invalid to many arguments 2": {[]string{"azure", "3", "2", "Standard_DC2as_v5", "2"}, true},
		"azure invalid no coordinators":     {[]string{"azure", "0", "1", "Standard_DC2as_v5"}, true},
		"azure invalid no nodes":            {[]string{"azure", "1", "0", "Standard_DC2as_v5"}, true},
		"azure invalid first is no int":     {[]string{"azure", "Standard_DC2as_v5", "1", "Standard_DC2as_v5"}, true},
		"azure invalid second is no int":    {[]string{"azure", "1", "Standard_DC2as_v5", "Standard_DC2as_v5"}, true},
		"azure invalid third is no size":    {[]string{"azure", "2", "2", "2"}, true},
		"azure invalid wrong order":         {[]string{"azure", "Standard_DC2as_v5", "2", "2"}, true},
		"aws waring":                        {[]string{"aws", "1", "2", "4xlarge"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := newCreateCmd().ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	testState := state.ConstellationState{Name: "test"}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		setupFs       func(*require.Assertions) afero.Fs
		creator       *stubCloudCreator
		provider      cloudprovider.CloudProvider
		yesFlag       bool
		devConfigFlag string
		nameFlag      string
		stdin         string
		wantErr       bool
		wantAbbort    bool
	}{
		"create": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{state: testState},
			provider: cloudprovider.GCP,
			yesFlag:  true,
		},
		"interactive": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{state: testState},
			provider: cloudprovider.GCP,
			stdin:    "yes\n",
		},
		"interactive abort": {
			setupFs:    func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:    &stubCloudCreator{},
			provider:   cloudprovider.GCP,
			stdin:      "no\n",
			wantAbbort: true,
		},
		"interactive error": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			stdin:    "foo\nfoo\nfoo\n",
			wantErr:  true,
		},
		"flag name to long": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			nameFlag: strings.Repeat("a", constellationNameLength+1),
			wantErr:  true,
		},
		"old state in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.StateFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"old adminConf in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"old masterSecret in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.MasterSecretFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"dev config does not exist": {
			setupFs:       func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:       &stubCloudCreator{},
			provider:      cloudprovider.GCP,
			yesFlag:       true,
			devConfigFlag: "dev-config.json",
			wantErr:       true,
		},
		"create error": {
			setupFs:  func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{createErr: someErr},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"write state error": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				return afero.NewReadOnlyFs(fs)
			},
			creator:  &stubCloudCreator{},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newCreateCmd()
			cmd.Flags().String("dev-config", "", "") // register persisten flag manually
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}
			if tc.nameFlag != "" {
				require.NoError(cmd.Flags().Set("name", tc.nameFlag))
			}
			if tc.devConfigFlag != "" {
				require.NoError(cmd.Flags().Set("dev-config", tc.devConfigFlag))
			}
			fileHandler := file.NewHandler(tc.setupFs(require))

			err := create(cmd, tc.creator, fileHandler, 3, 3, tc.provider, "type")

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.wantAbbort {
					assert.False(tc.creator.createCalled)
				} else {
					assert.True(tc.creator.createCalled)
					var state state.ConstellationState
					require.NoError(fileHandler.ReadJSON(constants.StateFilename, &state))
					assert.Equal(state, testState)
				}
			}
		})
	}
}

func TestCheckDirClean(t *testing.T) {
	testCases := map[string]struct {
		fileHandler   file.Handler
		existingFiles []string
		wantErr       bool
	}{
		"no file exists": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
		},
		"adminconf exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename},
			wantErr:       true,
		},
		"master secret exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.MasterSecretFilename},
			wantErr:       true,
		},
		"state file exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.StateFilename},
			wantErr:       true,
		},
		"multiple exist": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename, constants.MasterSecretFilename, constants.StateFilename},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			for _, f := range tc.existingFiles {
				require.NoError(tc.fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
			}

			err := checkDirClean(tc.fileHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestCreateCompletion(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		wantResult  []string
		wantShellCD cobra.ShellCompDirective
	}{
		"first arg": {
			args:        []string{},
			wantResult:  []string{"aws", "gcp", "azure"},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"second arg": {
			args:        []string{"gcp"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"third arg": {
			args:        []string{"gcp", "1"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"fourth arg aws": {
			args: []string{"aws", "1", "2"},
			wantResult: []string{
				"4xlarge",
				"8xlarge",
				"12xlarge",
				"16xlarge",
				"24xlarge",
			},
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"fourth arg gcp": {
			args:        []string{"gcp", "1", "2"},
			wantResult:  gcp.InstanceTypes,
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"fourth arg azure": {
			args:        []string{"azure", "1", "2"},
			wantResult:  azure.InstanceTypes,
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"fifth arg": {
			args:        []string{"aws", "1", "2", "4xlarge"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := createCompletion(cmd, tc.args, "")
			assert.Equal(tc.wantResult, result)
			assert.Equal(tc.wantShellCD, shellCD)
		})
	}
}
