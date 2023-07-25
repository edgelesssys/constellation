/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package internal

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

type TemplatePreparer struct{}

func NewTemplatePreparer() TemplatePreparer {
	return TemplatePreparer{}
}

func (p TemplatePreparer) template(fs embed.FS, templateFile string, templateData any) (*bytes.Buffer, error) {
	templates, err := template.ParseFS(fs, templateFile)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	buf := bytes.NewBuffer(nil)

	if err = templates.Execute(buf, templateData); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf, nil
}
