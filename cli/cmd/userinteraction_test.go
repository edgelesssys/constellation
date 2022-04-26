package cmd

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAskToConfirm(t *testing.T) {
	// errAborted is an error where the user aborted the action.
	errAborted := errors.New("user aborted")

	cmd := &cobra.Command{
		Use:  "test",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := askToConfirm(cmd, "777")
			if err != nil {
				return err
			}
			if !ok {
				return errAborted
			}
			return nil
		},
	}

	testCases := map[string]struct {
		input   string
		wantErr error
	}{
		"user confirms":                       {"y\n", nil},
		"user confirms long":                  {"yes\n", nil},
		"user disagrees":                      {"n\n", errAborted},
		"user disagrees long":                 {"no\n", errAborted},
		"user is first unsure, but agrees":    {"what?\ny\n", nil},
		"user is first unsure, but disagrees": {"wait.\nn\n", errAborted},
		"repeated invalid input":              {"h\nb\nq\n", ErrInvalidInput},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			in := bytes.NewBufferString(tc.input)
			cmd.SetIn(in)

			err := cmd.Execute()
			assert.ErrorIs(err, tc.wantErr)

			output, err := io.ReadAll(out)
			assert.NoError(err)
			assert.Contains(string(output), "777")
		})
	}
}

func TestWarnAboutPCRs(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")

	testCases := map[string]struct {
		pcrs         map[uint32][]byte
		dontWarnInit bool
		wantWarnings []string
		wantErr      bool
	}{
		"no warnings": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				6:  zero,
				7:  zero,
				8:  zero,
				9:  zero,
				10: zero,
				11: zero,
				12: zero,
			},
		},
		"no warnings for missing non critical values": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
		},
		"warn for BIOS": {
			pcrs: map[uint32][]byte{
				0:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"BIOS"},
		},
		"warn for OPROM": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"OPROM"},
		},
		"warn for MBR": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"MBR"},
		},
		"warn for kernel": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"kernel"},
		},
		"warn for initrd": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"initrd"},
		},
		"warn for initialization": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
			},
			dontWarnInit: false,
			wantWarnings: []string{"initialization"},
		},
		"don't warn for initialization": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
			},
			dontWarnInit: true,
		},
		"multi warning": {
			pcrs: map[uint32][]byte{},
			wantWarnings: []string{
				"BIOS",
				"OPROM",
				"MBR",
				"initialization",
				"initrd",
				"kernel",
			},
		},
		"bad config": {
			pcrs: map[uint32][]byte{
				0: []byte("000"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newInitCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			var errOut bytes.Buffer
			cmd.SetErr(&errOut)

			err := warnAboutPCRs(cmd, tc.pcrs, !tc.dontWarnInit)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if len(tc.wantWarnings) == 0 {
					assert.Empty(errOut.String())
				} else {
					for _, warning := range tc.wantWarnings {
						assert.Contains(errOut.String(), warning)
					}
				}
			}
		})
	}
}
