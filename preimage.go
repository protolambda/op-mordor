package main

type PreimageOracle struct {
	// TODO cache
}

func (p *PreimageOracle) GetPreimage(key [32]byte) (v []byte, err error) {
	return nil, nil
}

func (p *PreimageOracle) SetPreimage(key [32]byte, value []byte) error {
	return nil
}
