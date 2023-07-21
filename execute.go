package rp

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

var debug = true

// Run runs the route's Pipe and sets the network response based on the run results.
func (r *Route) Run(c *gin.Context) {

	o, e := r.Pipe.Run(c)
	if e != nil {
		c.JSON(e.Code, e.Obj)
		return
	}

	res := o.(*Response)
	c.JSON(res.Code, res.Obj)
}

// TODO: Replace the Response and StageError types with this single type
type JSONResponse struct {
	Code int // HTTP status code
	Obj  any // JSON response data
}

type Response struct {
	Code int // HTTP status code
	Obj  any // JSON response data
}

type StageError struct {
	Code int // HTTP status code
	Obj  any // JSON response data
}

// runPipeline runs all stages in the pipeline chain from the earliest to the latest.
// It returns the output of the last stage in the chain.
// As of now all calls to this function input an s that is the last stage in the chain.

func (ch *Chain) Run(c *gin.Context) (any, *StageError) {

	if debug {
		log.Print("Starting pipeline...")
	}

	s := ch.First
	var d any // Data passed between successive stages
	var e *StageError

	// Execute all stages
	for s != nil {

		t := time.Now()

		d, e = s.Execute(d, c)
		if e != nil {

			if debug {
				printResult(false, time.Since(t), s.Name)

				log.Printf("")
				log.Printf("Error: %s", e.Obj.(H)["error"])
				log.Printf("")
			}

			return nil, e
		}

		if debug {
			printResult(true, time.Since(t), s.Name)
		}

		s = s.n
	}

	return d, nil
}

// Execute executes the stage by calling the F function followed by the E function if there's an error.
func (s *Stage) Execute(in any, c *gin.Context) (any, *StageError) {

	out, err := s.F(in, c)
	if err != nil {
		return nil, s.E(err)
	}

	return out, nil
}

func printResult(success bool, elapsed time.Duration, name string) {
	lbl := color.New(color.FgWhite).Add(color.BgGreen).Sprintf(" OK  ")
	if !success {
		lbl = color.New(color.FgWhite).Add(color.BgRed).Sprintf(" ERR ")
	}

	tclr := color.New(color.FgWhite, color.Faint)
	if elapsed > time.Millisecond {
		tclr = color.New(color.FgWhite).Add(color.BgCyan)
	}

	time := tclr.Sprintf("%13v", elapsed)

	log.Print("|" + lbl + "| " + time + " | " + name)
}
