package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnv := func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("SERVER_PORT")
	}

	clearEnv()
	defer clearEnv()

	cfg := Load()

	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost 'localhost', got %q", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("expected DBPort '5432', got %q", cfg.DBPort)
	}
	if cfg.DBUser != "postgres" {
		t.Errorf("expected DBUser 'postgres', got %q", cfg.DBUser)
	}
	if cfg.DBPassword != "postgres" {
		t.Errorf("expected DBPassword 'postgres', got %q", cfg.DBPassword)
	}
	if cfg.DBName != "postgres" {
		t.Errorf("expected DBName 'postgres', got %q", cfg.DBName)
	}
	if cfg.ServerPort != "8080" {
		t.Errorf("expected ServerPort '8080', got %q", cfg.ServerPort)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-password")
	os.Setenv("DB_NAME", "test-db")
	os.Setenv("SERVER_PORT", "9090")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := Load()

	if cfg.DBHost != "test-host" {
		t.Errorf("expected DBHost 'test-host', got %q", cfg.DBHost)
	}
	if cfg.DBPort != "5433" {
		t.Errorf("expected DBPort '5433', got %q", cfg.DBPort)
	}
	if cfg.DBUser != "test-user" {
		t.Errorf("expected DBUser 'test-user', got %q", cfg.DBUser)
	}
	if cfg.DBPassword != "test-password" {
		t.Errorf("expected DBPassword 'test-password', got %q", cfg.DBPassword)
	}
	if cfg.DBName != "test-db" {
		t.Errorf("expected DBName 'test-db', got %q", cfg.DBName)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort '9090', got %q", cfg.ServerPort)
	}
}

func TestLoad_PartialEnvironmentVariables(t *testing.T) {
	// Set only some environment variables
	os.Setenv("DB_HOST", "custom-host")
	os.Setenv("SERVER_PORT", "3000")

	// Clear others to ensure defaults are used
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := Load()

	if cfg.DBHost != "custom-host" {
		t.Errorf("expected DBHost 'custom-host', got %q", cfg.DBHost)
	}
	if cfg.ServerPort != "3000" {
		t.Errorf("expected ServerPort '3000', got %q", cfg.ServerPort)
	}
	// These should use defaults
	if cfg.DBPort != "5432" {
		t.Errorf("expected default DBPort '5432', got %q", cfg.DBPort)
	}
	if cfg.DBUser != "postgres" {
		t.Errorf("expected default DBUser 'postgres', got %q", cfg.DBUser)
	}
}

func TestGetEnv_ExistingKey(t *testing.T) {
	os.Setenv("TEST_KEY", "test-value")
	defer os.Unsetenv("TEST_KEY")

	result := getEnv("TEST_KEY", "default")

	if result != "test-value" {
		t.Errorf("expected 'test-value', got %q", result)
	}
}

func TestGetEnv_NonExistingKey(t *testing.T) {
	os.Unsetenv("NON_EXISTING_KEY")

	result := getEnv("NON_EXISTING_KEY", "default-value")

	if result != "default-value" {
		t.Errorf("expected 'default-value', got %q", result)
	}
}

func TestGetEnv_EmptyValue(t *testing.T) {
	os.Setenv("EMPTY_KEY", "")
	defer os.Unsetenv("EMPTY_KEY")

	result := getEnv("EMPTY_KEY", "default")

	// Empty string is a valid value, should not fall back to default
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestConfig_StructFields(t *testing.T) {
	cfg := Config{
		DBHost:     "host",
		DBPort:     "5432",
		DBUser:     "user",
		DBPassword: "password",
		DBName:     "dbname",
		ServerPort: "8080",
	}

	if cfg.DBHost == "" {
		t.Error("DBHost should not be empty")
	}
	if cfg.DBPort == "" {
		t.Error("DBPort should not be empty")
	}
	if cfg.DBUser == "" {
		t.Error("DBUser should not be empty")
	}
	if cfg.DBPassword == "" {
		t.Error("DBPassword should not be empty")
	}
	if cfg.DBName == "" {
		t.Error("DBName should not be empty")
	}
	if cfg.ServerPort == "" {
		t.Error("ServerPort should not be empty")
	}
}

func TestLoad_MultipleCalls(t *testing.T) {
	os.Setenv("DB_HOST", "host1")
	defer os.Unsetenv("DB_HOST")

	cfg1 := Load()
	cfg2 := Load()

	if cfg1.DBHost != cfg2.DBHost {
		t.Error("multiple calls should return same values")
	}
}

func TestLoad_ConcurrentAccess(t *testing.T) {
	os.Setenv("DB_HOST", "concurrent-host")
	defer os.Unsetenv("DB_HOST")

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			cfg := Load()
			if cfg.DBHost != "concurrent-host" {
				t.Error("concurrent access should return correct value")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	os.Setenv("DB_HOST", "benchmark-host")
	defer os.Unsetenv("DB_HOST")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_KEY", "benchmark-value")
	defer os.Unsetenv("BENCH_KEY")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv("BENCH_KEY", "default")
	}
}
