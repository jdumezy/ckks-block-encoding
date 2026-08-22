package radix

import (
	"fmt"
	"math/big"
)

const (
	DefaultBitsPerDigit = 4
	DefaultRadix        = 1 << DefaultBitsPerDigit
)

type Spec struct {
	Radix int
	Width int
}

func NewSpec(wordBits, bitsPerDigit int) (Spec, error) {
	if wordBits <= 0 {
		return Spec{}, fmt.Errorf("radix.NewSpec: wordBits must be positive")
	}
	if bitsPerDigit <= 0 {
		return Spec{}, fmt.Errorf("radix.NewSpec: bitsPerDigit must be positive")
	}
	if wordBits%bitsPerDigit != 0 {
		return Spec{}, fmt.Errorf("radix.NewSpec: wordBits=%d is not divisible by bitsPerDigit=%d", wordBits, bitsPerDigit)
	}
	return Spec{
		Radix: 1 << bitsPerDigit,
		Width: wordBits / bitsPerDigit,
	}, nil
}

func New64BitSpec() Spec {
	return Spec{Radix: DefaultRadix, Width: 64 / DefaultBitsPerDigit}
}

func (s Spec) Validate() error {
	if s.Radix < 2 {
		return fmt.Errorf("radix.Spec: radix must be at least 2")
	}
	if s.Width <= 0 {
		return fmt.Errorf("radix.Spec: width must be positive")
	}
	return nil
}

func (s Spec) Modulus() *big.Int {
	out := big.NewInt(1)
	base := big.NewInt(int64(s.Radix))
	for i := 0; i < s.Width; i++ {
		out.Mul(out, base)
	}
	return out
}

type Word struct {
	Spec   Spec
	Digits []int
}

func Encode(s Spec, x *big.Int) (Word, error) {
	if err := s.Validate(); err != nil {
		return Word{}, err
	}
	if x == nil {
		return Word{}, fmt.Errorf("radix.Encode: nil input")
	}
	mod := s.Modulus()
	cur := new(big.Int).Mod(new(big.Int).Set(x), mod)
	base := big.NewInt(int64(s.Radix))
	digits := make([]int, s.Width)
	for i := range digits {
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(cur, base, r)
		digits[i] = int(r.Int64())
		cur = q
	}
	return Word{Spec: s, Digits: digits}, nil
}

func Decode(w Word) (*big.Int, error) {
	if err := w.check(); err != nil {
		return nil, err
	}
	out := big.NewInt(0)
	base := big.NewInt(int64(w.Spec.Radix))
	for i := len(w.Digits) - 1; i >= 0; i-- {
		out.Mul(out, base)
		out.Add(out, big.NewInt(int64(w.Digits[i])))
	}
	return out, nil
}

func (w Word) check() error {
	if err := w.Spec.Validate(); err != nil {
		return err
	}
	if len(w.Digits) != w.Spec.Width {
		return fmt.Errorf("radix.Word: got %d digits, expected %d", len(w.Digits), w.Spec.Width)
	}
	for i, d := range w.Digits {
		if d < 0 || d >= w.Spec.Radix {
			return fmt.Errorf("radix.Word: digit %d=%d outside [0,%d)", i, d, w.Spec.Radix)
		}
	}
	return nil
}
