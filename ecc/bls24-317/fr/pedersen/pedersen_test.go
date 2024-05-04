// Copyright 2020 Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package pedersen

import (
	"fmt"
	curve "github.com/consensys/gnark-crypto/ecc/bls24-317"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"github.com/consensys/gnark-crypto/utils/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func interfaceSliceToFrSlice(t *testing.T, values ...interface{}) []fr.Element {
	res := make([]fr.Element, len(values))
	for i, v := range values {
		_, err := res[i].SetInterface(v)
		assert.NoError(t, err)
	}
	return res
}

func randomFrSlice(t *testing.T, size int) []interface{} {
	res := make([]interface{}, size)
	var err error
	for i := range res {
		var v fr.Element
		res[i], err = v.SetRandom()
		assert.NoError(t, err)
	}
	return res
}

func randomOnG1() (curve.G1Affine, error) { // TODO: Add to G1.go?
	if gBytes, err := randomFrSizedBytes(); err != nil {
		return curve.G1Affine{}, err
	} else {
		return curve.HashToG1(gBytes, []byte("random on g1"))
	}
}

func randomG1Slice(t *testing.T, size int) []curve.G1Affine {
	res := make([]curve.G1Affine, size)
	for i := range res {
		var err error
		res[i], err = randomOnG1()
		assert.NoError(t, err)
	}
	return res
}

func testCommit(t *testing.T, values ...interface{}) {

	basis := randomG1Slice(t, len(values))

	var (
		pk              []ProvingKey
		vk              VerifyingKey
		err             error
		commitment, pok curve.G1Affine
	)
	valuesFr := interfaceSliceToFrSlice(t, values...)

	pk, vk, err = Setup(basis)
	assert.NoError(t, err)
	commitment, err = pk[0].Commit(valuesFr)
	assert.NoError(t, err)
	pok, err = pk[0].ProveKnowledge(valuesFr)
	assert.NoError(t, err)
	assert.NoError(t, vk.Verify(commitment, pok))

	pok.Neg(&pok)
	assert.NotNil(t, vk.Verify(commitment, pok))
}

func TestFoldProofs(t *testing.T) {

	values := [][]fr.Element{
		interfaceSliceToFrSlice(t, randomFrSlice(t, 5)...),
		interfaceSliceToFrSlice(t, randomFrSlice(t, 5)...),
		interfaceSliceToFrSlice(t, randomFrSlice(t, 5)...),
	}

	bases := make([][]curve.G1Affine, len(values))
	for i := range bases {
		bases[i] = randomG1Slice(t, len(values[i]))
	}

	pk, vk, err := Setup(bases...)
	assert.NoError(t, err)

	commitments := make([]curve.G1Affine, len(values))
	for i := range values {
		commitments[i], err = pk[i].Commit(values[i])
		assert.NoError(t, err)
	}

	t.Run("folding with zeros", func(t *testing.T) {
		pokFolded, err := BatchProve(pk[:2], [][]fr.Element{
			values[0],
			make([]fr.Element, len(values[1])),
		}, []byte("test"))
		assert.NoError(t, err)
		var pok curve.G1Affine
		pok, err = pk[0].ProveKnowledge(values[0])
		assert.NoError(t, err)
		assert.Equal(t, pok, pokFolded)
	})

	t.Run("run empty", func(t *testing.T) {
		var foldedCommitment curve.G1Affine
		pok, err := BatchProve([]ProvingKey{}, [][]fr.Element{}, []byte("test"))
		assert.NoError(t, err)

		foldedCommitment, err = FoldCommitments([]curve.G1Affine{}, []byte("test"))
		assert.NoError(t, err)
		assert.NoError(t, vk.Verify(foldedCommitment, pok))
	})

	run := func(values [][]fr.Element) func(t *testing.T) {
		return func(t *testing.T) {

			var foldedCommitment curve.G1Affine
			pok, err := BatchProve(pk[:len(values)], values, []byte("test"))
			assert.NoError(t, err)

			foldedCommitment, err = FoldCommitments(commitments[:len(values)], []byte("test"))
			assert.NoError(t, err)
			assert.NoError(t, vk.Verify(foldedCommitment, pok))

			pok.Neg(&pok)
			assert.NotNil(t, vk.Verify(foldedCommitment, pok))
		}
	}

	for i := range values {
		t.Run(fmt.Sprintf("folding %d", i+1), run(values[:i+1]))
	}
}

func TestCommitToOne(t *testing.T) {
	testCommit(t, 1)
}

func TestCommitSingle(t *testing.T) {
	testCommit(t, randomFrSlice(t, 1)...)
}

func TestCommitFiveElements(t *testing.T) {
	testCommit(t, randomFrSlice(t, 5)...)
}

func TestMarshal(t *testing.T) {
	var pk ProvingKey
	pk.BasisExpSigma = randomG1Slice(t, 5)
	pk.Basis = randomG1Slice(t, 5)

	var (
		vk  VerifyingKey
		err error
	)
	vk.G, err = randomOnG2()
	assert.NoError(t, err)
	vk.GRootSigmaNeg, err = randomOnG2()
	assert.NoError(t, err)

	t.Run("ProvingKey -> Bytes -> ProvingKey must remain identical.", testutils.SerializationRoundTrip(&pk))
	t.Run("ProvingKey -> Bytes (raw) -> ProvingKey must remain identical.", testutils.SerializationRoundTripRaw(&pk))
	t.Run("VerifyingKey -> Bytes -> VerifyingKey must remain identical.", testutils.SerializationRoundTrip(&vk))
	t.Run("VerifyingKey -> Bytes (raw) -> ProvingKey must remain identical.", testutils.SerializationRoundTripRaw(&vk))
}
