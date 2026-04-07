package pipeline

// Pipeline represents a data transformation pipeline.
// TODO: Implement type-safe pipeline with:
//   - Stage chaining (Then)
//   - Error propagation
//   - Parallel execution
//   - Retry wrapper

// Stage is a single transformation step.
// PROBLEM: Uses interface{}, no type safety.
type Stage struct {
	name string
	fn   func(interface{}) (interface{}, error)
}

// UnsafePipeline chains stages using interface{}.
// PROBLEM: No compile-time type checking between stages.
type UnsafePipeline struct {
	stages []Stage
}

// NewUnsafePipeline creates a new pipeline.
func NewUnsafePipeline() *UnsafePipeline {
	return &UnsafePipeline{}
}

// AddStage adds a stage to the pipeline.
func (p *UnsafePipeline) AddStage(name string, fn func(interface{}) (interface{}, error)) *UnsafePipeline {
	p.stages = append(p.stages, Stage{name: name, fn: fn})
	return p
}

// Execute runs the pipeline.
// PROBLEM: Type mismatches between stages cause runtime panics.
func (p *UnsafePipeline) Execute(input interface{}) (interface{}, error) {
	var result interface{} = input
	for _, stage := range p.stages {
		var err error
		result, err = stage.fn(result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
