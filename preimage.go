package main

import (
	"errors"
)

var ErrMissingData = errors.New("missing data")

type PreimageOracle struct {
	data map[[32]byte][]byte
}

func NewPreimageOracle() *PreimageOracle {
	return &PreimageOracle{
		data: make(map[[32]byte][]byte),
	}
}

func (p *PreimageOracle) GetPreimage(key [32]byte) (v []byte, err error) {
	bytes, ok := p.data[key]
	if !ok {
		return nil, ErrMissingData
	}
	return bytes, nil
}

func (p *PreimageOracle) SetPreimage(key [32]byte, value []byte) error {
	p.data[key] = value
	return nil
}
