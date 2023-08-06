package rp

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToObjectID() *Stage {
	return &Stage{

		P: func() string {
			return "  => .(ObjectID) =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {
			return primitive.ObjectIDFromHex(in.(string))
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: BR,
				Obj:  H{"error": "Invalid: " + err.Error()},
			}
		},
	}
}

// ToTime - Converts in to time.Time for the UTC timezone. in must be a string matching the given layout.
// It is equivalent to calling ToTimeInLocation with ctxTimezoneName = "".
func ToTime(layout string) *Stage {
	return &Stage{

		P: func() string {
			return "  => .(time.Time) =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			timeString, ok := in.(string)
			if !ok {
				return nil, errors.New("not a string")
			}

			return time.Parse(layout, timeString)
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: BR,
				Obj:  H{"error": "Invalid: " + err.Error()},
			}
		},
	}
}

// ToTimeInLocation - Converts in to time.Time. in must be a string. It will be parsed per the given timezone and layout.
// If ctxTimezoneName is "" or its value has not been set in the context, UTC will be used.
func ToTimeInLocation(ctxTimezoneName string, layout string) *Stage {
	return &Stage{

		P: func() string {
			return "  => .(time.Time) =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			timeString, ok := in.(string)
			if !ok {
				return nil, errors.New("not a string")
			}

			timezoneName := ""
			if tz, ok := c.Get(ctxTimezoneName); ok {
				timezoneName = tz.(string)
			}

			locationTimezone, err := time.LoadLocation(timezoneName)
			if err != nil {
				// Default to UTC
				locationTimezone, _ = time.LoadLocation("UTC")
			}

			return time.ParseInLocation(layout, timeString, locationTimezone)
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: BR,
				Obj:  H{"error": "Invalid: " + err.Error()},
			}
		},
	}
}

func FieldValue(key string) *Stage {
	return &Stage{

		P: func() string {
			return "  => Value(\"" + key + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			if m, ok := in.(map[string]any); ok {
				return m[key], nil
			}
			return nil, nil
		},
	}
}
