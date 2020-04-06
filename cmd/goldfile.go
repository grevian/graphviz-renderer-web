// +build goldfile

package main

import (
	"os"

	"github.com/goccy/go-graphviz"
	"github.com/sirupsen/logrus"
)

const testGraph = `
graph {
    a -- b;
    b -- c;
    a -- c;
    d -- c;
    e -- c;
    e -- a;
}
`

// Generate a goldfile for use in tests
func main() {
	renderer := graphviz.New()
	graph, err := graphviz.ParseBytes([]byte(testGraph))
	if err != nil {
		logrus.WithError(err).Fatal(`failed to parse test graph`)
	}

	// Create our output file
	f, err := os.Create(`./goldfile.png`)
	if err != nil {
		logrus.WithError(err).Fatal(`failed to create goldfile output`)
	}
	defer f.Close()

	// Render the graph to the file
	err = renderer.Render(graph, graphviz.PNG, f)
	if err != nil {
		logrus.WithError(err).Fatal(`failed to render graph`)
	}
}
