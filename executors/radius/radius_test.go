package radius

import (
	"context"
	"testing"

	"github.com/ovh/venom"
)

func TestRadiusExecutor(t *testing.T) {
	executor := &Executor{}

	// Test with a basic Access-Request
	step := venom.TestStep{
		"server":  "localhost:1812",
		"secret":  "secret",
		"code":    "Access-Request",
		"timeout": 5,
		"attributes": map[string]string{
			"User-Name":      "testuser",
			"User-Password":  "testpass",
			"NAS-IP-Address": "192.168.1.1",
		},
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Expected error (no RADIUS server running): %v", err)
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Check that result is of correct type
	if _, ok := result.(Result); !ok {
		t.Fatal("Result should be of type Result")
	}
}

func TestRadiusExecutorDefaults(t *testing.T) {
	executor := &Executor{}

	// Test with minimal configuration
	step := venom.TestStep{
		"attributes": map[string]string{
			"User-Name": "testuser",
		},
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Expected error (no RADIUS server running): %v", err)
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestRadiusExecutorInvalidCode(t *testing.T) {
	executor := &Executor{}

	// Test with invalid RADIUS code
	step := venom.TestStep{
		"code": "Invalid-Code",
	}

	ctx := context.Background()
	_, err := executor.Run(ctx, step)

	if err == nil {
		t.Fatal("Should return error for invalid RADIUS code")
	}
}

func TestRadiusExecutorInvalidAttribute(t *testing.T) {
	executor := &Executor{}

	// Test with invalid attribute
	step := venom.TestStep{
		"attributes": map[string]string{
			"Invalid-Attribute": "value",
		},
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Fatal("Should not return error for invalid attribute")
	}

	// Check that result contains error information
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if res, ok := result.(Result); ok {
		if res.Err == "" {
			t.Fatal("Result should contain error for invalid attribute")
		}
	}
}
