package server

import (
	"encoding/json"
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

type graphRequest struct {
	Cht  string `json:"cht"`  // Chart Type
	Chl  string `json:"chl"`  // Chart Language definition
	Chof string `json:"chof"` // Chart Output format
}

func RenderGV(w http.ResponseWriter, r *http.Request) {
	// These mappings help logrus better integrate with cloud run log formatting
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})
	entry := logrus.NewEntry(logrus.StandardLogger())

	request, err := readPayload(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		entry.WithError(err).Error(`failed to read request payload`)
	}
	entry = entry.WithFields(logrus.Fields{
		"Chl":  request.Chl,
		"Chof": request.Chof,
		"Cht":  request.Cht,
	})

	// Chart Type, must be supported by https://github.com/goccy/go-graphviz
	// so one of circo dot fdp neato nop nop1 nop2 osage patchwork sfdp twopi
	chartType := request.Cht

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
	outputFormat := request.Chof
	entry = entry.WithField(`outputFormat`, outputFormat)
	if outputFormat != `png` {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `output format ('chof') must be 'png'`)
		entry.Error(`output format ('chof') must be 'png'`)
		return
	}

	// Chart dot lang input, html escaped
	chartInput := request.Chl
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

	// Something internal to the graphviz libraries can return a nil graph and no error sometimes
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

func readPayload(r *http.Request) (graphRequest, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		return parseJsonBody(r)
	} else if contentType == "application/x-www-form-urlencoded" {
		return parseFormBody(r)
	} else {
		return graphRequest{}, fmt.Errorf("unexpected content-type: %s", contentType)
	}
}

func parseJsonBody(r *http.Request) (graphRequest, error) {
	// Read the request body, then deserialize it
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return graphRequest{}, fmt.Errorf("failed to read request body: %w", err)
	}
	var request graphRequest
	err = json.Unmarshal(bodyBytes, &request)
	if err != nil {
		return graphRequest{}, fmt.Errorf("failed to parse request body: %w", err)
	}

	return request, nil
}

func parseFormBody(r *http.Request) (graphRequest, error) {
	// Read the request form
	err := r.ParseForm()
	if err != nil {
		return graphRequest{}, fmt.Errorf("failed to parse request form: %w", err)
	}

	var request graphRequest
	request.Chl = r.Form.Get("chl")
	request.Cht = r.Form.Get("cht")
	request.Chof = r.Form.Get("chof")

	return request, nil
}
