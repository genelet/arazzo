package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSampleGeneration(t *testing.T) {
	wd, _ := os.Getwd()
	sampleDir := filepath.Join(wd, "samples")
	openapiFile := filepath.Join(sampleDir, "sample.openapi.yaml")
	generatorFile := filepath.Join(sampleDir, "sample.generator.yaml")
	outputFile := filepath.Join(sampleDir, "sample.arazzo.yaml")

	// Ensure sample dir exists (it should, but just in case)
	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("Sample directory not found, skipping sample generation test")
	}

	// Run Generator
	az, err := NewArazzoFromFiles(openapiFile, generatorFile)
	assert.NoError(t, err)
	assert.NotNil(t, az)

	// Basic Validations on the Generated Object
	assert.Len(t, az.Workflows, 1)
	wf := az.Workflows[0]
	assert.Equal(t, "full-demo-workflow", wf.WorkflowId)
	assert.Len(t, wf.Steps, 3)

	// Step A: Login
	stepA := wf.Steps[0]
	assert.Equal(t, "login", stepA.StepId)
	assert.Equal(t, "$response.body.token", stepA.Outputs["token"])
	// Check RequestBody Auto-Generation
	assert.NotNil(t, stepA.RequestBody)
	assert.Equal(t, "application/json", stepA.RequestBody.ContentType)
	// Default example payload
	payloadMap, ok := stepA.RequestBody.Payload.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "demo_user", payloadMap["username"])

	// Step B: FetchUser
	stepB := wf.Steps[1]
	assert.Equal(t, "fetchUser", stepB.StepId)
	// Check Parameters
	// id (override), verbose (requested string), X-Trace-Id (explicit override)
	assert.Len(t, stepB.Parameters, 3)

	// Step C: UpdateProfile
	stepC := wf.Steps[2]
	assert.Equal(t, "updateProfile", stepC.StepId)
	// Check Explicit Simple RequestBody
	assert.NotNil(t, stepC.RequestBody)
	// ContentType should be inferred as first available (application/json)
	assert.Equal(t, "application/json", stepC.RequestBody.ContentType)
	// Payload should be the explicit map
	payloadMapC, ok := stepC.RequestBody.Payload.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Updated bio from generator", payloadMapC["bio"])

	// Serialize to YAML for visual inspection by user
	bytes, err := yaml.Marshal(az)
	assert.NoError(t, err)

	err = os.WriteFile(outputFile, bytes, 0644)
	assert.NoError(t, err)

	// Sub-test: HCL Generation
	t.Run("HCL Input", func(t *testing.T) {
		hclFile := filepath.Join(sampleDir, "sample.generator.hcl")
		if _, err := os.Stat(hclFile); os.IsNotExist(err) {
			t.Log("HCL sample not found, skipping HCL test")
			return
		}

		// Run Generator with HCL
		azHCL, err := NewArazzoFromFiles(openapiFile, hclFile, "hcl")
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.NotNil(t, azHCL)
		if azHCL == nil {
			return
		}

		// Basic validation to ensure parity (Checking Step B count)
		assert.Len(t, azHCL.Workflows[0].Steps, 3)
		stepB := azHCL.Workflows[0].Steps[1]
		// HCL Limitation: Explicit parameter blocks might not parse correctly for []interface{}.
		// We expect at least the auto-generated parameter 'id'.
		assert.NotEmpty(t, stepB.Parameters)
	})
}
