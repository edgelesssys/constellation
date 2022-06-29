package logging

import "io"

// CloudLogger is used to log information to a **non-confidential** destination
// at cloud provider for early-boot debugging. Make sure to **NOT** include any
// sensitive information!
type CloudLogger interface {
	// Disclose is used to log information into a **non-confidential** destination at
	// cloud provider for early-boot debugging. Make sure to **NOT** Disclose any
	// sensitive information!
	Disclose(msg string)
	io.Closer
}

type NopLogger struct{}

func (l *NopLogger) Disclose(msg string) {}

func (l *NopLogger) Close() error {
	return nil
}
