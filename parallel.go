package rp

import "github.com/gin-gonic/gin"

type pipeResult struct {
	Out   any
	Error *StageError
}

func runInParallel(ch *Chain, c *gin.Context, r chan pipeResult) {
	o, e := Execute(ch, c, nil)
	r <- pipeResult{
		Out:   o,
		Error: e,
	}
}

type parallelError struct {
	error
	StageError *StageError
}

func (e parallelError) Error() string {
	return "parallel error"
}

func InParallel(chains ...*Chain) *Chain {
	return First(&Stage{

		P: func() string {
			return "InParallel"
		},

		F: func(in any, c *gin.Context) (any, error) {

			resultChans := make([](chan pipeResult), len(chains))

			for i, ch := range chains {
				chn := make(chan pipeResult)
				defer close(chn)
				go runInParallel(ch, c, chn)
				resultChans[i] = chn
			}

			out := make([]any, len(chains))
			outErr := make([]*StageError, len(chains))

			for i, rc := range resultChans {
				r := <-rc
				out[i] = r.Out
				outErr[i] = r.Error
			}

			for _, e := range outErr {
				if e != nil {
					return nil, parallelError{StageError: e}
				}
			}

			return out, nil
		},

		E: func(err error) *StageError {
			return err.(parallelError).StageError
		},
	})
}
