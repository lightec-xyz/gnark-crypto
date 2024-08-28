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
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bw6-756"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
	"github.com/consensys/gnark-crypto/utils/testutils"
	"github.com/stretchr/testify/assert"
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

	pk, vk, err = Setup([][]curve.G1Affine{basis})
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

	pk, vk, err := Setup(bases)
	assert.NoError(t, err)

	commitments := make([]curve.G1Affine, len(values))
	for i := range values {
		commitments[i], err = pk[i].Commit(values[i])
		assert.NoError(t, err)
	}

	hashes, err := fr.Hash([]byte("test"), []byte("pedersen"), 1)
	assert.NoError(t, err)

	t.Run("folding with zeros", func(t *testing.T) {
		pokFolded, err := BatchProve(pk[:2], [][]fr.Element{
			values[0],
			make([]fr.Element, len(values[1])),
		}, hashes[0])
		assert.NoError(t, err)
		var pok curve.G1Affine
		pok, err = pk[0].ProveKnowledge(values[0])
		assert.NoError(t, err)
		assert.Equal(t, pok, pokFolded)
	})

	t.Run("run empty", func(t *testing.T) {
		var foldedCommitment curve.G1Affine
		pok, err := BatchProve([]ProvingKey{}, [][]fr.Element{}, hashes[0])
		assert.NoError(t, err)

		_, err = foldedCommitment.Fold([]curve.G1Affine{}, hashes[0], ecc.MultiExpConfig{NbTasks: 1})
		assert.NoError(t, err)
		assert.NoError(t, vk.Verify(foldedCommitment, pok))
	})

	run := func(values [][]fr.Element) func(t *testing.T) {
		return func(t *testing.T) {

			var foldedCommitment curve.G1Affine
			pok, err := BatchProve(pk[:len(values)], values, hashes[0])
			assert.NoError(t, err)

			_, err = foldedCommitment.Fold(commitments[:len(values)], hashes[0], ecc.MultiExpConfig{NbTasks: 1})
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
	vk.G, err = curve.RandomOnG2()
	assert.NoError(t, err)
	vk.GSigma, err = curve.RandomOnG2()
	assert.NoError(t, err)

	t.Run("ProvingKey -> Bytes -> ProvingKey must remain identical.", testutils.SerializationRoundTrip(&pk))
	t.Run("ProvingKey -> Bytes (raw) -> ProvingKey must remain identical.", testutils.SerializationRoundTripRaw(&pk))
	t.Run("VerifyingKey -> Bytes -> VerifyingKey must remain identical.", testutils.SerializationRoundTrip(&vk))
	t.Run("VerifyingKey -> Bytes (raw) -> ProvingKey must remain identical.", testutils.SerializationRoundTripRaw(&vk))
}

func TestSemiFoldProofs(t *testing.T) {
	const (
		commitmentLength = 5
		nbCommitments    = 5
	)
	g, err := curve.RandomOnG2()
	assert.NoError(t, err)

	basis := randomG1Slice(t, commitmentLength*nbCommitments)

	vk, pk := make([]VerifyingKey, nbCommitments), make([]ProvingKey, nbCommitments)
	for i := range pk {
		var pk0 []ProvingKey
		pk0, vk[i], err = Setup([][]curve.G1Affine{basis[i*commitmentLength : (i+1)*commitmentLength]}, WithG2Point(g))
		assert.NoError(t, err)
		pk[i] = pk0[0]
	}

	values := make([][]fr.Element, nbCommitments)
	for i := range values {
		values[i] = make([]fr.Element, commitmentLength)
		for j := range values[i] {
			_, err = values[i][j].SetRandom()
			assert.NoError(t, err)
		}
	}

	commitments := make([]curve.G1Affine, nbCommitments)
	proofs := make([]curve.G1Affine, nbCommitments)
	for i := range commitments {
		commitments[i], err = pk[i].Commit(values[i])
		assert.NoError(t, err)
		proofs[i], err = pk[i].ProveKnowledge(values[i])
		assert.NoError(t, err)
	}

	var challenge fr.Element
	_, err = challenge.SetRandom()
	assert.NoError(t, err)

	assert.NoError(t, BatchVerifyMultiVk(vk, commitments, proofs, challenge))

	// send folded proof
	proof, err := new(curve.G1Affine).Fold(proofs, challenge, ecc.MultiExpConfig{NbTasks: 1})
	assert.NoError(t, err)
	assert.NoError(t, BatchVerifyMultiVk(vk, commitments, []curve.G1Affine{*proof}, challenge))
}
