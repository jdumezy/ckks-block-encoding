package crt

import (
	"fmt"
	"math"
	"math/big"

	"character-encoding/character-encoding/charenc"
)

type Spec struct {
	Bits    int
	Kind    charenc.EncodingKind
	Primes  []int
	Reduced bool
	Specs   []charenc.BlockSpec
	Log2T   float64
}

func NewSpecForBits(bits int, kind charenc.EncodingKind, maxPrime int) (Spec, error) {
	if bits <= 0 {
		return Spec{}, fmt.Errorf("crt.NewSpecForBits: bits must be positive")
	}
	if maxPrime == 0 {
		maxPrime = math.MaxInt
	}
	if maxPrime < 2 {
		return Spec{}, fmt.Errorf("crt.NewSpecForBits: maxPrime must be at least 2")
	}
	primes := make([]int, 0, bits/8+1)
	log2T := 0.0
	for p := 2; log2T < float64(bits); p++ {
		if p > maxPrime {
			return Spec{}, fmt.Errorf("crt.NewSpecForBits: product reached %.2f bits below target %d before maxPrime=%d", log2T, bits, maxPrime)
		}
		if !isPrime(p) {
			continue
		}
		primes = append(primes, p)
		log2T += math.Log2(float64(p))
	}
	return NewSpec(primes, kind, true, bits, log2T)
}

func NewSpec(primes []int, kind charenc.EncodingKind, reduced bool, bits int, log2T float64) (Spec, error) {
	if len(primes) == 0 {
		return Spec{}, fmt.Errorf("crt.NewSpec: empty prime list")
	}
	specs := make([]charenc.BlockSpec, len(primes))
	seen := make(map[int]struct{}, len(primes))
	if log2T == 0 {
		for _, p := range primes {
			log2T += math.Log2(float64(p))
		}
	}
	for i, p := range primes {
		if !isPrime(p) {
			return Spec{}, fmt.Errorf("crt.NewSpec: modulus %d is not prime", p)
		}
		if _, ok := seen[p]; ok {
			return Spec{}, fmt.Errorf("crt.NewSpec: repeated prime %d", p)
		}
		seen[p] = struct{}{}
		codec, err := codecForPrime(kind, p, reduced)
		if err != nil {
			return Spec{}, fmt.Errorf("crt.NewSpec: channel %d: %w", i, err)
		}
		specs[i] = codec.Spec()
	}
	return Spec{
		Bits:    bits,
		Kind:    kind,
		Primes:  append([]int(nil), primes...),
		Reduced: reduced,
		Specs:   specs,
		Log2T:   log2T,
	}, nil
}

func (s Spec) Channels() int {
	return len(s.Primes)
}

func (s Spec) MaxPrime() int {
	maxP := 0
	for _, p := range s.Primes {
		if p > maxP {
			maxP = p
		}
	}
	return maxP
}

func (s Spec) Modulus() *big.Int {
	out := big.NewInt(1)
	for _, p := range s.Primes {
		out.Mul(out, big.NewInt(int64(p)))
	}
	return out
}

func codecForPrime(kind charenc.EncodingKind, p int, reduced bool) (charenc.Codec, error) {
	switch kind {
	case charenc.BRU:
		return charenc.NewBRU(p, reduced)
	case charenc.LBRU:
		return charenc.NewLBRU(p, 0, reduced)
	case charenc.WH:
		return nil, fmt.Errorf("WH is not a prime-field CRT channel encoding")
	case charenc.IND:
		return charenc.NewIND(p, reduced, 0)
	default:
		return nil, fmt.Errorf("unsupported encoding kind %v", kind)
	}
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	for d := 3; d*d <= n; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}
