package cmd

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/cli/internal/azure"
	"github.com/edgelesssys/constellation/cli/internal/gcp"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
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
		"gcp":           {[]string{"gcp"}, false},
		"azure":         {[]string{"azure"}, false},
		"aws waring":    {[]string{"aws"}, true},
		"too many args": {[]string{"gcp", "1", "2"}, true},
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
		setupFs             func(*require.Assertions) afero.Fs
		creator             *stubCloudCreator
		provider            cloudprovider.Provider
		yesFlag             bool
		controllerCountFlag *int
		workerCountFlag     *int
		insTypeFlag         string
		configFlag          string
		nameFlag            string
		stdin               string
		wantErr             bool
		wantAbbort          bool
	}{
		"create": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{state: testState},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(2),
			yesFlag:             true,
		},
		"interactive": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{state: testState},
			provider:            cloudprovider.Azure,
			controllerCountFlag: intPtr(2),
			workerCountFlag:     intPtr(1),
			stdin:               "yes\n",
		},
		"interactive abort": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			stdin:               "no\n",
			wantAbbort:          true,
		},
		"interactive error": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			stdin:               "foo\nfoo\nfoo\n",
			wantErr:             true,
		},
		"flag name to long": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			nameFlag:            strings.Repeat("a", constants.ConstellationNameLength+1),
			wantErr:             true,
		},
		"flag control-plane-count invalid": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(0),
			workerCountFlag:     intPtr(3),
			wantErr:             true,
		},
		"flag worker-count invalid": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(3),
			workerCountFlag:     intPtr(0),
			wantErr:             true,
		},
		"flag control-plane-count missing": {
			setupFs:         func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:         &stubCloudCreator{},
			provider:        cloudprovider.GCP,
			workerCountFlag: intPtr(3),
			wantErr:         true,
		},
		"flag worker-count missing": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(3),
			wantErr:             true,
		},
		"flag invalid instance-type": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			insTypeFlag:         "invalid",
			wantErr:             true,
		},
		"old state in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.StateFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"old adminConf in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"old masterSecret in directory": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.MasterSecretFilename, []byte{1}, file.OptNone))
				return fs
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"config does not exist": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			configFlag:          constants.ConfigFilename,
			wantErr:             true,
		},
		"create error": {
			setupFs:             func(require *require.Assertions) afero.Fs { return afero.NewMemMapFs() },
			creator:             &stubCloudCreator{createErr: someErr},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"write state error": {
			setupFs: func(require *require.Assertions) afero.Fs {
				fs := afero.NewMemMapFs()
				return afero.NewReadOnlyFs(fs)
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newCreateCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			cmd.Flags().String("config", "", "") // register persisten flag manually
			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}
			if tc.nameFlag != "" {
				require.NoError(cmd.Flags().Set("name", tc.nameFlag))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}
			if tc.controllerCountFlag != nil {
				require.NoError(cmd.Flags().Set("control-plane-nodes", strconv.Itoa(*tc.controllerCountFlag)))
			}
			if tc.workerCountFlag != nil {
				require.NoError(cmd.Flags().Set("worker-nodes", strconv.Itoa(*tc.workerCountFlag)))
			}
			if tc.insTypeFlag != "" {
				require.NoError(cmd.Flags().Set("instance-type", tc.insTypeFlag))
			}

			fileHandler := file.NewHandler(tc.setupFs(require))

			err := create(cmd, tc.creator, fileHandler, tc.provider)

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
			args:        []string{"gcp", "foo"},
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

func TestInstanceTypeCompletion(t *testing.T) {
	testCases := map[string]struct {
		args        []string
		wantResult  []string
		wantShellCD cobra.ShellCompDirective
	}{
		"azure": {
			args:        []string{"azure"},
			wantResult:  azure.InstanceTypes,
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"gcp": {
			args:        []string{"gcp"},
			wantResult:  gcp.InstanceTypes,
			wantShellCD: cobra.ShellCompDirectiveNoFileComp,
		},
		"empty args": {
			args:        []string{},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
		"unknown provider": {
			args:        []string{"foo"},
			wantResult:  []string{},
			wantShellCD: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := instanceTypeCompletion(cmd, tc.args, "")

			assert.Equal(tc.wantResult, result)
			assert.Equal(tc.wantShellCD, shellCD)
		})
	}
}

func intPtr(i int) *int {
	return &i
}
