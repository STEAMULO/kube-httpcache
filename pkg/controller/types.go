package controller

import (
	"io"
	"io/ioutil"
	"text/template"

	"github.com/mittwald/kube-httpcache/pkg/signaller"
	"github.com/mittwald/kube-httpcache/pkg/watcher"
)

type TemplateData struct {
	Frontends       watcher.EndpointList
	PrimaryFrontend *watcher.Endpoint
	Backends        watcher.EndpointList
	PrimaryBackend  *watcher.Endpoint
}

type VarnishController struct {
	SecretFile           string
	Storage              string
	AdditionalParameters string
	FrontendAddr         string
	FrontendPort         int
	AdminAddr            string
	AdminPort            int

	vclTemplate        *template.Template
	vclTemplateUpdates chan []byte
	frontendUpdates    chan *watcher.EndpointConfig
	frontend           *watcher.EndpointConfig
	backendUpdates     chan *watcher.EndpointConfig
	backend            *watcher.EndpointConfig
	varnishSignaller   *signaller.Signaller
	configFile         string
	secret             []byte
	localAdminAddr     string
	currentVCLName     string
}

func NewVarnishController(
	secretFile string,
	storage string,
	additionalParameter string,
	frontendAddr string,
	frontendPort int,
	adminAddr string,
	adminPort int,
	frontendUpdates chan *watcher.EndpointConfig,
	backendUpdates chan *watcher.EndpointConfig,
	templateUpdates chan []byte,
	varnishSignaller *signaller.Signaller,
	vclTemplateFile string,
) (*VarnishController, error) {
	contents, err := ioutil.ReadFile(vclTemplateFile)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("vcl").Parse(string(contents))
	if err != nil {
		return nil, err
	}

	secret, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return nil, err
	}

	return &VarnishController{
		SecretFile:           secretFile,
		Storage:              storage,
		AdditionalParameters: additionalParameter,
		FrontendAddr:         frontendAddr,
		FrontendPort:         frontendPort,
		AdminAddr:            adminAddr,
		AdminPort:            adminPort,
		vclTemplate:          tmpl,
		vclTemplateUpdates:   templateUpdates,
		frontendUpdates:      frontendUpdates,
		backendUpdates:       backendUpdates,
		varnishSignaller:     varnishSignaller,
		configFile:           "/tmp/vcl",
		secret:               secret,
	}, nil
}

func (v *VarnishController) renderVCL(target io.Writer, frontendList watcher.EndpointList, primaryFrontend *watcher.Endpoint, backendList watcher.EndpointList, primaryBackend *watcher.Endpoint) error {
	err := v.vclTemplate.Execute(target, &TemplateData{
		Frontends:       frontendList,
		PrimaryFrontend: primaryFrontend,
		Backends:        backendList,
		PrimaryBackend:  primaryBackend,
	})

	return err
}
