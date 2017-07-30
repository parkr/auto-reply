# auto-reply

An open source gardener. This is a technology for powering GitHub bots. It's really rough around the edges but it currently powers [@jekyllbot](https://github.com/jekyllbot).

[![Build Status](https://travis-ci.org/parkr/auto-reply.svg?branch=master)](https://travis-ci.org/parkr/auto-reply)

## Configuring

If you want to configure a secret to validate your payload from GitHub,
then set it as the environment variable `GITHUB_WEBHOOK_SECRET`. This is
the same value you enter in the web interface when setting up the "Secret"
for your webhook.

I could use [your thoughts on this!](https://github.com/parkr/auto-reply/issues/4) Currently, it's a hodge-podge. The documentation for each package will provide more details on this. Currently we have the following packages, with varying levels of configuration:

- `affinity` – assigns issues based on team mentions and those team captains. See [Jekyll's docs for more info.](https://github.com/jekyll/jekyll/blob/master/docs/affinity-team-captain.md)
- `autopull` – detects pushes to branches which start with `pull/` and automatically creates a PR for them
- `chlog` – creates GitHub releases when a new tag is pushed, and powers "@jekyllbot: merge (+category)"
- `jekyll/deprecate` – comments on and closes issues to issues on certain repos with a per-repo stock message
- `jekyll/issuecomment` – provides handlers for removing `pending-feedback` and `stale` labels when a comment comes through
- `labeler` – removes `pending-rebase` label when a PR is pushed to and is mergeable (and helper functions for manipulating labels)
- `lgtm` – adds a `jekyllbot/lgtm` CI status and handles `LGTM` counting

## Installing

This is intended for use with servers, so you'd do something like:

```go
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/parkr/auto-reply/affinity"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
)

var context *ctx.Context

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	context = ctx.NewDefaultContext()

	http.HandleFunc("/_ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok\n"))
	}))

	// Add your event handlers. Check out the documentation for the
	// github.com/parkr/auto-reply/hooks package to see all supported events.
	eventHandlers := hooks.EventHandlerMap{}

	// Build the affinity handler.
	aff := &affinity.Handler{}
	aff.AddRepo("myorg", "myproject")
	aff.AddTeam(context, 123) // @myorg/performance
	aff.AddTeam(context, 456) // @myorg/documentation

	// Add the affinity handler's various event handlers to the event handlers map :)
	eventHandlers.AddHandler(hooks.IssuesEvent, aff.AssignIssueToAffinityTeamCaptain)
	eventHandlers.AddHandler(hooks.IssueCommentEvent, aff.AssignIssueToAffinityTeamCaptainFromComment)
	eventHandlers.AddHandler(hooks.PullRequestEvent, aff.RequestReviewFromAffinityTeamCaptain)

	// Create the webhook handler. GlobalHandler takes the list of event handlers from
	// its configuration and fires each of them based on the X-GitHub-Event header from
	// the webhook payload.
	myOrgHandler := &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: eventHandlers,
	}
	http.Handle("/_github/myproject", myOrgHandler)

	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

## Writing Custom Handlers

For now, all you have to do is write a function which satisfies the `hooks.EventHandler` type. At the moment, each handler can accept only **one** type of event. If you want to accept the `issue_comment` event, then you should be able to perform a successful type assertion:

```go
func MyIssueCommentHandler(context *ctx.Context, payload interface{}) error {
    event, err := payload.(*github.IssueCommentEvent)
    if err != nil {
        return context.NewError("MyIssueCommentHandler: hm, didn't get an IssueCommentEvent: %v", err)
    }

    // Handle your issue comment event in a type-safe way here.
}
```

Then you register that with your project. Taking the two examples above, you'd add `MyIssueCommentHandler` to the `eventHandlers[hooks.IssueCommentEvent]` array:

```go
eventHandlers := hooks.EventHandlerMap{}
eventHandlers.AddHandler(hooks.IssueCommentEvent, MyIssueCommentHandler)
```

And it should work!

## Optional: Mark-and-sweep Stale Issues

One big issue we have in Jekyll is "stale" issues, that is, issues which were opened and abandoned after a few months of activity. The code in `cmd/mark-and-sweep-stale-issues` is still Jekyll-specific but I'd love a PR which abstracts out the configuration into a file or something!

## License

This code is licensed under BSD 3-clause as specified in the [LICENSE](LICENSE) file in this repository.
