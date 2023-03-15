package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	TEST_CONFIG_FILENAME = ".test_config.yaml"
)

type TestConfig struct {
	SkipTursoTests bool   `koanf:"skip_turso_tests"`
	TursoDbPath    string `koanf:"turso_db_path" validate:"required_if=SkipTursoTests false"`
}

var alreadyReadTestConfig *TestConfig

func GetTestConfig(t *testing.T) TestConfig {
	if alreadyReadTestConfig != nil {
		return *alreadyReadTestConfig
	}

	readTestConfig := readTestConfig(t)
	alreadyReadTestConfig = &readTestConfig
	return *alreadyReadTestConfig
}

func readTestConfig(t *testing.T) TestConfig {
	var koanfInstance = koanf.New(".")
	loadConfigFileIfExists(t, koanfInstance)
	loadEnvironmentVariables(t, koanfInstance)

	return unmarshalAndValidateConfigs(t, koanfInstance)
}

func loadConfigFileIfExists(t *testing.T, koanfInstance *koanf.Koanf) {
	testConfigFilePath := getTestConfigurationPath(t)
	if fileExists(testConfigFilePath) {
		if err := koanfInstance.Load(file.Provider(testConfigFilePath), yaml.Parser()); err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	}
}

func loadEnvironmentVariables(t *testing.T, koanfInstance *koanf.Koanf) {
	err := koanfInstance.Load(env.Provider("TEST_CONFIG_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "TEST_CONFIG_")), "__", ".", -1)
	}), nil)
	if err != nil {
		log.Fatalf("Error reading environment variables: %v", err)
	}
}

func unmarshalAndValidateConfigs(t *testing.T, koanfInstance *koanf.Koanf) TestConfig {
	testConfig := TestConfig{}
	if err := koanfInstance.Unmarshal("", &testConfig); err != nil {
		t.Fatalf("Unable to unmarshal test configs. Error: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(testConfig); err != nil {
		t.Fatalf("Invalid test configs. Error: %v", err)
	}

	return testConfig
}

func getTestConfigurationPath(t *testing.T) string {
	projectRootPath := getProjectRootPath(t)

	return projectRootPath + "/" + TEST_CONFIG_FILENAME
}

func getProjectRootPath(t *testing.T) string {
	_, configsPath, _, _ := runtime.Caller(0)
	rootTestingUtilsDir := filepath.Dir(configsPath)
	rootTestingDir := filepath.Dir(rootTestingUtilsDir)
	rootDir := filepath.Dir(rootTestingDir)

	if rootDir == "" {
		t.Fatal("Unable to find project root path.")
	}

	return rootDir
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
