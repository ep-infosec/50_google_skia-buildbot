package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/unrolled/secure"

	"go.skia.org/infra/go/baseapp"
	"go.skia.org/infra/go/httputils"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/scrap/go/client"
	"go.skia.org/infra/scrap/go/fakeclient"
	"go.skia.org/infra/scrap/go/scrap"
)

// flags
var (
	scrapExchange     = flag.String("scrapexchange", "http://scrapexchange:9000", "Scrap exchange service HTTP address.")
	fakeScrapExchange = flag.Bool("fake_scrapexchange", false, "If set to true, --scrapexchange will be ignored and a fake, in-memory implementation will be used instead.")
)

// server is the state of the server.
type server struct {
	scrapClient scrap.ScrapExchange
	templates   *template.Template
}

// See baseapp.Constructor.
func new() (baseapp.App, error) {
	// Need to set the mime-type for wasm files so streaming compile works.
	if err := mime.AddExtensionType(".wasm", "application/wasm"); err != nil {
		return nil, skerr.Wrap(err)
	}
	var scrapClient scrap.ScrapExchange
	if *fakeScrapExchange {
		sklog.Infof("Using fake (in-memory) scrapexchange client")
		scrapClient = fakeclient.New(map[string]scrap.ScrapBody{
			"@default": {
				Type: "sksl",
				Body: blueNeuronShaderBody,
				SKSLMetaData: &scrap.SKSLMetaData{
					ImageURL: "/dist/mandrill.png",
				},
			},
		})
	} else {
		var err error
		scrapClient, err = client.New(*scrapExchange)
		if err != nil {
			sklog.Fatalf("Failed to create scrap exchange client: %s", err)
		}
	}

	srv := &server{
		scrapClient: scrapClient,
	}
	srv.loadTemplates()
	return srv, nil
}

func (srv *server) loadTemplates() {
	srv.templates = template.Must(template.New("").Delims("{%", "%}").ParseFiles(
		filepath.Join(*baseapp.ResourcesDir, "main.html"),
		filepath.Join(*baseapp.ResourcesDir, "debugger.html"),
	))
}

func (srv *server) pageHandler(w http.ResponseWriter, r *http.Request, p string) {
	w.Header().Set("Content-Type", "text/html")
	if *baseapp.Local {
		srv.loadTemplates()
	}
	if err := srv.templates.ExecuteTemplate(w, p, map[string]string{
		// Look in //shaders/pages/BUILD.bazel for where the nonce templates are injected.
		"Nonce": secure.CSPNonce(r.Context()),
	}); err != nil {
		sklog.Errorf("Failed to expand template: %s", err)
	}
}

func (srv *server) mainHandler(w http.ResponseWriter, r *http.Request) {
	srv.pageHandler(w, r, "main.html")
}

func (srv *server) debugHandler(w http.ResponseWriter, r *http.Request) {
	srv.pageHandler(w, r, "debugger.html")
}

func (srv *server) loadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	hashOrName := mux.Vars(r)["hashOrName"]

	body, err := srv.scrapClient.LoadScrap(r.Context(), scrap.SKSL, hashOrName)
	if err != nil {
		httputils.ReportError(w, err, "Failed to read JSON file.", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		sklog.Errorf("Failed to write response: %s", err)
	}
}

func (srv *server) saveHandler(w http.ResponseWriter, r *http.Request) {
	// Decode Request.
	var req scrap.ScrapBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.ReportError(w, err, "Error decoding JSON.", http.StatusBadRequest)
		return
	}
	if req.Type != scrap.SKSL {
		httputils.ReportError(w, fmt.Errorf("Received invalid scrap type: %q", req.Type), "Invalid Type.", http.StatusBadRequest)
		return
	}

	scrapID, err := srv.scrapClient.CreateScrap(r.Context(), req)
	if err != nil {
		httputils.ReportError(w, err, "Error creating scrap.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scrapID); err != nil {
		sklog.Errorf("Failed to write response: %s", err)
	}
}

// See baseapp.App.
func (srv *server) AddHandlers(r *mux.Router) {
	r.HandleFunc("/", srv.mainHandler)
	r.HandleFunc("/debug", srv.debugHandler)
	r.HandleFunc("/_/load/{hashOrName:[@0-9a-zA-Z-_]+}", srv.loadHandler).Methods("GET")
	r.HandleFunc("/_/save/", srv.saveHandler).Methods("POST")
}

// See baseapp.App.
func (srv *server) AddMiddleware() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{}
}

func main() {
	baseapp.Serve(new, []string{"shaders.skia.org"}, baseapp.AllowWASM{}, baseapp.AllowAnyImage{})
}

// This is the same shader that is the current default on shaders.skia.org (the
// blue neuron-looking one).
const blueNeuronShaderBody = `
// Source: @notargs https://twitter.com/notargs/status/1250468645030858753
float f(vec3 p) {
    p.z -= iTime * 10.;
    float a = p.z * .1;
    p.xy *= mat2(cos(a), sin(a), -sin(a), cos(a));
    return .1 - length(cos(p.xy) + sin(p.yz));
}

half4 main(vec2 fragcoord) {
    vec3 d = .5 - fragcoord.xy1 / iResolution.y;
    vec3 p=vec3(0);
    for (int i = 0; i < 32; i++) {
      p += f(p) * d;
    }
    return ((sin(p) + vec3(2, 5, 9)) / length(p)).xyz1;
}`
