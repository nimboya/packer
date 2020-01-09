package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestDecodeConfig_basic(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true
	}`

	var cfg config
	err := decodeConfig(strings.NewReader(packerConfig), &cfg)
	if err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	var expectedCfg config
	json.NewDecoder(strings.NewReader(packerConfig)).Decode(&expectedCfg)
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("failed to load custom configuration data; expected %v got %v", expectedCfg, cfg)
	}

}

func TestLoadExternalComponentsFromConfig(t *testing.T) {
	packerConfigData, cleanUpFunc, err := generateFakePackerConfigData()
	if err != nil {
		t.Fatalf("error encountered while creating fake Packer configuration data %v", err)
	}
	defer cleanUpFunc()

	var cfg config
	cfg.Builders = packer.MapOfBuilder{}
	cfg.PostProcessors = packer.MapOfPostProcessor{}
	cfg.Provisioners = packer.MapOfProvisioner{}

	if err := decodeConfig(strings.NewReader(packerConfigData), &cfg); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
	}

	if err := cfg.loadExternalComponentsFromConfig(); err != nil {
		t.Fatalf("error encountered discovering external components from configuration file: %v", err)
	}

	if !cfg.Builders.Has("cloud-xyz") {
		t.Errorf("failed to load external builders; got %#v as the resulting config", cfg)
	}

	if !cfg.Provisioners.Has("super-shell") {
		t.Errorf("failed to load external provisioners; got %#v as the resulting config", cfg)
	}

	if !cfg.PostProcessors.Has("noop") {
		t.Errorf("failed to load external post-processors; got %#v as the resulting config", cfg)
	}
}

/* generateFakePackerConfigData creates a collection of mock plugins along with a basic packerconfig.
The return packerConfigData is a valid packerconfig file that can be used for configuring external plugins,
cleanUpFunc is a function that should be called for cleaning up any generated mock data.

This function will only clean up if there is an error, on successful runs the caller
is responsible for cleaning up the data via defer cleanUpFunc().
*/
func generateFakePackerConfigData() (packerConfigData string, cleanUpFunc func(), err error) {
	dir, err := ioutil.TempDir("", "random-testdata")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary test directory: %v", err)
	}

	cleanUpFunc = func() {
		os.RemoveAll(dir)
	}

	plugins := [...]string{
		filepath.Join(dir, "packer-builder-cloud-xyz"),
		filepath.Join(dir, "packer-provisioner-super-shell"),
		filepath.Join(dir, "packer-post-processor-noop"),
	}
	for _, plugin := range plugins {
		_, err := os.Create(plugin)
		if err != nil {
			cleanUpFunc()
			return "", nil, fmt.Errorf("failed to create temporary plugin file (%s): %v", plugin, err)
		}
	}

	packerConfigData = fmt.Sprintf(`
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true,
		"builders": {
			"cloud-xyz": %q
		},
		"provisioners": {
			"super-shell": %q
		},
		"post-processors": {
			"noop": %q
		}
	}`, plugins[0], plugins[1], plugins[2])

	return
}
