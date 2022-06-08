package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/file"
)

func readConfig(out io.Writer, fileHandler file.Handler, name string, provider cloudprovider.Provider) (*config.Config, error) {
	cnf, err := config.FromFile(fileHandler, name)
	if err != nil {
		return nil, err
	}
	if err := validateConfig(out, cnf, provider); err != nil {
		return nil, err
	}
	return cnf, nil
}

func validateConfig(out io.Writer, cnf *config.Config, provider cloudprovider.Provider) error {
	msgs, err := cnf.Validate()
	if err != nil {
		return err
	}

	if len(msgs) > 0 {
		fmt.Fprintln(out, "Invalid fields in config file:")
		for _, m := range msgs {
			fmt.Fprintln(out, "\t"+m)
		}
		return errors.New("invalid configuration")
	}

	if provider != cloudprovider.Unknown && !cnf.HasProvider(provider) {
		return fmt.Errorf("configuration doesn't contain provider: %v", provider)
	}

	return nil
}
