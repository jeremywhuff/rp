package rpout

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/jeremywhuff/rp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoFindOneOptions struct {
	// If non-nil, the result will be unmarshalled into this object. Default is nil.
	// If nil, the result will be unmarshalled into an object of type map[string]any.
	// It is sent to the mongo.SingleResult.Decode() method.
	Result any
}

func MongoFindOne(ctxDatabaseName string, collectionName string, opts ...MongoFindOneOptions) *Stage {
	return &Stage{

		P: func() string {
			return `  => MongoFindOne("` + collectionName + `") =>`
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			db := c.MustGet(ctxDatabaseName).(*mongo.Database)
			coll := db.Collection(collectionName)

			var result any
			if len(opts) > 0 && opts[0].Result != nil {
				result = opts[0].Result
			} else {
				result = map[string]any{}
			}

			err := coll.FindOne(context.Background(), in).Decode(&result)
			if err != nil {
				return nil, err
			}

			return &result, nil
		},

		E: func(err error) *StageError {
			if err == mongo.ErrNoDocuments {
				return &StageError{
					Code: http.StatusNotFound,
					Obj:  H{"error": "MongoFindOne: Document not found"},
				}
			}
			return &StageError{
				Code: ISR,
				Obj:  H{"error": "MongoFindOne: " + err.Error()},
			}
		},
	}
}

func MongoFetch(ctxDatabaseName string, collectionName string, projection map[string]any) *Stage {
	return &Stage{

		P: func() string {
			return "  => MongoFetch(\"" + collectionName + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

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

type MongoPipeOptions struct {
	// If non-nil, the results will be unmarshalled into this object. Default is nil.
	// It must be a pointer to a slice. It is sent to the mongo.Cursor.All() method.
	Results any
}

// MongoPipe executes the given pipeline of the in parameter on the given collection.
// The *mongo.Database instance must be set in the context with the given ctxDatabaseName as the key.
// in must be a valid pipeline for the mongo.Collection.Aggregate() method.
func MongoPipe(ctxDatabaseName string, collectionName string, opts *MongoPipeOptions) *Stage {
	return &Stage{

		P: func() string {
			return "  => MongoPipe(\"" + collectionName + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

			db, ok := c.MustGet(ctxDatabaseName).(*mongo.Database)
			if !ok {
				return nil, errors.New("mongo database not found in context")
			}
			coll := db.Collection(collectionName)

			cur, err := coll.Aggregate(context.Background(), in)
			if err != nil {
				return nil, err
			}
			defer cur.Close(context.Background())

			var results any
			if opts != nil && opts.Results != nil {
				results = opts.Results
			} else {
				results = make([]map[string]any, 0)
			}

			if err = cur.All(context.Background(), &results); err != nil {
				return nil, err
			}

			return results, nil
		},

		E: func(err error) *StageError {
			return &StageError{
				Code: ISR,
				Obj:  H{"error": err.Error()},
			}
		},
	}
}

// MongoInsert inserts in as a document
func MongoInsert(ctxDatabaseName string, collectionName string) *Stage {
	return &Stage{

		P: func() string {
			return "  => MongoInsert(\"" + collectionName + "\") =>"
		},

		F: func(in any, c *gin.Context, lgr Logger) (any, error) {

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
