// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package shplonk

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bn254"
)

func (proof *OpeningProof) ReadFrom(r io.Reader) (int64, error) {

	dec := bn254.NewDecoder(r)

	toDecode := []interface{}{
		&proof.W,
		&proof.WPrime,
		&proof.ClaimedValues,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of a OpeningProof
func (proof *OpeningProof) WriteTo(w io.Writer) (int64, error) {

	enc := bn254.NewEncoder(w)

	toEncode := []interface{}{
		&proof.W,
		&proof.WPrime,
		proof.ClaimedValues,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}
