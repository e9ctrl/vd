package api

import (
	"net/http"
	"testing"

	"github.com/e9ctrl/vd/stream"
)

const (
	FILE1 = "../stream/vdfile"
	FILE2 = "../stream/vdfile_delays"
	FILE3 = "../stream/vdfile_mismatch"
)

func TestGetMismatch(t *testing.T) {
	t.Parallel()
	vdfile, err := stream.ReadVDFile(FILE3)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
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
	vdfile, err := stream.ReadVDFile(FILE3)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
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
	vdfile, err := stream.ReadVDFile(FILE1)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
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
		{"get wrong paramter", "test", "Error: parameter test not found", http.StatusInternalServerError},
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
	vdfile, err := stream.ReadVDFile(FILE3)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
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
		{"set wrong parameter", "test", "20.1", "Error: parameter test not found", http.StatusInternalServerError, "Error: parameter test not found", http.StatusInternalServerError},
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

func TestGetGlobDel(t *testing.T) {
	t.Parallel()
	vdfile, err := stream.ReadVDFile(FILE2)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name    string
		typ     string
		exp     string
		expCode int
	}{
		{"get global result delay", "res", "1s", http.StatusOK},
		{"get global acknowledge delay", "ack", "1s", http.StatusOK},
		{"wrong delay type", "test", "Error: delay test not found", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/delay/"+tt.typ)
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

func TestSetGlobDel(t *testing.T) {
	t.Parallel()
	vdfile, err := stream.ReadVDFile(FILE2)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name       string
		typ        string
		set        string
		expSet     string
		expSetCode int
		expGet     string
		expGetCode int
	}{
		{"set global result delay", "res", "2s", "Delay set successfully", http.StatusOK, "2s", http.StatusOK},
		{"set global acknowledge delay", "ack", "3s", "Delay set successfully", http.StatusOK, "3s", http.StatusOK},
		{"set wrong delay type", "test", "5s", "Error: delay test not found", http.StatusInternalServerError, "Error: delay test not found", http.StatusInternalServerError},
		{"set wrong delay duration", "res", "test", "Error: time: invalid duration \"test\"", http.StatusInternalServerError, "2s", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.set(t, "/delay/"+tt.typ+"/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/delay/"+tt.typ)
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

func TestGetDel(t *testing.T) {
	t.Parallel()
	vdfile, err := stream.ReadVDFile(FILE2)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name    string
		typ     string
		param   string
		exp     string
		expCode int
	}{
		{"get psi result delay", "res", "psi", "3s", http.StatusOK},
		{"get psi acknowledge delay", "ack", "psi", "3s", http.StatusOK},
		{"get temp result delay", "res", "temp", "1s", http.StatusOK},
		{"get temp acknowledge delay", "ack", "temp", "1s", http.StatusOK},
		{"get wrong type of delay", "test", "psi", "Error: delay test not found", http.StatusInternalServerError},
		{"get wrong parameter name", "res", "test", "Error: param test not found", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/delay/"+tt.typ+"/"+tt.param)
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

func TestSetDel(t *testing.T) {
	t.Parallel()
	vdfile, err := stream.ReadVDFile(FILE2)
	if err != nil {
		t.Fatal(err)
	}

	dev, err := stream.NewDevice(vdfile)
	if err != nil {
		t.Fatal(err)
	}

	a := &api{
		d: dev,
	}

	ts := newTestServer(t, a.routes())

	defer ts.Close()

	tests := []struct {
		name       string
		typ        string
		param      string
		set        string
		expSet     string
		expSetCode int
		expGet     string
		expGetCode int
	}{
		{"set current result delay", "res", "current", "2s", "Delay set successfully", http.StatusOK, "2s", http.StatusOK},
		{"set current acknowledge delay", "ack", "current", "3s", "Delay set successfully", http.StatusOK, "3s", http.StatusOK},
		{"set wrong delay type", "test", "psi", "5s", "Error: delay test not found", http.StatusInternalServerError, "Error: delay test not found", http.StatusInternalServerError},
		{"set wrong parameter name", "res", "test", "5s", "Error: param test not found", http.StatusInternalServerError, "Error: param test not found", http.StatusInternalServerError},
		{"set wrong delay value", "res", "current", "10test", "Error: time: unknown unit \"test\" in duration \"10test\"", http.StatusInternalServerError, "2s", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.set(t, "/delay/"+tt.typ+"/"+tt.param+"/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/delay/"+tt.typ+"/"+tt.param)
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
