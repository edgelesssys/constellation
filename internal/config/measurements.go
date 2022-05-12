package config

import "encoding/base64"

type Measurements map[uint32][]byte

func (m Measurements) MarshalYAML() (interface{}, error) {
	base64Map := make(map[uint32]string)

	for key, value := range m {
		base64Map[key] = base64.StdEncoding.EncodeToString(value[:])
	}

	return base64Map, nil
}

func (m *Measurements) UnmarshalYAML(unmarshal func(interface{}) error) error {
	base64Map := make(map[uint32]string)
	err := unmarshal(base64Map)
	if err != nil {
		return err
	}

	*m = make(Measurements)
	for key, value := range base64Map {
		measurement, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
		(*m)[key] = measurement
	}
	return nil
}
