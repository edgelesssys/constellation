package client

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateInstanceTemplateName(t *testing.T) {
	testCases := map[string]struct {
		last     string
		wantNext string
		wantErr  bool
	}{
		"no numbering yet": {
			last:     "prefix",
			wantNext: "prefix-1",
		},
		"ends in -": {
			last:     "prefix-",
			wantNext: "prefix-1",
		},
		"has number": {
			last:     "prefix-1",
			wantNext: "prefix-2",
		},
		"last number too small": {
			last:    "prefix-0",
			wantErr: true,
		},
		"last number would overflow": {
			last:    fmt.Sprintf("prefix-%d", math.MaxInt),
			wantErr: true,
		},
		"integer out of range": {
			last:    "prefix-999999999999999999999999999999999999999999",
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			gotNext, err := generateInstanceTemplateName(tc.last)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNext, gotNext)
		})
	}
}

func TestSplitInstanceTemplateID(t *testing.T) {
	testCases := map[string]struct {
		instanceTemplateID string

		wantProject      string
		wantTemplateName string
		wantErr          bool
	}{
		"valid request": {
			instanceTemplateID: "projects/project/global/instanceTemplates/template",
			wantProject:        "project",
			wantTemplateName:   "template",
		},
		"wrong format": {
			instanceTemplateID: "wrong-format",
			wantErr:            true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			gotProject, gotTemplateName, err := splitInstanceTemplateID(tc.instanceTemplateID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantProject, gotProject)
			assert.Equal(tc.wantTemplateName, gotTemplateName)
		})
	}
}

func TestJoinInstanceTemplateID(t *testing.T) {
	assert := assert.New(t)
	project := "project"
	templateName := "template"
	wantInstanceTemplateURI := "https://www.googleapis.com/compute/v1/projects/project/global/instanceTemplates/template"
	gotInstancetemplateURI := joinInstanceTemplateURI(project, templateName)
	assert.Equal(wantInstanceTemplateURI, gotInstancetemplateURI)
}
