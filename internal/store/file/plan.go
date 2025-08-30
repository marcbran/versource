package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type PlanStore struct {
	workDir string
}

func NewPlanStore(workDir string) *PlanStore {
	return &PlanStore{
		workDir: workDir,
	}
}

func (s *PlanStore) StorePlan(ctx context.Context, planID uint, planFilePath string) error {
	plansDir := filepath.Join(s.workDir, "plans")
	err := os.MkdirAll(plansDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create plans directory: %w", err)
	}

	targetPath := filepath.Join(plansDir, fmt.Sprintf("%d.tfplan", planID))

	sourceFile, err := os.Open(planFilePath)
	if err != nil {
		return fmt.Errorf("failed to open source plan file: %w", err)
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target plan file: %w", err)
	}
	defer targetFile.Close()

	_, err = targetFile.ReadFrom(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy plan file: %w", err)
	}

	return nil
}

func (s *PlanStore) LoadPlan(ctx context.Context, planID uint) (string, error) {
	planPath := filepath.Join(s.workDir, "plans", fmt.Sprintf("%d.tfplan", planID))

	_, err := os.Stat(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("plan file not found for plan ID %d", planID)
		}
		return "", fmt.Errorf("failed to stat plan file: %w", err)
	}

	tempFile, err := os.CreateTemp("", "plan-*.tfplan")
	if err != nil {
		return "", fmt.Errorf("failed to create temp plan file: %w", err)
	}
	defer tempFile.Close()

	sourceFile, err := os.Open(planPath)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to open source plan file: %w", err)
	}
	defer sourceFile.Close()

	_, err = tempFile.ReadFrom(sourceFile)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to copy plan file: %w", err)
	}

	return tempFile.Name(), nil
}
