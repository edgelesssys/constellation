package user

import (
	"bufio"
	"errors"
	"strings"

	"github.com/spf13/afero"
)

// Entry is an entry of a '/etc/passwd' file.
type Entry struct {
	Password  string
	UID       string
	GID       string
	GECOS     string
	Directory string
	Shell     string
}

// Entries contains the information for each user defined in '/etc/passwd'.
type Entries map[string]Entry

// Passwd allows to parse users from '/etc/passwd' on the local system.
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

	entries := Entries{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// File format: https://man7.org/linux/man-pages/man5/passwd.5.html

		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 7 {
			return nil, errors.New("invalid number of fields")
		}

		entries[fields[0]] = Entry{
			Password:  fields[1],
			UID:       fields[2],
			GID:       fields[3],
			GECOS:     fields[4],
			Directory: fields[5],
			Shell:     fields[6],
		}
	}

	return entries, scanner.Err()
}
