package crt

import (
	"fmt"
	"math/big"

	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

type Plaintext struct {
	Spec   Spec
	Blocks []charenc.PlainBlock
}

type Ciphertext struct {
	Spec   Spec
	Blocks []charenc.CipherBlock
}

func Encode(s Spec, x *big.Int) (Plaintext, error) {
	if x == nil {
		return Plaintext{}, fmt.Errorf("crt.Encode: nil input")
	}
	blocks := make([]charenc.PlainBlock, s.Channels())
	for i, p := range s.Primes {
		codec, err := codecForPrime(s.Kind, p, s.Reduced)
		if err != nil {
			return Plaintext{}, err
		}
		residue := new(big.Int).Mod(x, big.NewInt(int64(p))).Int64()
		blocks[i] = charenc.PlainBlock{
			Spec:   codec.Spec(),
			Layout: charenc.Layout{Kind: charenc.BlockInOneCiphertext, BlockSlots: codec.Spec().Slots, Stride: 1},
			Values: codec.EncodeValue(int(residue)),
		}
	}
	return Plaintext{Spec: s, Blocks: blocks}, nil
}

func Encrypt(ctx *charctx.Context, pt Plaintext, level int) (Ciphertext, error) {
	blocks := make([]charenc.CipherBlock, len(pt.Blocks))
	for i, block := range pt.Blocks {
		cb, err := ctx.EncryptBlock(block, level)
		if err != nil {
			return Ciphertext{}, fmt.Errorf("crt.Encrypt: channel %d: %w", i, err)
		}
		blocks[i] = cb
	}
	return Ciphertext{Spec: pt.Spec, Blocks: blocks}, nil
}

func DecodeResidues(pt Plaintext) ([]int, error) {
	out := make([]int, pt.Spec.Channels())
	for i, block := range pt.Blocks {
		codec, err := codecForPrime(pt.Spec.Kind, pt.Spec.Primes[i], pt.Spec.Reduced)
		if err != nil {
			return nil, err
		}
		v, err := codec.DecodeValue(block.Values)
		if err != nil {
			return nil, fmt.Errorf("crt.DecodeResidues: channel %d: %w", i, err)
		}
		out[i] = v
	}
	return out, nil
}
