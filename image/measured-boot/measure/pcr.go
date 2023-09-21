/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measure

import (
	"crypto/sha256"
	"fmt"
)

// PCR256 is a 256-bit PCR value.
type PCR256 [32]byte

// MarshalJSON implements json.Marshaler.
func (p PCR256) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"expected\": \"%x\"}", p[:])), nil
}

// Digest256 is a 256-bit digest value (sha256).
type Digest256 [32]byte

// MarshalJSON implements json.Marshaler.
func (d Digest256) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%x\"", d[:])), nil
}

// PCR256Bank is a map of PCR index to PCR256 value.
type PCR256Bank map[uint32]PCR256

// Event is a pcr extend event.
type Event struct {
	PCRIndex    uint32
	Digest      Digest256
	Data        []byte `json:",omitempty"`
	Description string
}

// EventLog is a list of events.
type EventLog struct {
	Events []Event
}

// Simulator is a TPM PCR simulator.
type Simulator struct {
	Bank     PCR256Bank `json:"measurements"`
	EventLog EventLog
}

// NewDefaultSimulator returns a new Simulator with default PCR values.
func NewDefaultSimulator() *Simulator {
	return &Simulator{
		Bank: PCR256Bank{
			4:  ZeroPCR256(),
			8:  ZeroPCR256(),
			9:  ZeroPCR256(),
			11: ZeroPCR256(),
			12: ZeroPCR256(),
			13: ZeroPCR256(),
			15: ZeroPCR256(),
		},
	}
}

// ExtendPCR extends the PCR at index with the digest and data.
func (s *Simulator) ExtendPCR(index uint32, digest [32]byte, data []byte, description string) error {
	hashCtx := sha256.New()

	old, ok := s.Bank[index]
	if !ok {
		return fmt.Errorf("PCR index %d not found", index)
	}

	hashCtx.Write(old[:])
	hashCtx.Write(digest[:])
	newHash := hashCtx.Sum(nil)
	s.Bank[index] = PCR256(newHash)

	var eventData []byte
	if data != nil {
		eventData = make([]byte, len(data))
		copy(eventData, data)
	}

	s.EventLog.Events = append(s.EventLog.Events, Event{
		PCRIndex:    index,
		Digest:      digest,
		Data:        eventData,
		Description: description,
	})

	return nil
}

// ZeroPCR256 returns a zeroed PCR256 value.
func ZeroPCR256() PCR256 {
	return PCR256{}
}

// EVEFIActionPCR256 returns the expected PCR256 value for EV_EFI_ACTION.
func EVEFIActionPCR256() PCR256 {
	return PCR256{
		0x3d, 0x67, 0x72, 0xb4, 0xf8, 0x4e, 0xd4, 0x75,
		0x95, 0xd7, 0x2a, 0x2c, 0x4c, 0x5f, 0xfd, 0x15,
		0xf5, 0xbb, 0x72, 0xc7, 0x50, 0x7f, 0xe2, 0x6f,
		0x2a, 0xae, 0xe2, 0xc6, 0x9d, 0x56, 0x33, 0xba,
	}
}

// EVSeparatorPCR256 returns the expected PCR256 value for EV_SEPARATOR.
func EVSeparatorPCR256() PCR256 {
	return PCR256{
		0xdf, 0x3f, 0x61, 0x98, 0x04, 0xa9, 0x2f, 0xdb,
		0x40, 0x57, 0x19, 0x2d, 0xc4, 0x3d, 0xd7, 0x48,
		0xea, 0x77, 0x8a, 0xdc, 0x52, 0xbc, 0x49, 0x8c,
		0xe8, 0x05, 0x24, 0xc0, 0x14, 0xb8, 0x11, 0x19,
	}
}
