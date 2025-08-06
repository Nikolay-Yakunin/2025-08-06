package config

import (
	"os"
	"reflect"
	"testing"
)

func TestNewConfig_DefaultValues(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	config := NewConfig()

	if config.Port != ":8080" {
		t.Errorf("Expected Port ':8080', got '%s'", config.Port)
	}
	if config.MaxTasks != 3 {
		t.Errorf("Expected MaxTasks 3, got %d", config.MaxTasks)
	}
	if config.MaxFiles != 3 {
		t.Errorf("Expected MaxFiles 3, got %d", config.MaxFiles)
	}
	if config.MaxFileSize != 300*1024*1024 {
		t.Errorf("Expected MaxFileSize %d, got %d", 300*1024*1024, config.MaxFileSize)
	}
	if config.TmpPath != "/tmp/archiver/" {
		t.Errorf("Expected TmpPath '/tmp/archiver/', got '%s'", config.TmpPath)
	}
	expectedExtensions := []string{".jpg", ".jepg", ".pdf"}
	if !reflect.DeepEqual(config.AllowedExtensions, expectedExtensions) {
		t.Errorf("Expected AllowedExtensions %v, got %v", expectedExtensions, config.AllowedExtensions)
	}
	if config.Mode != "development" {
		t.Errorf("Expected Mode 'development', got '%s'", config.Mode)
	}
}

func TestNewConfig_WithEnvVars(t *testing.T) {
	setEnvOrFatal(t, "PORT", ":9090")
	setEnvOrFatal(t, "MAX_TASKS", "5")
	setEnvOrFatal(t, "MAX_FILES", "10")
	setEnvOrFatal(t, "MAX_FILE_SIZE_MB", "500")
	setEnvOrFatal(t, "TMP_PATH", "/custom/tmp/")
	setEnvOrFatal(t, "ALLOWED_EXT", ".png .gif .txt")
	setEnvOrFatal(t, "MODE", "production")
	
	defer clearEnvVars()

	config := NewConfig()

	if config.Port != ":9090" {
		t.Errorf("Expected Port ':9090', got '%s'", config.Port)
	}
	if config.MaxTasks != 5 {
		t.Errorf("Expected MaxTasks 5, got %d", config.MaxTasks)
	}
	if config.MaxFiles != 10 {
		t.Errorf("Expected MaxFiles 10, got %d", config.MaxFiles)
	}
	if config.MaxFileSize != 500*1024*1024 {
		t.Errorf("Expected MaxFileSize %d, got %d", 500*1024*1024, config.MaxFileSize)
	}
	if config.TmpPath != "/custom/tmp/" {
		t.Errorf("Expected TmpPath '/custom/tmp/', got '%s'", config.TmpPath)
	}
	expectedExtensions := []string{".png", ".gif", ".txt"}
	if !reflect.DeepEqual(config.AllowedExtensions, expectedExtensions) {
		t.Errorf("Expected AllowedExtensions %v, got %v", expectedExtensions, config.AllowedExtensions)
	}
	if config.Mode != "production" {
		t.Errorf("Expected Mode 'production', got '%s'", config.Mode)
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	key := "TEST_KEY"
	value := "test_value"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()

	result := getEnv(key, "default")
	
	if result != value {
		t.Errorf("Expected '%s', got '%s'", value, result)
	}
}

func TestGetEnv_WithEmptyValue(t *testing.T) {
	key := "EMPTY_TEST_KEY"
	setEnvOrFatal(t, key, "")
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := "default_value"

	result := getEnv(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected '%s', got '%s'", defaultValue, result)
	}
}

func TestGetEnv_WithUnsetVariable(t *testing.T) {
	key := "UNSET_TEST_KEY"
	_ = os.Unsetenv(key)
	defaultValue := "default_value"

	result := getEnv(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected '%s', got '%s'", defaultValue, result)
	}
}

func TestParseIntEnv_ValidValue(t *testing.T) {
	key := "INT_TEST_KEY"
	value := "42"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()

	result := parseIntEnv(key, 10)
	
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestParseIntEnv_InvalidValue(t *testing.T) {
	key := "INVALID_INT_TEST_KEY"
	value := "not_a_number"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := 15

	result := parseIntEnv(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseIntEnv_EmptyValue(t *testing.T) {
	key := "EMPTY_INT_TEST_KEY"
	setEnvOrFatal(t, key, "")
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := 20

	result := parseIntEnv(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseIntEnv_UnsetVariable(t *testing.T) {
	key := "UNSET_INT_TEST_KEY"
	_ = os.Unsetenv(key)
	defaultValue := 25

	result := parseIntEnv(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt8Env_ValidValue(t *testing.T) {
	key := "INT8_TEST_KEY"
	value := "100"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()

	result := parseInt8Env(key, 50)
	
	if result != int8(100) {
		t.Errorf("Expected 100, got %d", result)
	}
}

func TestParseInt8Env_InvalidValue(t *testing.T) {
	key := "INVALID_INT8_TEST_KEY"
	value := "not_a_number"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := int8(30)

	result := parseInt8Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt8Env_EmptyValue(t *testing.T) {
	key := "EMPTY_INT8_TEST_KEY"
	setEnvOrFatal(t, key, "")
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := int8(40)

	result := parseInt8Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt8Env_UnsetVariable(t *testing.T) {
	key := "UNSET_INT8_TEST_KEY"
	_ = os.Unsetenv(key)
	defaultValue := int8(60)

	result := parseInt8Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt8Env_OutOfRangeValue(t *testing.T) {
	key := "OUT_OF_RANGE_INT8_TEST_KEY"
	value := "300" // Больше чем int8 max (127)
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := int8(70)

	result := parseInt8Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt64Env_ValidValue(t *testing.T) {
	key := "INT64_TEST_KEY"
	value := "1000000000"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()

	result := parseInt64Env(key, 500000000)
	
	if result != int64(1000000000) {
		t.Errorf("Expected 1000000000, got %d", result)
	}
}

func TestParseInt64Env_InvalidValue(t *testing.T) {
	key := "INVALID_INT64_TEST_KEY"
	value := "not_a_number"
	setEnvOrFatal(t, key, value)
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := int64(1500000000)

	result := parseInt64Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt64Env_EmptyValue(t *testing.T) {
	key := "EMPTY_INT64_TEST_KEY"
	setEnvOrFatal(t, key, "")
	defer func() {
			if err := os.Unsetenv(key); err != nil {
				return
			}
		}()
	defaultValue := int64(2000000000)

	result := parseInt64Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

func TestParseInt64Env_UnsetVariable(t *testing.T) {
	key := "UNSET_INT64_TEST_KEY"
	_ = os.Unsetenv(key)
	defaultValue := int64(2500000000)

	result := parseInt64Env(key, defaultValue)
	
	if result != defaultValue {
		t.Errorf("Expected %d, got %d", defaultValue, result)
	}
}

// Хелперы
func setEnvOrFatal(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Failed to set environment variable %s: %v", key, err)
	}
}

func clearEnvVars() {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("MAX_TASKS")
	_ = os.Unsetenv("MAX_FILES")
	_ = os.Unsetenv("MAX_FILE_SIZE_MB")
	_ = os.Unsetenv("TMP_PATH")
	_ = os.Unsetenv("ALLOWED_EXT")
	_ = os.Unsetenv("MODE")
}