package flow

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Flow represents an orchestration flow with a fluent DSL
type Flow struct {
	name        string
	description string
	steps       []Step
	middleware  []Middleware
	timeout     time.Duration
}

// NewFlow creates a new flow with the given name
func NewFlow(name string) *Flow {
	return &Flow{
		name:        name,
		description: fmt.Sprintf("Flow: %s", name),
		steps:       make([]Step, 0),
		middleware:  make([]Middleware, 0),
		timeout:     30 * time.Second,
	}
}

// WithDescription sets the flow description
func (f *Flow) WithDescription(description string) *Flow {
	f.description = description
	return f
}

// WithTimeout sets the flow timeout
func (f *Flow) WithTimeout(timeout time.Duration) *Flow {
	f.timeout = timeout
	return f
}

// Use adds middleware to the flow
func (f *Flow) Use(middleware Middleware) *Flow {
	f.middleware = append(f.middleware, middleware)
	return f
}

// Step adds a step to the flow
func (f *Flow) Step(name string, step Step) *Flow {
	// Wrap step with name if it's anonymous
	if step.Name() == "anonymous" {
		step = &namedStep{
			Step: step,
			name: name,
		}
	}
	f.steps = append(f.steps, step)
	return f
}

// StepFunc adds a function as a step
func (f *Flow) StepFunc(name string, fn func(*Context) error) *Flow {
	step := &namedStep{
		Step: StepFunc(fn),
		name: name,
	}
	f.steps = append(f.steps, step)
	return f
}

// Choice starts a conditional branch
func (f *Flow) Choice(name string) *ChoiceBuilder {
	return &ChoiceBuilder{
		flow: f,
		name: name,
	}
}

// Parallel starts a parallel execution block
func (f *Flow) Parallel(name string) *ParallelBuilder {
	return &ParallelBuilder{
		flow: f,
		name: name,
	}
}

// Transform adds a transformation step
func (f *Flow) Transform(name string, transformer func(*Context) error) *Flow {
	return f.Step(name, NewTransformStep(name, transformer))
}

// Delay adds a delay step
func (f *Flow) Delay(name string, duration time.Duration) *Flow {
	return f.Step(name, NewDelayStep(name, duration))
}

// Retry wraps the last step with retry logic
func (f *Flow) Retry(maxRetries int, retryDelay time.Duration) *Flow {
	if len(f.steps) == 0 {
		return f
	}

	lastStep := f.steps[len(f.steps)-1]
	f.steps[len(f.steps)-1] = NewRetryStep(
		fmt.Sprintf("retry_%s", lastStep.Name()),
		lastStep,
		maxRetries,
		retryDelay,
	)
	return f
}

// Execute runs the flow with the given context
func (f *Flow) Execute(ctx *Context) (*ExecutionResult, error) {
	// Set flow name in context
	ctx.WithFlowName(f.name)

	// Apply middleware
	handler := f.executeSteps
	for i := len(f.middleware) - 1; i >= 0; i-- {
		handler = f.middleware[i](handler)
	}

	// Create execution result
	result := &ExecutionResult{
		FlowName:  f.name,
		StartTime: time.Now(),
		Context:   ctx,
	}

	// Execute with timeout
	ctx.WithTimeout(f.timeout)

	ctx.Logger().Info("Starting flow execution",
		zap.String("flow", f.name),
		zap.String("execution_id", ctx.ExecutionID()),
		zap.Int("steps", len(f.steps)))

	// Execute the flow
	err := handler(ctx)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = err == nil
	result.Error = err

	if err != nil {
		ctx.Logger().Error("Flow execution failed",
			zap.String("flow", f.name),
			zap.String("execution_id", ctx.ExecutionID()),
			zap.Duration("duration", result.Duration),
			zap.Error(err))
	} else {
		ctx.Logger().Info("Flow execution completed",
			zap.String("flow", f.name),
			zap.String("execution_id", ctx.ExecutionID()),
			zap.Duration("duration", result.Duration))
	}

	return result, err
}

