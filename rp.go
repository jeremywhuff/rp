package rp

// rp stands for "request pipeline"

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Route struct {
	// TODO: Path, method
	Pipe *Chain
}

func (r *Route) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		r.Run(c)
	}
}

// TODO: Remove this
func StartPipe() *Stage {
	return &Stage{
		Name: "Start",
		F: func(in any, c *gin.Context) (any, error) {
			return in, nil
		},
	}
}

type H map[string]any

//

var (
	BR  = http.StatusBadRequest
	ISR = http.StatusInternalServerError
)

func S(name string, f func(any, *gin.Context) (any, error), errorCode int, errorPrefix string) *Stage {

	prefix := ""
	if errorPrefix != "" {
		prefix = errorPrefix + ": "
	}

	return &Stage{
		Name: name,
		F:    f,
		E: func(err error) *StageError {
			return &StageError{
				Code: errorCode,
				Obj:  H{"error": prefix + err.Error()},
			}
		},
	}
}

type pipeResult struct {
	Out   any
	Error *StageError
}

func runInParallel(ch *Chain, c *gin.Context, r chan pipeResult) {
	o, e := ch.Run(c)
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

		Name: "InParallel",

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

//

func MongoFetch(ctxDatabaseName string, collectionName string, projection map[string]any) *Stage {
	return &Stage{

		Name: "  => MongoFetch(\"" + collectionName + "\") =>",

		F: func(in any, c *gin.Context) (any, error) {

			db := c.MustGet(ctxDatabaseName).(*mongo.Database)
			coll := db.Collection(collectionName)

			pipeline := []H{{
				"$match": H{
					"_id": in.(primitive.ObjectID)}}, {
				"$project": projection},
			}

			results := make([]map[string]any, 0)
			cur, err := coll.Aggregate(context.Background(), pipeline)
			if err != nil {
				return nil, err
			}
			defer cur.Close(context.Background())

			if err = cur.All(context.Background(), &results); err != nil {
				return nil, err
			}

			if len(results) > 0 {
				return results[0], nil
			}
			return nil, mongo.ErrNoDocuments
		},

		E: func(err error) *StageError {
			if err == mongo.ErrNoDocuments {
				return &StageError{
					Code: http.StatusNotFound,
					Obj:  H{"error": "MongoFetch: Document not found"},
				}
			}
			return &StageError{
				Code: ISR,
				Obj:  H{"error": err.Error()},
			}
		},
	}
}

func FieldValue(key string) *Stage {
	return &Stage{

		Name: "  => Value(\"" + key + "\") =>",

		F: func(in any, c *gin.Context) (any, error) {
			return in.(map[string]any)[key], nil
		},
	}
}

func MongoPipe(ctxDatabaseName string, collectionName string) *Stage {
	return &Stage{

		Name: "  => MongoPipe(\"" + collectionName + "\") =>",

		F: func(in any, c *gin.Context) (any, error) {

			db := c.MustGet(ctxDatabaseName).(*mongo.Database)
			coll := db.Collection(collectionName)

			results := make([]map[string]any, 0)
			cur, err := coll.Aggregate(context.Background(), in.([]H))
			if err != nil {
				return nil, err
			}
			defer cur.Close(context.Background())

			if err = cur.All(context.Background(), &results); err != nil {
				return nil, err
			}

			if len(results) > 0 {
				return results[0], nil
			}
			return nil, mongo.ErrNoDocuments
		},

		E: func(err error) *StageError {
			if err == mongo.ErrNoDocuments {
				return &StageError{
					Code: http.StatusNotFound,
					Obj:  H{"error": "MongoFetch: Document not found"},
				}
			}
			return &StageError{
				Code: ISR,
				Obj:  H{"error": err.Error()},
			}
		},
	}
}

var ErrNotFound = errors.New("not found")

// MongoInsert inserts in as a document
func MongoInsert(ctxDatabaseName string, collectionName string) *Stage {
	return &Stage{

		Name: "  => MongoInsert(\"" + collectionName + "\") =>",

		F: func(in any, c *gin.Context) (any, error) {

			db := c.MustGet(ctxDatabaseName).(*mongo.Database)
			coll := db.Collection(collectionName)

			insertResult, err := coll.InsertOne(c, in)
			if err != nil {
				return nil, err
			}

			out := insertResult.InsertedID.(primitive.ObjectID)

			return out, nil
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: ISR,
				Obj:  H{"error": err.Error()},
			}
		},
	}
}

func CtxGet(key string) *Stage {
	return &Stage{

		Name: "[\"" + key + "\"] =>",

		F: func(in any, c *gin.Context) (any, error) {
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

		Name: "  => [\"" + key + "\"]",

		F: func(in any, c *gin.Context) (any, error) {
			c.Set(key, in)
			return in, nil
		},
	}
}

func Bind(obj any) *Stage {
	return &Stage{

		Name: "Req.Body =>",

		F: func(in any, c *gin.Context) (any, error) {
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

		Name: "Req.URL(\"" + key + "\") =>",

		F: func(in any, c *gin.Context) (any, error) {
			return c.Param(key), nil
		},
	}
}

func ToObjectID() *Stage {
	return &Stage{

		Name: "  => .(ObjectID) =>",

		F: func(in any, c *gin.Context) (any, error) {
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

func (s *Stage) CatchPrefix(errorPrefix string) *Stage {

	if errorPrefix == "" || s.E == nil {
		return s
	}

	s.E = func(err error) *StageError {
		stageError := s.E(err)
		stageError.Obj.(H)["error"] = errorPrefix + ": " + stageError.Obj.(H)["error"].(string)
		return stageError
	}

	return s
}
