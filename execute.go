package rp

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// Run runs the route's Pipe and sets the network response based on the run results.
func (r *Route) Run(c *gin.Context) {

	o, e := Execute(r.Pipe, c, r.Logger)
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

type Logger interface {
	LogStart()
	LogStage(success bool, elapsed time.Duration, print string)
	LogError(e *StageError)
}

type DefaultLogger struct {
	Logger
}

func (l DefaultLogger) LogStart() {
	log.Print("Starting pipeline...")
}

func (l DefaultLogger) LogStage(success bool, elapsed time.Duration, print string) {

	// Column 1: Success or failure
	lbl := color.New(color.FgWhite).Add(color.BgGreen).Sprintf(" OK  ")
	if !success {
		lbl = color.New(color.FgWhite).Add(color.BgRed).Sprintf(" ERR ")
	}

	// Column 2: Time elapsed
	tclr := color.New(color.FgWhite, color.Faint)
	if elapsed > time.Millisecond {
		tclr = color.New(color.FgWhite).Add(color.BgCyan)
	}
	time := tclr.Sprintf("%13v", elapsed)

	// Column 3: Stage print

	log.Print("|" + lbl + "| " + time + " | " + print)
}

func (l DefaultLogger) LogError(e *StageError) {
	log.Printf("")
	log.Printf("Error: %s", e.Obj.(H)["error"])
	log.Printf("")
}

func Execute(ch *Chain, c *gin.Context, lgr Logger) (any, *StageError) {

	if lgr != nil {
		lgr.LogStart()
	}

	s := ch.First
	var d any // Data passed between successive stages
	var e *StageError

	// Execute all stages
	for s != nil {

		t := time.Now()

		d, e = s.Execute(d, c)

		if lgr != nil {
			lgr.LogStage(e == nil, time.Since(t), s.P())
			if e != nil {
				lgr.LogError(e)
			}
		}

		if e != nil {
			return nil, e
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
