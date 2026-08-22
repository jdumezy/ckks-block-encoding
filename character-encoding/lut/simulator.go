package lut

import (
	"fmt"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charenc"
)

// Simulator is the plaintext oracle used to validate ciphertext LUTs.
type Simulator struct{}

func NewSimulator() *Simulator { return &Simulator{} }

func (s *Simulator) EvalUnary(in charenc.PlainBlock, table UnaryTable) (charenc.PlainBlock, error) {
	if in.Spec != table.In {
		return charenc.PlainBlock{}, fmt.Errorf("lut.Simulator.EvalUnary: input spec %+v does not match table.In %+v", in.Spec, table.In)
	}
	inCodec, err := charenc.NewCodec(table.In)
	if err != nil {
		return charenc.PlainBlock{}, err
	}
	outCodec, err := charenc.NewCodec(table.Out)
	if err != nil {
		return charenc.PlainBlock{}, err
	}
	tr, err := blt.CompileUnary(inCodec, outCodec, table.Eval)
	if err != nil {
		return charenc.PlainBlock{}, err
	}
	return blt.Apply(in, tr)
}

func (s *Simulator) EvalMulti(inputs []charenc.PlainBlock, table MultiTable) (charenc.PlainBlock, error) {
	if len(inputs) != len(table.In) {
		return charenc.PlainBlock{}, fmt.Errorf("lut.Simulator.EvalMulti: got %d inputs, expected %d", len(inputs), len(table.In))
	}
	codecs := make([]charenc.Codec, len(inputs))
	values := make([]int, len(inputs))
	for i, in := range inputs {
		if in.Spec != table.In[i] {
			return charenc.PlainBlock{}, fmt.Errorf("lut.Simulator.EvalMulti: input %d spec mismatch", i)
		}
		c, err := charenc.NewCodec(in.Spec)
		if err != nil {
			return charenc.PlainBlock{}, err
		}
		codecs[i] = c
		v, err := c.DecodeValue(in.Values)
		if err != nil {
			return charenc.PlainBlock{}, fmt.Errorf("lut.Simulator.EvalMulti: decode input %d: %w", i, err)
		}
		values[i] = v
	}
	out, err := charenc.NewCodec(table.Out)
	if err != nil {
		return charenc.PlainBlock{}, err
	}
	res := out.EncodeValue(table.Eval(values))
	return charenc.PlainBlock{Spec: table.Out, Values: res}, nil
}
