package rp

import "github.com/gin-gonic/gin"

type ChainExecutionError struct {
	StageError *StageError
}

func (e ChainExecutionError) Error() string {
	return "chain execution error"
}

func If(cond func(any, *gin.Context) bool, then *Chain, els *Chain) *Stage {
	return &Stage{

		P: func() string {
			if then != nil && els == nil {
				return "If => then"
			}
			return "If => then/else"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			var ch *Chain
			if cond(in, c) {
				ch = then
			} else {
				ch = els
			}

			if ch == nil {
				return nil, nil
			}

			o, e := Execute(ch, c, lgr)
			if e != nil {
				return nil, ChainExecutionError{StageError: e}
			}
			return o, nil
		},

		E: func(err error) *StageError {
			return err.(ChainExecutionError).StageError
		},
	}
}
