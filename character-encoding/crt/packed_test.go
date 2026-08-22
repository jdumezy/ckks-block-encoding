package crt

import (
	"math/big"
	"testing"

	"character-encoding/character-encoding/charenc"
)

func TestPackedSpecMaximalWordBatching(t *testing.T) {
	for _, tc := range []struct {
		bits       int
		wordSlots  int
		batchSize  int
		totalSlots int
	}{
		{bits: 64, wordSlots: 365, batchSize: 44, totalSlots: 16060},
		{bits: 256, wordSlots: 3787, batchSize: 4, totalSlots: 15148},
	} {
		base, err := NewSpecForBits(tc.bits, charenc.BRU, 0)
		if err != nil {
			t.Fatalf("NewSpecForBits(%d): %v", tc.bits, err)
		}
		spec, err := NewPackedSpec(base, 1<<14)
		if err != nil {
			t.Fatalf("NewPackedSpec(%d): %v", tc.bits, err)
		}
		if spec.Ciphertexts() != 1 {
			t.Fatalf("%d-bit Ciphertexts()=%d, want 1", tc.bits, spec.Ciphertexts())
		}
		if spec.WordStride != tc.wordSlots {
			t.Fatalf("%d-bit WordStride=%d, want %d", tc.bits, spec.WordStride, tc.wordSlots)
		}
		if spec.BatchSize != tc.batchSize {
			t.Fatalf("%d-bit BatchSize=%d, want %d", tc.bits, spec.BatchSize, tc.batchSize)
		}
		if spec.TotalUsed != tc.totalSlots {
			t.Fatalf("%d-bit TotalUsed=%d, want %d", tc.bits, spec.TotalUsed, tc.totalSlots)
		}
	}
}

func TestPackedBatchEncodeDecodeResidues(t *testing.T) {
	base, err := NewSpecForBits(64, charenc.LBRU, 0)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := NewPackedSpec(base, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	xs := make([]*big.Int, spec.BatchSize)
	for i := range xs {
		xs[i] = big.NewInt(int64(1000 + 17*i))
	}
	pt, err := EncodePackedBatch(spec, xs)
	if err != nil {
		t.Fatal(err)
	}
	residues, err := DecodePackedBatchResidues(pt)
	if err != nil {
		t.Fatal(err)
	}
	if len(residues) != spec.BatchSize {
		t.Fatalf("decoded %d words, want %d", len(residues), spec.BatchSize)
	}
	for w := range residues {
		for ch, p := range spec.Base.Primes {
			want := int(new(big.Int).Mod(xs[w], big.NewInt(int64(p))).Int64())
			if residues[w][ch] != want {
				t.Fatalf("word %d channel %d residue=%d, want %d", w, ch, residues[w][ch], want)
			}
		}
	}
}

func TestSwitchPackedSpecPreservesWordLayout(t *testing.T) {
	base, err := NewSpecForBits(256, charenc.BRU, 0)
	if err != nil {
		t.Fatal(err)
	}
	bru, err := NewPackedSpec(base, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	lbru, err := SwitchPackedSpec(bru, charenc.LBRU)
	if err != nil {
		t.Fatal(err)
	}
	if !samePackedLayout(bru, lbru) {
		t.Fatalf("BRU -> LBRU changed packed word layout")
	}
}
