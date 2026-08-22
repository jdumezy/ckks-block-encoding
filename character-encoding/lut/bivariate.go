package lut

import (
	"fmt"

	"character-encoding/character-encoding/charctx"
)

func CompileBinary(table MultiTable, ctx *charctx.Context, inputLevel int) (*CompiledTensor, error) {
	if len(table.In) != 2 {
		return nil, fmt.Errorf("lut.CompileBinary: arity %d, expected 2", len(table.In))
	}
	return CompileMultivariate(table, ctx, inputLevel)
}
