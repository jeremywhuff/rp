package rp

// rp stands for "request pipeline"

import (
	"net/http"
)

// Helpers

type H map[string]any

var (
	BR  = http.StatusBadRequest
	ISR = http.StatusInternalServerError
)

// Files within this package:
//
// | File           | Description                                                        |
// | -------------- | ------------------------------------------------------------------ |
// | CORE FUNCTIONALITY                                                                  |
// | rp.go          | This file; Some helpers and documentation                          |
// | route.go       | Route type, the top-level object that contains the pipeline        |
// | pipeline.go    | Stage & Chain types; Basic building blocks for defining pipelines  |
// | execute.go     | Execute func that runs pipelines; Logging via the Logger interface |
// | conditional.go | Stage that wraps chains into an if/else control flow               |
// | parallel.go    | Stage that runs multiple chains in parallel                        |
// | -------------- | ------------------------------------------------------------------ |
// | STAGE GENERATOR FUNCTIONS                                                           |
// | basic.go       | Generic stage generator, context get/set stages                    |
// | parse.go       | Request parsing stages                                             |
// | conversion.go  | Type conversion stages                                             |
// | -------------- | ------------------------------------------------------------------ |
// | INTEGRATIONS																		 |
// | rpout/mongo.go | Stages that use the MongoDB Go driver                              |
// |                | "go.mongodb.org/mongo-driver/mongo"								 |
