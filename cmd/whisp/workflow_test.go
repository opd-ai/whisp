package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// GitHubWorkflow represents a basic GitHub Actions workflow structure
type GitHubWorkflow struct {
	Name string                 `yaml:"name"`
	On   map[string]interface{} `yaml:"on"`
	Env  map[string]string      `yaml:"env,omitempty"`
	Jobs map[string]Job         `yaml:"jobs"`
}

// Job represents a GitHub Actions job
type Job struct {
	Name        string            `yaml:"name,omitempty"`
	RunsOn      interface{}       `yaml:"runs-on"`         // Can be string or matrix
	Needs       interface{}       `yaml:"needs,omitempty"` // Can be string or array
	Steps       []Step            `yaml:"steps,omitempty"`
	Strategy    *Strategy         `yaml:"strategy,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	If          string            `yaml:"if,omitempty"`
	Permissions map[string]string `yaml:"permissions,omitempty"`
}

// Strategy represents a job strategy matrix
type Strategy struct {
	Matrix map[string]interface{} `yaml:"matrix,omitempty"`
}

// Step represents a GitHub Actions step
type Step struct {
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
}

// TestWorkflowFiles tests that all GitHub Actions workflow files are valid YAML
func TestWorkflowFiles(t *testing.T) {
	workflowDir := "../../../.github/workflows"

	// Check if workflows directory exists
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		t.Skip("Workflows directory does not exist")
	}

	files, err := filepath.Glob(filepath.Join(workflowDir, "*.yml"))
	if err != nil {
		t.Fatalf("Failed to find workflow files: %v", err)
	}

	if len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(workflowDir, "*.yaml"))
		if err != nil {
			t.Fatalf("Failed to find yaml workflow files: %v", err)
		}
	}

	if len(files) == 0 {
		t.Skip("No workflow files found")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			testWorkflowFile(t, file)
		})
	}
}

func testWorkflowFile(t *testing.T, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read workflow file %s: %v", filename, err)
	}

	var workflow GitHubWorkflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("Failed to parse workflow YAML %s: %v", filename, err)
	}

	// Validate basic workflow structure
	if workflow.Name == "" {
		t.Error("Workflow must have a name")
	}

	if len(workflow.Jobs) == 0 {
		t.Error("Workflow must have at least one job")
	}

	// Validate each job
	for jobName, job := range workflow.Jobs {
		t.Run("job_"+jobName, func(t *testing.T) {
			validateJob(t, jobName, job)
		})
	}
}

func validateJob(t *testing.T, jobName string, job Job) {
	// Validate runs-on
	if job.RunsOn == nil {
		t.Errorf("Job %s must specify runs-on", jobName)
	}

	// Validate steps exist for jobs that aren't just dependencies
	if len(job.Steps) == 0 && job.Needs == nil {
		t.Logf("Warning: Job %s has no steps and no dependencies", jobName)
	}

	// Validate steps
	for i, step := range job.Steps {
		validateStep(t, jobName, i, step)
	}
}

func validateStep(t *testing.T, jobName string, stepIndex int, step Step) {
	// A step must have either 'uses' or 'run'
	if step.Uses == "" && step.Run == "" {
		t.Errorf("Job %s step %d must have either 'uses' or 'run'", jobName, stepIndex)
	}

	// Validate common action patterns
	if step.Uses != "" {
		validateActionUsage(t, jobName, stepIndex, step.Uses)
	}
}

func validateActionUsage(t *testing.T, jobName string, stepIndex int, action string) {
	// Validate action format (should be user/repo@version)
	if !strings.Contains(action, "@") {
		t.Errorf("Job %s step %d: Action %s should specify version with @", jobName, stepIndex, action)
	}

	// Check for known good actions
	knownGoodActions := []string{
		"actions/checkout",
		"actions/setup-go",
		"actions/upload-artifact",
		"actions/download-artifact",
		"actions/setup-java",
		"softprops/action-gh-release",
		"peter-evans/create-pull-request",
	}

	actionName := strings.Split(action, "@")[0]
	isKnownGood := false
	for _, known := range knownGoodActions {
		if actionName == known {
			isKnownGood = true
			break
		}
	}

	if !isKnownGood {
		t.Logf("Job %s step %d: Using action %s (not in known good list)", jobName, stepIndex, actionName)
	}
}

// TestWorkflowCoverage tests that we have essential workflows
func TestWorkflowCoverage(t *testing.T) {
	workflowDir := "../../../.github/workflows"

	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		t.Skip("Workflows directory does not exist")
	}

	requiredWorkflows := []string{
		"ci", "build", // Core workflows
	}

	files, _ := filepath.Glob(filepath.Join(workflowDir, "*.yml"))
	yamlFiles, _ := filepath.Glob(filepath.Join(workflowDir, "*.yaml"))
	files = append(files, yamlFiles...)

	workflowNames := make(map[string]bool)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		workflowNames[name] = true
	}

	for _, required := range requiredWorkflows {
		if !workflowNames[required] {
			t.Errorf("Missing required workflow: %s", required)
		}
	}
}