// executeSteps executes all steps in the flow
func (f *Flow) executeSteps(ctx *Context) error {
	for i, step := range f.steps {
		stepStart := time.Now()

		ctx.Logger().Info("Executing step",
			zap.String("flow", f.name),
			zap.String("step", step.Name()),
			zap.Int("step_index", i),
			zap.String("execution_id", ctx.ExecutionID()))

		err := step.Run(ctx)
		stepDuration := time.Since(stepStart)

		if err != nil {
			ctx.Logger().Error("Step execution failed",
				zap.String("flow", f.name),
				zap.String("step", step.Name()),
				zap.Int("step_index", i),
				zap.Duration("duration", stepDuration),
				zap.Error(err))
			return fmt.Errorf("step '%s' failed: %w", step.Name(), err)
		}

		ctx.Logger().Info("Step execution completed",
			zap.String("flow", f.name),
			zap.String("step", step.Name()),
			zap.Int("step_index", i),
			zap.Duration("duration", stepDuration))
	}

	return nil
}

// Name returns the flow name
func (f *Flow) Name() string {
	return f.name
}

// Description returns the flow description
func (f *Flow) Description() string {
	return f.description
}

// Steps returns the flow steps
func (f *Flow) Steps() []Step {
	return f.steps
}

// ChoiceBuilder builds conditional branches
type ChoiceBuilder struct {
	flow      *Flow
	name      string
	branches  []conditionalBranch
	otherwise Step
}

type conditionalBranch struct {
	condition func(*Context) bool
	steps     []Step
}

// When adds a conditional branch
func (cb *ChoiceBuilder) When(condition func(*Context) bool) *WhenBuilder {
	return &WhenBuilder{
		choiceBuilder: cb,
		condition:     condition,
	}
}

// Otherwise sets the default branch
func (cb *ChoiceBuilder) Otherwise() *OtherwiseBuilder {
	return &OtherwiseBuilder{
		choiceBuilder: cb,
	}
}

// EndChoice completes the choice and returns to the flow
func (cb *ChoiceBuilder) EndChoice() *Flow {
	// Create choice step
	choiceStep := &choiceStep{
		BaseStep:  NewBaseStep(cb.name, fmt.Sprintf("Choice: %s", cb.name)),
		branches:  cb.branches,
		otherwise: cb.otherwise,
	}

	cb.flow.steps = append(cb.flow.steps, choiceStep)
	return cb.flow
}

// WhenBuilder builds a when branch
type WhenBuilder struct {
	choiceBuilder *ChoiceBuilder
	condition     func(*Context) bool
	steps         []Step
}

// Step adds a step to the when branch
func (wb *WhenBuilder) Step(name string, step Step) *WhenBuilder {
	if step.Name() == "anonymous" {
		step = &namedStep{Step: step, name: name}
	}
	wb.steps = append(wb.steps, step)
	return wb
}

// StepFunc adds a function step to the when branch
func (wb *WhenBuilder) StepFunc(name string, fn func(*Context) error) *WhenBuilder {
	step := &namedStep{Step: StepFunc(fn), name: name}
	wb.steps = append(wb.steps, step)
	return wb
}

// When adds another conditional branch
func (wb *WhenBuilder) When(condition func(*Context) bool) *WhenBuilder {
	// Add current branch to choice builder
	wb.choiceBuilder.branches = append(wb.choiceBuilder.branches, conditionalBranch{
		condition: wb.condition,
		steps:     wb.steps,
	})

	// Return new when builder
	return &WhenBuilder{
		choiceBuilder: wb.choiceBuilder,
		condition:     condition,
	}
}

// Otherwise sets the default branch
func (wb *WhenBuilder) Otherwise() *OtherwiseBuilder {
	// Add current branch to choice builder
	wb.choiceBuilder.branches = append(wb.choiceBuilder.branches, conditionalBranch{
		condition: wb.condition,
		steps:     wb.steps,
	})

	return &OtherwiseBuilder{
		choiceBuilder: wb.choiceBuilder,
	}
}

