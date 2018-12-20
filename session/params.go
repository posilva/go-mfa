package session

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	defaultMFADuration = 3600
)

// Params are used to create the instance
type Params struct {
	SerialDevice string
	Profile      string
	MFAToken     string
	MFADuration  int64
}

// DefaultParams returns a minimal session parameters
func DefaultParams(serial string, token string) Params {
	return Params{
		SerialDevice: serial,
		MFAToken:     token,
		MFADuration:  defaultMFADuration,
	}
}

// ToOptions returns session options based on the parameters
func (p *Params) ToOptions() session.Options {
	if p.MFADuration == 0 {
		p.MFADuration = 3600
	}
	if p.Profile != "" {
		return session.Options{Profile: p.Profile}
	}
	return session.Options{}
}
