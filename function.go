package gvRender

import (
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

// Go is a stupid language sometimes
var supportedCharts = map[string]bool{
	`circo`:     true,
	`dot`:       true,
	`fdp`:       true,
	`neato`:     true,
	`nop`:       true,
	`nop1`:      true,
	`nop2`:      true,
	`osage`:     true,
	`patchwork`: true,
	`sfdp`:      true,
	`twopi`:     true,
}

func RenderGV(w http.ResponseWriter, r *http.Request) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})

	entry := logrus.NewEntry(logrus.StandardLogger())

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf(`failed to parse request body: %s`, err.Error()))
		entry.WithError(err).Error(`failed to parse request body`)
		return
	}

	// Chart Type, must be supported by https://github.com/goccy/go-graphviz
	// so one of circo dot fdp neato nop nop1 nop2 osage patchwork sfdp twopi
	chartType := r.Form.Get(`cht`)

	// Maintaining google charts api compatibility may mean a gv prefix on cht, remove it if we find it
	if strings.HasPrefix(chartType, `gv:`) {
		chartType = strings.TrimPrefix(chartType, `gv:`)
	}

	entry = entry.WithField(`chartType`, chartType)

	if _, ok := supportedCharts[chartType]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `chart type ('cht') must be one of: circo, dot, fdp, neato, nop, nop1, nop2, osage, patchwork, sfdp, twopi`)
		entry.Error(`chart type ('cht') must be one of: circo, dot, fdp, neato, nop, nop1, nop2, osage, patchwork, sfdp, twopi`)
		return
	}

	// Output Format, we only support png right now
	outputFormat := r.Form.Get(`chof`)
	entry = entry.WithField(`outputFormat`, outputFormat)
	if outputFormat != `png` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `output format ('chof') must be 'png'`)
		entry.Error(`output format ('chof') must be 'png'`)
		return
	}

	// Chart dot lang input, html escaped
	chartInput := r.Form.Get(`chl`)
	if strings.TrimSpace(chartInput) == `` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `chart definition ('chl') must not be empty`)
		entry.Error(`chart definition ('chl') must not be empty`)
		return
	}

	// Parse the user input to a graph
	graph, err := graphviz.ParseBytes([]byte(chartInput))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`failed to parse input: %s`, err.Error()))
		entry.WithError(err).Error(`failed to parse chart input`)
		return
	}

	if graph == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `failed to parse input`)
		entry.Error(`failed to parse chart input`)
		return
	}

	// Render the graph to a png, and return it in our response
	graphvizRenderer := graphviz.New()

	// Apply the chart layout style
	graph = graph.SetLayout(chartType)

	// Render the graph to a PNG format, and serve it as our response
	if err := graphvizRenderer.Render(graph, graphviz.PNG, w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`failed to render input: %s`, err.Error()))
		entry.WithError(err).Error(`failed to render input`)
		return
	}

	entry.Info(`OK`)
}
