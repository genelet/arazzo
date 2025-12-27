package generator

import (
	"path/filepath"
	"testing"
)

func TestNewArazzoFromFiles(t *testing.T) {
	tests := []struct {
		name          string
		generatorFile string
		format        []string
		wantErr       bool
	}{
		{
			name:          "YAML Default",
			generatorFile: "generator.yaml",
			format:        nil,
			wantErr:       false,
		},
		{
			name:          "JSON Explicit",
			generatorFile: "generator.json",
			format:        []string{"json"},
			wantErr:       false,
		},
		{
			name:          "HCL Explicit",
			generatorFile: "generator.hcl",
			format:        []string{"hcl"},
			wantErr:       false,
		},
	}

	openapiFile := "openapi.yaml"
	testDir := "testdata"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genPath := filepath.Join(testDir, tt.generatorFile)
			oaPath := filepath.Join(testDir, openapiFile)

			az, err := NewArazzoFromFiles(oaPath, genPath, tt.format...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewArazzoFromFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && az == nil {
				t.Errorf("NewArazzoFromFiles() returned nil result without error")
				return
			}
			if az != nil && len(az.Workflows) == 0 {
				t.Errorf("NewArazzoFromFiles() returned 0 workflows")
			}
			// Basic checks
			if az != nil && len(az.Workflows) > 0 {
				if len(az.Workflows[0].Steps) != 1 {
					t.Errorf("expected 1 step, got %d", len(az.Workflows[0].Steps))
				}
				if az.Workflows[0].Steps[0].StepId != "my-op" {
					t.Errorf("expected StepId 'my-op', got '%s'", az.Workflows[0].Steps[0].StepId)
				}
			}
		})
	}
}
