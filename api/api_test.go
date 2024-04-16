package api

import (
	"net/http"
	"testing"

	"github.com/e9ctrl/vd/device"
	"github.com/e9ctrl/vd/vdfile"
)

// path to vdfile used in tests
const (
	FILE_STREAM = "../vdfile/vdfile_stream"
	FILE_MODBUS = "../vdfile/vdfile_modbus"
)

var (
	vdfileTestStream vdfile.ConfigStream
	vdfileTestModbus vdfile.ConfigModbus
)

func init() {
	// use one, common vdfile as a template to create vdfile.ConfigStream structure for tests
	config, err := vdfile.DecodeVDFileStream(FILE_STREAM)
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
	// vdfile with changed mismatch message and delays
	vdfileTestStream = config

	// use one, common vdfile as a template to create vdfile.ConfigModbus structure for tests
	configMod, err := vdfile.DecodeVDFileModbus(FILE_MODBUS)
	if err != nil {
		panic(err)
	}
	configMod.Delay = "3s"
	// vdfile with changed delay
	vdfileTestModbus = configMod
}

func TestGetMismatchStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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

func TestSetMismatchStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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

func TestGetParameterStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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

func TestSetParameterStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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

func TestGetParameterTypeStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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
		{"get version", "version", "string", http.StatusOK},
		{"get current", "current", "int64", http.StatusOK},
		{"get psi", "psi", "float64", http.StatusOK},
		{"get wrong parameter", "test", "Error: parameter not found: test", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/type/"+tt.param)
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

func TestGetCommandDelayStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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
			code, _, body := ts.get(t, "/delay/stream/"+tt.command)
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

func TestSetCommandDelayStream(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileStreamFromConfig(vdfileTestStream)
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
			code, _, body := ts.set(t, "/delay/stream/"+tt.command+"/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/delay/stream/"+tt.command)
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

func TestGetMismatchModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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
	expected := ``

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

func TestSetMismatchModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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

func TestGetParameterModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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
		{"get state", "state", "1", http.StatusOK},
		{"get temp", "temp", "20", http.StatusOK},
		{"get mode", "mode", "1", http.StatusOK},
		{"get pressure", "pressure", "34.5", http.StatusOK},
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

func TestSetParameterModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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
		{"set state3", "state3", "0", "Parameter set successfully", http.StatusOK, "0", http.StatusOK},
		{"set volt", "volt", "50", "Parameter set successfully", http.StatusOK, "50", http.StatusOK},
		{"set wrong temp2 value", "temp2", "4s", "Error: received param type that cannot be converted to int", http.StatusInternalServerError, "454211", http.StatusOK},
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

func TestGetParameterTypeModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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
		{"get state", "state", "uint8", http.StatusOK},
		{"get temp", "temp", "uint16", http.StatusOK},
		{"get volt", "volt", "int64", http.StatusOK},
		{"get wrong parameter", "test", "Error: parameter not found: test", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, "/type/"+tt.param)
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

func TestGetDelayModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
	t.Log(vdfile.Modbus.Delay)
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

	code, _, body := ts.get(t, "/delay/modbus")
	if code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			code, http.StatusOK)
	}
	if string(body) != "3s" {
		t.Errorf("handler returned unexpected body: got\n %s want\n %v",
			body, "3s")
	}
}

func TestSetDelayModbus(t *testing.T) {
	t.Parallel()
	vdfile, err := vdfile.ReadVDFileModbusFromConfig(vdfileTestModbus)
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
		set        string
		expSet     string
		expSetCode int
		expGet     string
		expGetCode int
	}{
		{"set delay", "2s", "Delay set successfully", http.StatusOK, "2s", http.StatusOK},
		{"set wrong delay value", "10test", "Error: time: unknown unit \"test\" in duration \"10test\"", http.StatusInternalServerError, "2s", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.set(t, "/delay/modbus/"+tt.set)
			if code != tt.expSetCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					code, tt.expSetCode)
			}
			if string(body) != tt.expSet {
				t.Errorf("handler returned unexpected body: got\n %s want\n %v",
					body, tt.expSet)
			}

			code, _, body = ts.get(t, "/delay/modbus")
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
