package processor

import "fmt"

// Item represents a data item to process.
type Item struct {
	ID   int
	Name string
	Data string
}

// ProcessResult holds the processing result.
type ProcessResult struct {
	ItemID  int
	Success bool
	Output  string
}

// Processor processes items.
// BUG: Uses fmt.Printf directly - not testable, not replaceable.
type Processor struct {
	prefix string
}

// NewProcessor creates a new Processor.
func NewProcessor(prefix string) *Processor {
	return &Processor{prefix: prefix}
}

// Process processes an item and returns a result.
func (p *Processor) Process(item Item) ProcessResult {
	fmt.Printf("[INFO] Processing item %d: %s\n", item.ID, item.Name)

	if item.Data == "" {
		fmt.Printf("[ERROR] Item %d has empty data\n", item.ID)
		return ProcessResult{ItemID: item.ID, Success: false, Output: "empty data"}
	}

	output := p.prefix + ":" + item.Data
	fmt.Printf("[INFO] Item %d processed successfully: %s\n", item.ID, output)
	return ProcessResult{ItemID: item.ID, Success: true, Output: output}
}

// ProcessBatch processes multiple items.
func (p *Processor) ProcessBatch(items []Item) []ProcessResult {
	fmt.Printf("[INFO] Processing batch of %d items\n", len(items))
	results := make([]ProcessResult, 0, len(items))
	for _, item := range items {
		results = append(results, p.Process(item))
	}
	fmt.Printf("[INFO] Batch processing complete: %d results\n", len(results))
	return results
}
