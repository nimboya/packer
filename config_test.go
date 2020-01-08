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
)

func TestDecodeConfig_basic(t *testing.T) {

	packerConfig := `
	{
		"PluginMinPort": 10,
		"PluginMaxPort": 25,
		"disable_checkpoint": true,
		"disable_checkpoint_signature": true,
		"builders": { "cloud-xyz": "packer-builder-cloud-xyz" },
		"provisioners": { "super-shell": "packer-builder-super-shell" },
		"post-processors": { "noop": "packer-post-processor-noop" }
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

func TestLoadFromFile(t *testing.T) {
	configFilePath, err := generateFakePackerConfigData()
	if err != nil {
		t.Fatalf("error encountered while creating fake Packer configuration data %v", err)
	}
	defer destroyFakePackerConfigData(filepath.Base(configFilePath))

	// Set PACKER_CONFIG to test Packer config
	os.Setenv("PACKER_CONFIG", configFilePath)
	var cfg config
	if err := cfg.LoadFromFile(); err != nil {
		t.Fatalf("error encountered decoding configuration: %v", err)
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

/* generateFakePackerConfigData creates a collection of mock plugins along with a
basic packerconfig.  The return string of this function is points to a valid packerconfig
file that can be used for configuring external plugins.

This function will only clean up if there is an error, on successful runs the caller
is responsible for cleaning up the data via destroyFakePackerConfigData(packerConfigPath).
*/
func generateFakePackerConfigData() (string, error) {
	dir, err := ioutil.TempDir("", "random-testdata")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary test directory: %v", err)
	}

	tmpFile, err := ioutil.TempFile("", "packerconfig")
	if err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("failed to create temporary test configuration file: %v", err)
	}

	plugins := [...]string{
		filepath.Join(dir, "packer-builder-cloud-xyz"),
		filepath.Join(dir, "packer-provisioner-super-shell"),
		filepath.Join(dir, "packer-post-processor-noop"),
	}
	for _, plugin := range plugins {
		_, err := os.Create(plugin)
		if err != nil {
			os.RemoveAll(dir)
			return "", fmt.Errorf("failed to create temporary plugin file (%s): %v", plugin, err)
		}
	}

	packerConfig := fmt.Sprintf(`
	{
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

	fmt.Println(packerConfig)

	if _, err := tmpFile.Write([]byte(packerConfig)); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("failed to write testdata to packerconfig file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("failed to close packerconfig file: %v", err)
	}

	return tmpFile.Name(), nil
}

// destroyFakePackerConfigData will destroy all data under path
func destroyFakePackerConfigData(path string) {
	os.RemoveAll(path)
}
