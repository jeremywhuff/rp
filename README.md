# rp
rp: "Request Pipeline". A Go language Gin-based framework for building server endpoints more easily via chains of well-defined, modular execution stages.

UNDER DEVELOPMENT - I expect to have a stable version in a couple of weeks. This README will be updated as progress is made.

## The Basics

### Why does rp exist?

It started because I recognized a few problems had dogged me for years in our Gin-based backend, and I had an idea for a lightweight solution that could knock those all down.

Our startup, Purple Go, has been running continually since the end of 2016, and it's extraordinarily rare for it to go down thanks to Gin's bulletproof stability. I've also been quite happy with MongoDB as the app's database. I have directly maintained the Go codebases myself for the past ~5 years, so I have spent countless hours building out new features, fixing bugs, and thinking deep on architecture stack and code structure. I have had a number of thorough cleanup cycles where I did iterative architecture overhauls to make sure it's always getting better.

Here are the problems that I kept wrestling with, which rp is designed to solve.

1. Lack of Readability / Transparency High Level Server Route Code.

   I want to be able to skim the entry point of a server route's handler code and immediately know what the required request input struct is, what it's response structures are, what application logic it follows during execution, and whether it takes any actions external services, e.g. sending and email or running a credit card charge. Breaking certain functionality out into internal packages seemed like a clean structure initially, but it had the byproduct of burying these code details into functions call stacks, and it was tough to handle just with documentation and smart function naming.

(More to come here. The images below will be explained soon.)

<img align="left" width="588px" src="docs/rpFrameworkDemoCode1.png">
<img align="left" width="673px" src="docs/rpFrameworkDemoCode2.png">
<img align="left" width="711px" src="docs/rpFrameworkDemoCode3.png">
<img align="left" width="766px" src="docs/rpFrameworkDemoCode4.png">
