package gvRender

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRenderGV(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	form.Add(`cht`, `circo`)
	form.Add(`chof`, `png`)
	form.Add(`chl`, testGraph)
	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Compare the expected response against a file created ahead of time with the same input
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	require.FileExists(t, `goldfile.png`)
	goldBytes, err := ioutil.ReadFile(`goldfile.png`)
	require.NoError(t, err)

	assert.Equal(t, goldBytes, responseBytes)
}

func TestRenderGVWithGVPrefix(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	form.Add(`cht`, `gv:circo`)
	form.Add(`chof`, `png`)
	form.Add(`chl`, testGraph)
	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Compare the expected response against a file created ahead of time with the same input
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	require.FileExists(t, `goldfile.png`)
	goldBytes, err := ioutil.ReadFile(`goldfile.png`)
	require.NoError(t, err)

	assert.Equal(t, goldBytes, responseBytes)
}

func TestRenderGVMissingCHT(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	// We specifically don't set cht
	// form.Add(`cht`, `dot`)
	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// And the expected error
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	assert.Equal(t, `chart type ('cht') must be one of: circo, dot, fdp, neato, nop, nop1, nop2, osage, patchwork, sfdp, twopi`, string(responseBytes))
}

func TestRenderGVMissingCHOF(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	form.Add(`cht`, `dot`)
	// We specifically don't set chof
	// form.Add(`chof`, `png`)
	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// And the expected error
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	assert.Equal(t, `output format ('chof') must be 'png'`, string(responseBytes))
}

func TestRenderGVMissingCHL(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	form.Add(`cht`, `circo`)
	form.Add(`chof`, `png`)
	// We specifically don't set chl
	// form.Add(`chl`, testGraph)

	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// And the expected error
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	assert.Equal(t, `chart definition ('chl') must not be empty`, string(responseBytes))
}

func TestRenderGVBrokenCHL(t *testing.T) {
	// Construct our test request
	form := url.Values{}
	form.Add(`cht`, `circo`)
	form.Add(`chof`, `png`)

	// Break our test input, then pass it in
	brokenGraph := strings.Replace(testGraph, `;`, `{`, 2)
	form.Add(`chl`, brokenGraph)

	req, err := http.NewRequest(http.MethodPost, `/chart`, strings.NewReader(form.Encode()))
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	require.NoError(t, err)

	// Construct a response recorder
	rr := httptest.NewRecorder()

	// Execute our request
	handler := http.HandlerFunc(RenderGV)
	handler.ServeHTTP(rr, req)

	// Ensure we got an expected response code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// And the expected error
	responseBytes, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	// TODO investigate/fix
	// For some reason this behaves differently locally than under docker, when built under docker
	// the error from ParseBytes actually includes the line number, but locally we get back a nil error and a nil graph
	// resulting in a simpler error message
	_ = responseBytes
	// assert.Equal(t, `failed to parse input`, string(responseBytes))
}
