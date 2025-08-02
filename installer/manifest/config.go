package manifest

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stackrox/rox/pkg/utils"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

const (
	localStackroxImage = "localhost:5001/stackrox/stackrox:latest"
	localDbImage       = "localhost:5001/stackrox/db:latest"
)

type EnvVarConfig struct {
	Global     []v1.EnvVar            `yaml:"global"`
	Generators map[string][]v1.EnvVar `yaml:"generators"`
	Pods       map[string][]v1.EnvVar `yaml:"pods"`
	Containers map[string][]v1.EnvVar `yaml:"containers"`
}

type Config struct {
	Action               string       `yaml:"action"`
	ApplyNetworkPolicies bool         `yaml:"applyNetworkPolicies"`
	CRS                  CRS          `yaml:"crs"`
	CertPath             string       `yaml:"certPath"`
	DevMode              bool         `yaml:"devMode"`
	EnvVars              EnvVarConfig `yaml:"envVars"`
	Images               Images       `yaml:"images"`
	ImageArchitecture    string       `yaml:"imageArchitecture"`
	Namespace            string       `yaml:"namespace"`
	ScannerV4            bool         `yaml:"scannerV4"`
}

type CRS struct {
	PortForward bool `yaml:"portForward"`
}

type Images struct {
	AdmissionControl string `yaml:"admissionControl"`
	Sensor           string `yaml:"sensor"`
	Collector        string `yaml:"collector"`
	ConfigController string `yaml:"configController"`
	Central          string `yaml:"central"`
	CentralDB        string `yaml:"centralDb"`
	Scanner          string `yaml:"scanner"`
	ScannerDB        string `yaml:"scannerDb"`
	ScannerV4        string `yaml:"scannerv4"`
	ScannerV4DB      string `yaml:"scannerv4Db"`
	VSOCKListener    string `yaml:"vsockListener"`
}

var DefaultConfig Config = Config{
	Namespace:            "stackrox",
	ScannerV4:            false,
	DevMode:              false,
	ApplyNetworkPolicies: false,
	CertPath:             "./certs",
	ImageArchitecture:    "single",
	EnvVars: EnvVarConfig{
		Global:     []v1.EnvVar{},
		Generators: make(map[string][]v1.EnvVar),
		Pods:       make(map[string][]v1.EnvVar),
		Containers: make(map[string][]v1.EnvVar),
	},
	Images: Images{
		AdmissionControl: localStackroxImage,
		Sensor:           localStackroxImage,
		Collector:        localStackroxImage,
		ConfigController: localStackroxImage,
		Central:          localStackroxImage,
		Scanner:          localStackroxImage,
		ScannerV4:        localStackroxImage,
		VSOCKListener:    localStackroxImage,
		CentralDB:        localDbImage,
		ScannerDB:        localDbImage,
		ScannerV4DB:      localDbImage,
	},
}

func ReadConfig(filename string) (*Config, error) {
	if filename == "" {
		cfg := DefaultConfig
		return &cfg, nil
	}
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer utils.IgnoreError(r.Close)
	return load(r)
}

func load(r io.Reader) (*Config, error) {
	yd := yaml.NewDecoder(r)
	yd.KnownFields(true)
	cfg := DefaultConfig
	if err := yd.Decode(&cfg); err != nil {
		msg := strings.TrimPrefix(err.Error(), `yaml: `)
		return nil, fmt.Errorf("malformed yaml: %v", msg)
	}
	return &cfg, nil
}
