package rp

import (
	"github.com/gin-gonic/gin"
)

// Stage is a step in a request pipeline. Stages are connected together as double-linked lists by n and l.
// When a pipeline is run, it executes each Stage's F function. The input to F is the output of the last Stage
// plus the request's context. The output of F is, in turn, passed into the next Stage.
// When F returns an error, it is passed in to the E function, which generates the HTTP status code and JSON
// response data that should be returned in the network response.
// The last Stage of a pipeline should return a *Response as the output of F.
// When a stage completes, P() will be logged to the console with the results of the stage.
type Stage struct {
	P func() string                                // Printed name of the stage, for logging
	F func(any, *gin.Context, Logger) (any, error) // Function to execute. Optional logger for stages that nest chains.
	E func(error) *StageError                      // Network error to return for F's error
	n *Stage                                       // Next stage
	l *Stage                                       // Last stage
}

func (s *Stage) Chain() *Chain {
	ch := Chain{
		First: s,
		Last:  s,
	}
	return &ch
}

type Chain struct {
	First *Stage
	Last  *Stage
}

// Pipelines should be defined by sending the first Stage in to the First function and then each following
// Stage into the Then function. The pipeline definition should read like:
//
//	pipeline := First(stage0).Then(stage1).Then(stage2) ...
//
// or alternatively:
//
//	pipeline := First(
//	    stage0).Then(
//	    stage1).Then(
//	    stage2) ...
func First(s *Stage) *Chain {
	ch := Chain{
		First: s,
		Last:  s,
	}
	return &ch
}
func (ch *Chain) Then(n *Stage) *Chain {
	ch.Last.n = n
	n.l = ch.Last
	ch.Last = n
	return ch
}

// Alternatively, MakeChain can build a chain from a slice of stages.
func MakeChain(stages ...*Stage) *Chain {
	if len(stages) == 0 {
		return nil
	}
	ch := First(stages[0])
	for _, s := range stages[1:] {
		ch.Then(s)
	}
	return ch
}

// Catch can be used to optionally override a stage's E function like:
//
//	pipeline := First(
//	    stage0).Then(
//	    stage1).Catch(http.StatusBadRequest, "stage1 failed").Then(
//	    stage2) ...
func (ch *Chain) Catch(Code int, Message string) *Chain {
	ch.Last.E = func(err error) *StageError {
		return &StageError{
			Code: Code,
			Obj:  H{"error": Message},
		}
	}
	return ch
}

func (ch *Chain) CatchError(E func(err error) *StageError) *Chain {
	ch.Last.E = E
	return ch
}

// InSequence concatenates together multiple chains defined by the above First+Then method.
func InSequence(chains ...*Chain) *Chain {

	if len(chains) == 0 {
		return nil
	}

	// Start with the first chain as the base
	ch := chains[0]

	for i := range chains {

		// Break if there are no more chains to link
		if i == len(chains)-1 {
			break
		}

		// Link ch's last stage to the next chain's first stage
		ch.Last.n = chains[i+1].First
		chains[i+1].First.l = ch.Last

		// Include all of the next chain's stages into ch
		ch.Last = chains[i+1].Last
	}

	return ch
}
