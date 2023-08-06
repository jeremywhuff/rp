package rp

import "github.com/gin-gonic/gin"

type Route struct {
	HttpMethod   string
	RelativePath string
	Pipe         *Chain
	Logger       Logger
}

func AddRoute(engine *gin.Engine, route *Route) {
	engine.Handle(route.HttpMethod, route.RelativePath, route.Handler())
}

func (r *Route) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		r.Run(c)
	}
}

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
