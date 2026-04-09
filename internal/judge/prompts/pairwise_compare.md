You are comparing two code implementations of the same task. Both attempted the same plan.

## Plan
{{.PlanContent}}

## Implementation A
```diff
{{.DiffA}}
```

## Implementation B
```diff
{{.DiffB}}
```

Compare A and B across these dimensions: correctness, readability, simplicity, robustness, minimality, maintainability.

For each dimension, state which is better (A, B, or Tie).
Then give an overall winner (A, B, or Tie) with reasoning.

Respond with ONLY a JSON object (no markdown code fences, no explanation outside the JSON):
{
  "overall_winner": "A" | "B" | "Tie",
  "reasoning": "...",
  "dimensions": {
    "correctness": "A" | "B" | "Tie",
    "readability": "A" | "B" | "Tie",
    "simplicity": "A" | "B" | "Tie",
    "robustness": "A" | "B" | "Tie",
    "minimality": "A" | "B" | "Tie",
    "maintainability": "A" | "B" | "Tie"
  }
}
