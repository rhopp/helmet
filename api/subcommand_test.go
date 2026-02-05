package api

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"

	o "github.com/onsi/gomega"
)

// mockSubCommand is a test implementation of SubCommand
type mockSubCommand struct {
	cmd              *cobra.Command
	completeFunc     func([]string) error
	validateFunc     func() error
	runFunc          func() error
	completeCalled   bool
	validateCalled   bool
	runCalled        bool
	completeArgs     []string
}

func newMockSubCommand() *mockSubCommand {
	return &mockSubCommand{
		cmd: &cobra.Command{
			Use:   "mock",
			Short: "Mock command for testing",
		},
		completeFunc: func([]string) error { return nil },
		validateFunc: func() error { return nil },
		runFunc:      func() error { return nil },
	}
}

func (m *mockSubCommand) Cmd() *cobra.Command {
	return m.cmd
}

func (m *mockSubCommand) Complete(args []string) error {
	m.completeCalled = true
	m.completeArgs = args
	return m.completeFunc(args)
}

func (m *mockSubCommand) Validate() error {
	m.validateCalled = true
	return m.validateFunc()
}

func (m *mockSubCommand) Run() error {
	m.runCalled = true
	return m.runFunc()
}

// TestNewRunner tests the Runner constructor
func TestNewRunner(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	runner := NewRunner(mock)

	g.Expect(runner).ToNot(o.BeNil())
	g.Expect(runner.Cmd()).To(o.Equal(mock.cmd))
	g.Expect(mock.cmd.PreRunE).ToNot(o.BeNil())
	g.Expect(mock.cmd.RunE).ToNot(o.BeNil())
}

// TestRunnerCmd tests the Cmd method
func TestRunnerCmd(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	runner := NewRunner(mock)

	cmd := runner.Cmd()
	g.Expect(cmd).To(o.Equal(mock.cmd))
	g.Expect(cmd.Use).To(o.Equal("mock"))
	g.Expect(cmd.Short).To(o.Equal("Mock command for testing"))
}

// TestRunnerWorkflowSuccess tests successful execution of the workflow
func TestRunnerWorkflowSuccess(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	runner := NewRunner(mock)

	// Execute PreRunE (Complete + Validate)
	args := []string{"arg1", "arg2"}
	err := runner.Cmd().PreRunE(nil, args)
	g.Expect(err).To(o.Succeed())
	g.Expect(mock.completeCalled).To(o.BeTrue())
	g.Expect(mock.validateCalled).To(o.BeTrue())
	g.Expect(mock.completeArgs).To(o.Equal(args))

	// Execute RunE (Run)
	err = runner.Cmd().RunE(nil, args)
	g.Expect(err).To(o.Succeed())
	g.Expect(mock.runCalled).To(o.BeTrue())
}

// TestRunnerCompleteError tests error handling in Complete phase
func TestRunnerCompleteError(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	expectedErr := errors.New("complete failed")
	mock.completeFunc = func([]string) error {
		return expectedErr
	}

	runner := NewRunner(mock)

	// PreRunE should fail during Complete
	err := runner.Cmd().PreRunE(nil, []string{})
	g.Expect(err).To(o.HaveOccurred())
	g.Expect(err).To(o.Equal(expectedErr))
	g.Expect(mock.completeCalled).To(o.BeTrue())
	// Validate should not be called if Complete fails
	g.Expect(mock.validateCalled).To(o.BeFalse())
}

// TestRunnerValidateError tests error handling in Validate phase
func TestRunnerValidateError(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	expectedErr := errors.New("validation failed")
	mock.validateFunc = func() error {
		return expectedErr
	}

	runner := NewRunner(mock)

	// PreRunE should fail during Validate
	err := runner.Cmd().PreRunE(nil, []string{})
	g.Expect(err).To(o.HaveOccurred())
	g.Expect(err).To(o.Equal(expectedErr))
	g.Expect(mock.completeCalled).To(o.BeTrue())
	g.Expect(mock.validateCalled).To(o.BeTrue())
}

// TestRunnerRunError tests error handling in Run phase
func TestRunnerRunError(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	expectedErr := errors.New("run failed")
	mock.runFunc = func() error {
		return expectedErr
	}

	runner := NewRunner(mock)

	// PreRunE should succeed
	err := runner.Cmd().PreRunE(nil, []string{})
	g.Expect(err).To(o.Succeed())

	// RunE should fail
	err = runner.Cmd().RunE(nil, []string{})
	g.Expect(err).To(o.HaveOccurred())
	g.Expect(err).To(o.Equal(expectedErr))
	g.Expect(mock.runCalled).To(o.BeTrue())
}

// TestRunnerExecutionOrder tests that methods are called in correct order
func TestRunnerExecutionOrder(t *testing.T) {
	g := o.NewWithT(t)

	executionOrder := []string{}
	mock := newMockSubCommand()

	mock.completeFunc = func([]string) error {
		executionOrder = append(executionOrder, "complete")
		return nil
	}

	mock.validateFunc = func() error {
		executionOrder = append(executionOrder, "validate")
		return nil
	}

	mock.runFunc = func() error {
		executionOrder = append(executionOrder, "run")
		return nil
	}

	runner := NewRunner(mock)

	// Execute workflow
	err := runner.Cmd().PreRunE(nil, []string{})
	g.Expect(err).To(o.Succeed())

	err = runner.Cmd().RunE(nil, []string{})
	g.Expect(err).To(o.Succeed())

	// Verify order: Complete -> Validate -> Run
	g.Expect(executionOrder).To(o.Equal([]string{"complete", "validate", "run"}))
}

// TestRunnerWithEmptyArgs tests execution with empty arguments
func TestRunnerWithEmptyArgs(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	runner := NewRunner(mock)

	err := runner.Cmd().PreRunE(nil, []string{})
	g.Expect(err).To(o.Succeed())
	g.Expect(mock.completeCalled).To(o.BeTrue())
	g.Expect(mock.completeArgs).To(o.BeEmpty())
}

// TestRunnerWithNilArgs tests execution with nil arguments
func TestRunnerWithNilArgs(t *testing.T) {
	g := o.NewWithT(t)

	mock := newMockSubCommand()
	runner := NewRunner(mock)

	err := runner.Cmd().PreRunE(nil, nil)
	g.Expect(err).To(o.Succeed())
	g.Expect(mock.completeCalled).To(o.BeTrue())
	g.Expect(mock.completeArgs).To(o.BeNil())
}
