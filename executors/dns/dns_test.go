package dns

import (
	"context"
	"testing"

	"github.com/ovh/venom"
)

func TestDNSExecutor(t *testing.T) {
	executor := &Executor{}

	// Test with a basic A query
	step := venom.TestStep{
		"server":  "8.8.8.8:53",
		"query":   "example.com",
		"qtype":   "A",
		"timeout": 5,
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.Fail()
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Check that result is of correct type
	if _, ok := result.(Result); !ok {
		t.Fatal("Result should be of type Result")
	}

	res := result.(Result)
	if res.Query != "example.com" {
		t.Errorf("Expected query 'example.com', got '%s'", res.Query)
	}
	if res.QType != "A" {
		t.Errorf("Expected qtype 'A', got '%s'", res.QType)
	}
	if res.Server != "8.8.8.8:53" {
		t.Errorf("Expected server '8.8.8.8:53', got '%s'", res.Server)
	}
}

func TestDNSExecutorDefaults(t *testing.T) {
	executor := &Executor{}

	// Test with minimal configuration (type defaults to A)
	step := venom.TestStep{
		"server": "8.8.8.8:53",
		"query":  "example.com",
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.Fail()
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	res := result.(Result)
	if res.QType != "A" {
		t.Errorf("Expected default qtype 'A', got '%s'", res.QType)
	}
}

func TestDNSExecutorMissingServer(t *testing.T) {
	executor := &Executor{}

	// Test without mandatory server parameter
	step := venom.TestStep{
		"query": "example.com",
		"qtype": "A",
	}

	ctx := context.Background()
	_, err := executor.Run(ctx, step)

	if err == nil {
		t.Fatal("Should return error when server is not provided")
	}

	if err.Error() != "server is mandatory for DNS executor" {
		t.Errorf("Expected error message about mandatory server, got: %v", err)
	}
}

func TestDNSExecutorInvalidType(t *testing.T) {
	executor := &Executor{}

	// Test with invalid DNS type
	step := venom.TestStep{
		"server": "8.8.8.8:53",
		"query":  "example.com",
		"qtype":  "INVALID",
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Fatal("Should not return error for invalid type, should be in result.err")
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	res := result.(Result)
	if res.Err == "" {
		t.Fatal("Result should contain error for invalid DNS type")
	}
}

func TestDNSExecutorMXQuery(t *testing.T) {
	executor := &Executor{}

	// Test with MX query
	step := venom.TestStep{
		"server":  "8.8.8.8:53",
		"query":   "example.com",
		"qtype":   "MX",
		"timeout": 5,
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.Fail()
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	res := result.(Result)
	if res.QType != "MX" {
		t.Errorf("Expected qtype 'MX', got '%s'", res.QType)
	}
	if res.Query != "example.com" {
		t.Errorf("Expected query 'example.com', got '%s'", res.Query)
	}
}

func TestDNSExecutorTXTQuery(t *testing.T) {
	executor := &Executor{}

	// Test with TXT query
	step := venom.TestStep{
		"server":  "8.8.8.8:53",
		"query":   "example.com",
		"qtype":   "TXT",
		"timeout": 5,
	}

	ctx := context.Background()
	result, err := executor.Run(ctx, step)

	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.Fail()
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	res := result.(Result)
	if res.QType != "TXT" {
		t.Errorf("Expected qtype 'TXT', got '%s'", res.QType)
	}
}

func TestDNSExecutorZeroValueResult(t *testing.T) {
	executor := &Executor{}
	result := executor.ZeroValueResult()

	if result == nil {
		t.Fatal("ZeroValueResult should not return nil")
	}

	if _, ok := result.(Result); !ok {
		t.Fatal("ZeroValueResult should return Result type")
	}
}

func TestDNSExecutorDefaultAssertions(t *testing.T) {
	executor := &Executor{}
	assertions := executor.GetDefaultAssertions()

	if assertions == nil {
		t.Fatal("GetDefaultAssertions should not return nil")
	}

	if len(assertions.Assertions) == 0 {
		t.Fatal("GetDefaultAssertions should return at least one assertion")
	}

	expectedAssertion := "result.err ShouldBeEmpty"
	if assertions.Assertions[0] != expectedAssertion {
		t.Errorf("Expected assertion '%s', got '%v'", expectedAssertion, assertions.Assertions[0])
	}
}
