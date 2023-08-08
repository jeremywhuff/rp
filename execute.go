package rp

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

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
	LogMessage(msg string)
	LogStageStart(print string, in any)
	LogStageComplete(success bool, elapsed time.Duration, print string, out any)
	LogStageError(e *StageError)
}

type DefaultLogger struct {
	Logger
}

func (l DefaultLogger) LogMessage(msg string) {
	log.Print(msg)
}

func (l DefaultLogger) LogStageStart(print string, in any) {
	// Ignore
}

func (l DefaultLogger) LogStageComplete(success bool, elapsed time.Duration, print string, out any) {

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

func (l DefaultLogger) LogStageError(e *StageError) {
	log.Printf("")
	log.Printf("Error: %s", e.Obj.(H)["error"])
	log.Printf("")
}

func Execute(ch *Chain, c *gin.Context, lgr Logger) (any, *StageError) {

	if lgr != nil {
		lgr.LogMessage("Starting execution chain...")
	}

	s := ch.First
	var d any // Data passed between successive stages
	var e *StageError

	// Execute all stages
	for s != nil {

		if lgr != nil {
			lgr.LogStageStart(s.P(), d)
		}

		t := time.Now()

		d, e = s.Execute(d, c, lgr)

		if lgr != nil {
			lgr.LogStageComplete(e == nil, time.Since(t), s.P(), d)
			if e != nil {
				lgr.LogStageError(e)
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
func (s *Stage) Execute(in any, c *gin.Context, lgr Logger) (any, *StageError) {

	out, err := s.F(in, c, lgr)
	if err != nil {
		return nil, s.E(err)
	}

	return out, nil
}

func MakeGinHandlerFunc(ch *Chain, lgr Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := Execute(ch, c, lgr)
		if e != nil {
			c.JSON(e.Code, e.Obj)
			return
		}

		res := o.(*Response)
		c.JSON(res.Code, res.Obj)
	}
}
