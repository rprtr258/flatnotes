package healthcheck

import (
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"

	"github.com/rprtr258/flatnotes/internal"
)

type diagnostic struct {
	note       string
	text       string // TODO: positions
	group      string
	diagnostic string
}

// Run executes all health checks and prints results to stdout.
// Intended to be extended with more checks over time; each check is a
// function that prints its findings to the terminal.
func Run(app internal.App) {
	diagnostics := iter.Concat(
		checkNoteLinks(app.Notes, app.Dir),
	).Slice()
	for k, v := range fun.GroupBy(func(d diagnostic) string {
		return d.note
	}, diagnostics...) {
		_ = k
		_ = v
		// fmt.Println(k)
		// for _, x := range v {
		// 	fmt.Printf("%s[%s]: %s\n",x.group,x.diagnostic,x.text)
		// }
		// fmt.Println()
	}
}
