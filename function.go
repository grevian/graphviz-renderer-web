package gvRender

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/goccy/go-graphviz"
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
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf(`failed to parse request body: %s`, err.Error()))
		return
	}

	// Chart Type, must be supported by https://github.com/goccy/go-graphviz
	// so one of circo dot fdp neato nop nop1 nop2 osage patchwork sfdp twopi
	chartType := r.Form.Get(`cht`)
	if _, ok := supportedCharts[chartType]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `chart type ('cht') must be one of: circo, dot, fdp, neato, nop, nop1, nop2, osage, patchwork, sfdp, twopi`)
		return
	}

	// Output Format, we only support png right now
	outputFormat := r.Form.Get(`chof`)
	if outputFormat != `png` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `output format ('chof') must be 'png'`)
		return
	}

	// Chart dot lang input, html escaped
	chartInput := r.Form.Get(`chl`)
	if strings.TrimSpace(chartInput) == `` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `chart definition ('chl') must not be empty`)
		return
	}

	// Parse the user input to a graph
	graph, err := graphviz.ParseBytes([]byte(chartInput))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`failed to render input: %s`, err.Error()))
		return
	}

	if graph == nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `failed to render input`)
		return
	}

	// Render the graph to a png, and return it in our response
	graphvizRenderer := graphviz.New()
	if err := graphvizRenderer.Render(graph, graphviz.PNG, w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`failed to render input: %s`, err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
