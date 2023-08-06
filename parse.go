package rp

import "github.com/gin-gonic/gin"

func Bind(obj any) *Stage {
	return &Stage{

		P: func() string {
			return "Req.Body =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			err := c.ShouldBindJSON(obj)
			if err != nil {
				return nil, err
			}
			return obj, nil
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: BR,
				Obj:  H{"error": "Invalid request: " + err.Error()},
			}
		},
	}
}

func URLParam(key string) *Stage {
	return &Stage{

		P: func() string {
			return "Req.URL(\"" + key + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			return c.Param(key), nil
		},
	}
}

func QueryParam(key string) *Stage {
	return &Stage{

		P: func() string {
			return "Req.Query(\"" + key + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			return c.Query(key), nil
		},
	}
}
