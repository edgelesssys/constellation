package passwd

import (
	"github.com/spf13/afero"
	"github.com/willdonnelly/passwd"
)

// An Entry contains all the fields for a specific user. Re-exported to allow other module to only import this passwd module.
type Entries map[string]passwd.Entry

type Passwd struct{}

// Parse opens the '/etc/passwd' file and parses it into a map from usernames to Entries.
func (p Passwd) Parse(fs afero.Fs) (Entries, error) {
	return p.parseFile(fs, "/etc/passwd")
}

// parseFile opens the file and parses it into a map from usernames to Entries.
func (p Passwd) parseFile(fs afero.Fs, path string) (Entries, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries, err := passwd.ParseReader(file)
	return Entries(entries), err
}
