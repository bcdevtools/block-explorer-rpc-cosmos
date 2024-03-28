package config

import (
	"bytes"
	tmconfig "github.com/tendermint/tendermint/config"
	"path/filepath"
	"strings"
	"text/template"

	tmos "github.com/tendermint/tendermint/libs/os"
)

const (
	DefaultConfigDirName  = "config"
	DefaultConfigName     = "be-json-rpc"
	DefaultConfigFileName = DefaultConfigName + ".toml"
)

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("beJsonRpcConfigFileTemplate").Funcs(template.FuncMap{
		"StringsJoin": strings.Join,
	})
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

/****** these are for production settings ***********/

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func EnsureRoot(rootDir string, defaultConfig *BeJsonRpcConfig) {
	if err := tmos.EnsureDir(rootDir, tmconfig.DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := tmos.EnsureDir(filepath.Join(rootDir, DefaultConfigDirName), tmconfig.DefaultDirPerm); err != nil {
		panic(err.Error())
	}

	if defaultConfig == nil {
		return
	}

	configFilePath := filepath.Join(rootDir, DefaultConfigDirName, DefaultConfigFileName)

	// Write default config file if missing.
	if !tmos.FileExists(configFilePath) {
		WriteConfigFile(configFilePath, defaultConfig)
	}
}

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configFilePath string, config *BeJsonRpcConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0o644)
}

// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in config/config.go
const defaultConfigTemplate = `
#######################################################
###  Block Explorer Json-RPC Configuration Options  ###
#######################################################

# defines if the Be Json RPC server should be enabled.
enable = "{{ .Enable }}"

# defines the HTTP server to listen on
address = "{{ .Address }}"

# http read/write timeout of Block Explorer Json-RPC server.
http-timeout = "{{ .HTTPTimeout }}"

# http idle timeout of Block Explorer Json-RPC server.
http-idle-timeout = "{{ .HTTPIdleTimeout }}"

# maximum number of simultaneous connections for the server listener.
max-open-connections = {{ .MaxOpenConnections }}

# defines if the server should allow CORS requests.
allow-cors = {{ .AllowCORS }}

`