// EndChoice completes the choice
func (wb *WhenBuilder) EndChoice() *Flow {
	// Add current branch to choice builder
	wb.choiceBuilder.branches = append(wb.choiceBuilder.branches, conditionalBranch{
		condition: wb.condition,
		steps:     wb.steps,
	})

	return wb.choiceBuilder.EndChoice()
}

// OtherwiseBuilder builds the otherwise branch
type OtherwiseBuilder struct {
	choiceBuilder *ChoiceBuilder
	steps         []Step
}

// Step adds a step to the otherwise branch
func (ob *OtherwiseBuilder) Step(name string, step Step) *OtherwiseBuilder {
	if step.Name() == "anonymous" {
		step = &namedStep{Step: step, name: name}
	}
	ob.steps = append(ob.steps, step)
	return ob
}

// StepFunc adds a function step to the otherwise branch
func (ob *OtherwiseBuilder) StepFunc(name string, fn func(*Context) error) *OtherwiseBuilder {
	step := &namedStep{Step: StepFunc(fn), name: name}
	ob.steps = append(ob.steps, step)
	return ob
}

// EndChoice completes the choice
func (ob *OtherwiseBuilder) EndChoice() *Flow {
	if len(ob.steps) > 0 {
		ob.choiceBuilder.otherwise = NewSequentialStep("otherwise", ob.steps...)
	}
	return ob.choiceBuilder.EndChoice()
}

// ParallelBuilder builds parallel execution blocks
type ParallelBuilder struct {
	flow  *Flow
	name  string
	steps []Step
}

// Step adds a step to the parallel block
func (pb *ParallelBuilder) Step(name string, step Step) *ParallelBuilder {
	if step.Name() == "anonymous" {
		step = &namedStep{Step: step, name: name}
	}
	pb.steps = append(pb.steps, step)
	return pb
}

// StepFunc adds a function step to the parallel block
func (pb *ParallelBuilder) StepFunc(name string, fn func(*Context) error) *ParallelBuilder {
	step := &namedStep{Step: StepFunc(fn), name: name}
	pb.steps = append(pb.steps, step)
	return pb
}

// EndParallel completes the parallel block
func (pb *ParallelBuilder) EndParallel() *Flow {
	parallelStep := NewParallelStep(pb.name, pb.steps...)
	pb.flow.steps = append(pb.flow.steps, parallelStep)
	return pb.flow
}

// choiceStep implements conditional execution
type choiceStep struct {
	*BaseStep
	branches  []conditionalBranch
	otherwise Step
}

func (cs *choiceStep) Run(ctx *Context) error {
	// Check each branch condition
	for _, branch := range cs.branches {
		if branch.condition(ctx) {
			ctx.Logger().Info("Choice condition met",
				zap.String("choice", cs.Name()))

			// Execute branch steps
			for _, step := range branch.steps {
				if err := step.Run(ctx); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Execute otherwise branch if no condition matched
	if cs.otherwise != nil {
		ctx.Logger().Info("Choice executing otherwise branch",
			zap.String("choice", cs.Name()))
		return cs.otherwise.Run(ctx)
	}

	ctx.Logger().Info("Choice no condition matched, no otherwise branch",
		zap.String("choice", cs.Name()))
	return nil
}

// namedStep wraps a step with a custom name
type namedStep struct {
	Step
	name string
}

func (ns *namedStep) Name() string {
	return ns.name
}

// ExecutionResult represents the result of a flow execution
type ExecutionResult struct {
	FlowName  string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Success   bool
	Error     error
	Context   *Context
}

// GetResponse returns the context data as a response
func (er *ExecutionResult) GetResponse() map[string]interface{} {
	response := map[string]interface{}{
		"flow_name":    er.FlowName,
		"success":      er.Success,
		"duration_ms":  er.Duration.Milliseconds(),
		"execution_id": er.Context.ExecutionID(),
	}

	if er.Error != nil {
		response["error"] = er.Error.Error()
	}

	// Add context data
	response["data"] = er.Context.ToMap()

	return response
}

// Middleware represents flow middleware
type Middleware func(next func(*Context) error) func(*Context) error
