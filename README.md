# rp

rp, the “request pipeline” framework, makes server endpoints with multiple execution steps easier to build, maintain, and optimize. It is built with [Gin](https://github.com/gin-gonic/gin), Go's [top web framework](https://github.com/EvanLi/Github-Ranking/blob/master/Top100/Go.md).

It works by wrapping execution steps of any arbitrary code into stages that can be linked together into execution chains. Chains, in turn, can be executed, in sequence or in parallel (concurrently), with a logger that automatically tracks each stage’s success or failure along with performance metrics like latency.

Check out [the tutorial](https://medium.com/@jeremywhuff/950a10c3c31f) for more.

### UNDER DEVELOPMENT - I expect to have a first-release version shortly. This README will be updated.
