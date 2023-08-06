package rp

// rp stands for "request pipeline"

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Helpers

type H map[string]any

var (
	BR  = http.StatusBadRequest
	ISR = http.StatusInternalServerError
)

// S creates a generic stage that executes the given function.
// E's default code is http.StatusBadRequest since that is common.
func S(name string, f func(any, *gin.Context, Logger) (any, error)) *Stage {

	return &Stage{
		P: func() string {
			return name
		},
		F: f,
		E: func(err error) *StageError {
			return &StageError{
				Code: BR,
				Obj:  H{"error": err.Error()},
			}
		},
	}
}

// CtxGet / CtxSet

var ErrNotFound = errors.New("not found")

func CtxGet(key string) *Stage {
	return &Stage{

		P: func() string {
			return "[\"" + key + "\"] =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			val, ok := c.Get(key)
			if !ok {
				return nil, ErrNotFound
			}
			return val, nil
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: ISR,
				Obj:  H{"error": "Key not found: " + key},
			}
		},
	}
}

func CtxSet(key string) *Stage {
	return &Stage{

		P: func() string {
			return "  => [\"" + key + "\"]"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			c.Set(key, in)
			return in, nil
		},
	}
}
