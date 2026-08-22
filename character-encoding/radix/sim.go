package radix

import (
	"fmt"
	"math/big"
)

func SimAdd(spec Spec, x, y *big.Int) (*big.Int, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	mod := spec.Modulus()
	out := new(big.Int).Add(x, y)
	return out.Mod(out, mod), nil
}

func SimSub(spec Spec, x, y *big.Int) (*big.Int, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	mod := spec.Modulus()
	out := new(big.Int).Sub(x, y)
	return out.Mod(out, mod), nil
}

func SimEq(spec Spec, x, y *big.Int) (int, error) {
	if err := spec.Validate(); err != nil {
		return 0, err
	}
	mod := spec.Modulus()
	xr := new(big.Int).Mod(x, mod)
	yr := new(big.Int).Mod(y, mod)
	if xr.Cmp(yr) == 0 {
		return 1, nil
	}
	return 0, nil
}

func SimCmp(spec Spec, x, y *big.Int) ([3]int, error) {
	if err := spec.Validate(); err != nil {
		return [3]int{}, err
	}
	mod := spec.Modulus()
	xr := new(big.Int).Mod(x, mod)
	yr := new(big.Int).Mod(y, mod)
	switch xr.Cmp(yr) {
	case 0:
		return [3]int{1, 0, 0}, nil
	case -1:
		return [3]int{0, 1, 0}, nil
	case 1:
		return [3]int{0, 0, 1}, nil
	}
	return [3]int{}, fmt.Errorf("radix.SimCmp: unreachable")
}

func SimLt(spec Spec, x, y *big.Int) (int, error) {
	res, err := SimCmp(spec, x, y)
	if err != nil {
		return 0, err
	}
	return res[1], nil
}
