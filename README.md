# rp
rp: "Request Pipeline". A Go language Gin-based framework for building server endpoints more easily via chains of well-defined, modular execution stages.

UNDER DEVELOPMENT - I expect to have a stable version in a couple of weeks. This README will be updated as progress is made.

## The Basics

### Why does rp exist?

It started because I recognized a few problems had dogged me for years in our Gin-based backend, and I had an idea for a lightweight solution that could knock those all down at once. This thinking dovetailed nicely into the "chaining" architectures used in AI, even though my concept for rp doesn't work directly with AI models much at all. I believe that this similarity will actually make it very complimentary to AI chain frameworks. I have borrowed some ideas, and rp's more unique features may volunteer some ideas back towards AI.

Our startup, Purple Go, has been running continually since the end of 2016, and it's extraordinarily rare for it to go down thanks to Gin's bulletproof stability. I've also been quite happy with MongoDB as the app's database. I have directly maintained the Go codebases myself for the past ~5 years, so I have spent countless hours building out new features, fixing bugs, and thinking deep on architecture stack and code structure. I have had a number of thorough cleanup cycles where I did iterative architecture overhauls to make sure it's always getting better.

Here are the problems that I kept wrestling with, which rp is designed to solve.

1. Lack of Readability / Transparency of High Level Server Route Code.

   I want to be able to skim the entry point of a server route's handler code and immediately know what the required request input struct is, what its response structures are, what application logic it follows during execution, and whether it takes any actions external services, e.g. sending and email or running a credit card charge. Breaking certain functionality out into internal packages seemed like a clean structure initially, but it had the byproduct of burying these code details into functions call stacks, and it was tough to handle just with documentation and smart function naming.

2. Following Go's common error checking style of "if err != nil" statements led to a lot of redundant Gin response code.

   Overall this makes it more error prone and harder to maintain code.
   I had 100's of occurences of "c.JSON(http.StatusOK" and a lot of c.JSON calls for statuses like http.StatusBadRequest. This lead me to constantly making microdecisions during development around which response code to use and error message to send. Yet the overall behavior was very similar for every server route: 1) If err != nil then send an error response and stop route execution, 2) If no function ever returns a non-nil error, then send the success response at the end.
   So I want to separate server error responses from internal Go function errors more systematically so that I don't have to pepper network code and thinking into Go application logic. I also want some strong default structures so that, in most cases, I'm happy allow defaults and not even think about what the server response is. If I do need to explicitly set it, then I only do it once and in one place.

3. Every route also includes code that is unique to itself, yet is very similar to code in a bunch of other routes.

   This made code updates on routes frustratingly slow, adding to the effect of #1. (Add more...)

4. Latency optimization was customized to each route, yet the mitigations used tricks like parallelization frequently and in the same way. (The other main way to reduce latency was to reorder and/or consolidate calls to mongodb, which is slow)

5. Developer experience building routes was often cludgy and slow. Debuggers like Delve were OK but tricky to work with. Without this print statements were the typical way, but they're annoying to write and then remove. You also had to kick of local server instances, and that added some overhead & ambiguities about host environment. Now that I've learned Next.js, I want server route building in Go to be like that: Builds happen automatically in the background as you're coding, triggered by source file saves, and the new version, with a lot of helpful debug information, is instantly visible on another window on your desktop. Having a structure that solves for 1, 2, and 3 above gives you the foundation to make it all work with just a few, probably easy-to-make framework components.

6. More fine-grained, automatic, and well-labeled version data would help with debugging a lot. (I'm planning to version every route and every stage with a hash and human-readable metadata, and that can be bundled into your server releases.)

(More to come here. The images below will be explained soon.)

<img align="left" width="588px" src="docs/rpFrameworkDemoCode1.png">
<img align="left" width="673px" src="docs/rpFrameworkDemoCode2.png">
<img align="left" width="711px" src="docs/rpFrameworkDemoCode3.png">
<img align="left" width="766px" src="docs/rpFrameworkDemoCode4.png">
