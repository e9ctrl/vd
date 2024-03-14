package api

import (
	"net/http"
	"testing"

	"github.com/e9ctrl/vd/device"
	"github.com/e9ctrl/vd/vdfile"
)

// path to vdfile used in tests
const FILE1 = "../vdfile/vdfile"

var (
	vdfileTest vdfile.Config
)

func init() {
	// use one, common vdfile as a template to create vdfile.Config structures for tests
	config, err := vdfile.DecodeVDFile(FILE1)
	if err != nil {
		panic(err)
	}

	// add delays to get_psi and get_temp commands
	for i := 0; i < len(config.Commands); i++ {
		switch config.Commands[i].Name {
		case "get_psi":
			config.Commands[i].Dly = "3s"
		case "get_temp":
			config.Commands[i].Dly = "1s"
		}
	}

	config.Mismatch = "Wrong query"
	// vvdfile with changed mismatch message and delays
	vdfileTest = config
}

func TestGetMismatch(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()
	expected := `Wrong query`

	code, _, body := ts.get(t, "/mismatch")
	if code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			code, http.StatusOK)
	}
	if string(body) != expected {
		t.Errorf("handler returned unexpected body: got\n %s want\n %v",
			body, expected)
	}
}

func TestSetMismatch(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()
	expectedSet := `Mismatch set successfully`
	expectedGet := `found error`

	code, _, body := ts.set(t, "/mismatch/found error")
	if code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			code, http.StatusOK)
	}
	if string(body) != expectedSet {
		t.Errorf("handler returned unexpected body: got\n %s want\n %v",
			body, expectedSet)
	}

	code, _, body = ts.get(t, "/mismatch")
	if code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			code, http.StatusOK)
	}
	if string(body) != expectedGet {
		t.Errorf("handler returned unexpected body: got\n %s want\n %v",
			body, expectedGet)
	}
}

func TestGetParameter(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()
	tests := []struct {
		name    string
		param   string
		exp     string
		expCode int
	}{
		{"get version", "version", "version 1.0", http.StatusOK},
		{"get current", "current", "300", http.StatusOK},
		{"get mode", "mode", "NORM", http.StatusOK},
		{"get wrong parameter", "test", "Error: parameter not found: test", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/"+tt.param)
			if code != tt.expCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expCode)
			}
			if string(body) != tt.exp {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.exp)
			}
		})
	}
}

func TestSetParameter(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name       string
		param      string
		set        string
		expSet     string
		expSetCode int
		expGet     string
		expGetCode int
	}{
		{"set psi", "psi", "4.56", "Parameter set successfully", http.StatusOK, "4.56", http.StatusOK},
		{"set current", "current", "34", "Parameter set successfully", http.StatusOK, "34", http.StatusOK},
		{"set wrong psi value", "psi", "5s", "Error: received param type that cannot be converted to float", http.StatusInternalServerError, "4.56", http.StatusOK},
		{"set wrong parameter", "test", "20.1", "Error: parameter not found: test", http.StatusInternalServerError, "Error: parameter not found: test", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.set(t, "/"+tt.param+"/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/"+tt.param)
			if code != tt.expGetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expGetCode)
			}
			if string(body) != tt.expGet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expGet)
			}
		})
	}
}

func TestGetCommandDelay(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name    string
		command string
		exp     string
		expCode int
	}{
		{"get psi result delay", "get_psi", "3s", http.StatusOK},
		{"get temp result delay", "get_temp", "1s", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/delay/"+tt.command)
			if code != tt.expCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expCode)
			}
			if string(body) != tt.exp {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.exp)
			}
		})
	}
}

func TestSetCommandDelay(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileFromConfig(vdfileTest)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := device.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &Api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name       string
		command    string
		set        string
		expSet     string
		expSetCode int
		expGet     string
		expGetCode int
	}{
		{"set current result delay", "set_current", "2s", "Delay set successfully", http.StatusOK, "2s", http.StatusOK},
		{"set wrong command name", "test", "5s", "Error: command not found: test", http.StatusInternalServerError, "Error: command not found: test", http.StatusInternalServerError},
		{"set wrong delay value", "set_current", "10test", "Error: time: unknown unit \"test\" in duration \"10test\"", http.StatusInternalServerError, "2s", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.set(t, "/delay/"+tt.command+"/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/delay/"+tt.command)
			if code != tt.expGetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expGetCode)
			}
			if string(body) != tt.expGet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expGet)
			}
		})
	}
}
