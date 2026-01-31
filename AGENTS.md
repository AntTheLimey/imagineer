# Imagineer Agent Registry

## Overview
Agents are focused analysis tools that can be composed into workflows.

## Agent Interface
All agents implement:
```go
type Agent interface {
    Name() string
    Description() string
    Run(ctx context.Context, params map[string]any) (Result, error)
}
```

## Available Agents

### consistency-checker
- **Purpose:** Find plot holes, timeline conflicts, orphaned entities
- **Inputs:** campaign_id, optional entity_type filter
- **Outputs:** List of inconsistencies with severity and suggested fixes

### canon-validator
- **Purpose:** Detect contradictions between sources
- **Inputs:** campaign_id, optional source filter
- **Outputs:** List of conflicts with all conflicting values and sources

### clue-redundancy
- **Purpose:** Ensure critical clues have multiple discovery paths
- **Inputs:** campaign_id, clue_ids (optional)
- **Outputs:** Clues with <3 discovery methods, suggested alternatives

### relationship-mapper
- **Purpose:** Visualize connection networks
- **Inputs:** campaign_id, entity_id (optional for ego network)
- **Outputs:** Graph data (nodes, edges) in mermaid or graphviz format

### session-prep
- **Purpose:** Generate condensed reference sheets
- **Inputs:** campaign_id, session_id
- **Outputs:** Markdown document with relevant NPCs, locations, clues, timeline

### name-checker
- **Purpose:** Flag duplicate or similar names
- **Inputs:** campaign_id, optional new_name to check
- **Outputs:** Similar names with Levenshtein distance

### timeline-validator
- **Purpose:** Check event sequence logic
- **Inputs:** campaign_id, date_range (optional)
- **Outputs:** Timeline inconsistencies, travel time violations

### import-parser
- **Purpose:** Extract entities from unstructured text
- **Inputs:** text content, campaign_id, entity_type hints
- **Outputs:** Proposed entities for confirmation

## Planned Agents
- foreshadowing-tracker
- discovery-state
- style-enforcer
- npc-generator
