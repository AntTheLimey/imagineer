<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Phase Screens Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Replace the monolithic AnalysisTriagePage
(4,400 lines) with a step-by-step wizard where each
phase (Identify, Revise, Enrich) gets a dedicated
screen. The backend gains failure tracking and LLM
quota detection.

**Architecture:** Three separate page components inside
a shared AnalysisWizard shell. The backend adds a
`failure_reason` column, LLM quota detection in retry
logic, failure reason storage in handlers, and
relationship population for the enrichment pipeline.
Delivery proceeds in five independently shippable
phases: infrastructure first, then one phase screen at
a time.

**Tech Stack:** React 18, TypeScript, MUI v5,
React Router v6, React Query, Go 1.22, PostgreSQL

---

## Phase A: Backend Infrastructure

Phase A delivers the backend changes that the frontend
phases depend on: the database migration, LLM quota
detection, failure reason storage, and relationship
population.

### Task A1: Database Migration for failure_reason

This task adds the `failure_reason TEXT` column to the
`content_analysis_jobs` table and updates the Go model
and database layer to read and write the new column.

**Files:**

- Create: `migrations/005_add_failure_reason.sql`
- Modify: `internal/models/models.go`
- Modify: `internal/database/content_analysis.go`
- Test: `internal/database/content_analysis_test.go`

**Step 1: Write the migration file**

Create the migration file at
`migrations/005_add_failure_reason.sql` with the
following content.

```sql
-- 005_add_failure_reason.sql
-- Adds a failure_reason column to content_analysis_jobs
-- so the frontend can display why a job failed
-- (e.g., quota exceeded, rate limited).

ALTER TABLE content_analysis_jobs
    ADD COLUMN failure_reason TEXT;

COMMENT ON COLUMN content_analysis_jobs.failure_reason
    IS 'Human-readable reason when status is failed '
       '(e.g. API quota exceeded, rate limited).';
```

**Step 2: Add FailureReason to the Go model**

In `internal/models/models.go`, add a `FailureReason`
field to the `ContentAnalysisJob` struct after the
`CurrentPhase` field.

```go
// ContentAnalysisJob represents an analysis run for
// a content field.
type ContentAnalysisJob struct {
	ID                 int64     `json:"id"`
	CampaignID         int64     `json:"campaignId"`
	SourceTable        string    `json:"sourceTable"`
	SourceID           int64     `json:"sourceId"`
	SourceField        string    `json:"sourceField"`
	Status             string    `json:"status"`
	TotalItems         int       `json:"totalItems"`
	ResolvedItems      int       `json:"resolvedItems"`
	EnrichmentTotal    int       `json:"enrichmentTotal"`
	EnrichmentResolved int       `json:"enrichmentResolved"`
	CurrentPhase       *string   `json:"currentPhase"`
	FailureReason      *string   `json:"failureReason,omitempty"`
	Phases             []string  `json:"phases,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
```

**Step 3: Update all database queries**

In `internal/database/content_analysis.go`, update every
query and scan call that touches `content_analysis_jobs`
to include the `failure_reason` column.

Update `CreateAnalysisJob` to include `failure_reason`
in the INSERT and RETURNING clauses, and update the
Scan call.

```go
query := `
    INSERT INTO content_analysis_jobs
        (campaign_id, source_table, source_id,
         source_field, status, total_items,
         resolved_items, enrichment_total,
         enrichment_resolved, current_phase,
         failure_reason)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
            $9, $10, $11)
    RETURNING id, campaign_id, source_table,
              source_id, source_field, status,
              total_items, resolved_items,
              enrichment_total, enrichment_resolved,
              current_phase, failure_reason,
              created_at, updated_at`

var j models.ContentAnalysisJob
err = tx.QueryRow(ctx, query,
    job.CampaignID, job.SourceTable, job.SourceID,
    job.SourceField, job.Status, job.TotalItems,
    job.ResolvedItems, job.EnrichmentTotal,
    job.EnrichmentResolved, currentPhase,
    job.FailureReason,
).Scan(
    &j.ID, &j.CampaignID, &j.SourceTable,
    &j.SourceID, &j.SourceField, &j.Status,
    &j.TotalItems, &j.ResolvedItems,
    &j.EnrichmentTotal, &j.EnrichmentResolved,
    &j.CurrentPhase, &j.FailureReason,
    &j.CreatedAt, &j.UpdatedAt,
)
```

Update `GetAnalysisJob` to SELECT and Scan
`failure_reason`.

```go
query := `
    SELECT id, campaign_id, source_table, source_id,
           source_field, status, total_items,
           resolved_items, enrichment_total,
           enrichment_resolved, current_phase,
           failure_reason, created_at, updated_at
    FROM content_analysis_jobs
    WHERE id = $1`

var j models.ContentAnalysisJob
err := db.QueryRow(ctx, query, id).Scan(
    &j.ID, &j.CampaignID, &j.SourceTable,
    &j.SourceID, &j.SourceField, &j.Status,
    &j.TotalItems, &j.ResolvedItems,
    &j.EnrichmentTotal, &j.EnrichmentResolved,
    &j.CurrentPhase, &j.FailureReason,
    &j.CreatedAt, &j.UpdatedAt,
)
```

Apply the same pattern to
`ListAnalysisJobsByCampaign`,
`GetLatestAnalysisJob`, and `scanAnalysisJobs`.
Each SELECT gains `failure_reason` after
`current_phase`, and each Scan gains
`&j.FailureReason` in the matching position.

**Step 4: Add a SetJobFailureReason helper**

Add a new helper function in
`internal/database/content_analysis.go`.

```go
// SetJobFailureReason updates the failure_reason and
// sets the status to 'failed' for a content analysis
// job.
func (db *DB) SetJobFailureReason(
	ctx context.Context,
	jobID int64,
	reason string,
) error {
	query := `
		UPDATE content_analysis_jobs
		SET status = 'failed',
		    failure_reason = $2
		WHERE id = $1`
	return db.Exec(ctx, query, jobID, reason)
}
```

**Step 5: Apply the migration and verify**

Run the migration against the development database
and verify the column exists.

```bash
psql $DATABASE_URL \
    -f migrations/005_add_failure_reason.sql
```

Verify with the following query.

```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'content_analysis_jobs'
  AND column_name = 'failure_reason';
```

**Commit message:**
`feat: add failure_reason column to content_analysis_jobs`

---

### Task A2: LLM Quota Detection in Retry Logic

This task updates the `doWithRetry` function to
distinguish API quota exhaustion from rate limiting.
The function fails immediately on quota errors
instead of retrying.

**Files:**

- Modify: `internal/llm/retry.go`
- Modify: `internal/llm/client.go`
- Test: `internal/llm/client_test.go`

**Step 1: Write the failing tests**

Add these test functions to
`internal/llm/client_test.go`.

```go
func TestRetryOn402_FailsImmediately(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter,
				r *http.Request) {
				attempts++
				w.WriteHeader(
					http.StatusPaymentRequired)
				_, _ = w.Write([]byte(
					`{"error":{"message":` +
						`"quota exceeded"}}`))
			}))
	defer server.Close()

	provider := &AnthropicProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	_, err := provider.Complete(
		context.Background(),
		CompletionRequest{UserPrompt: "test"},
	)
	if err == nil {
		t.Fatal("expected error for 402")
	}

	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Errorf(
			"expected QuotaExceededError, got %T: %v",
			err, err)
	}

	if attempts != 1 {
		t.Errorf(
			"expected 1 attempt (no retry), got %d",
			attempts)
	}
}

func TestRetryOn429Quota_FailsImmediately(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter,
				r *http.Request) {
				attempts++
				w.WriteHeader(
					http.StatusTooManyRequests)
				_, _ = w.Write([]byte(
					`{"error":{"message":` +
						`"You exceeded your ` +
						`current quota"}}`))
			}))
	defer server.Close()

	provider := &OpenAIProvider{
		apiKey: "test-key",
		client: server.Client(),
	}
	// Override the URL for testing.
	origURL := openaiAPIURL
	defer func() { _ = origURL }()

	// Use AnthropicProvider to test the retry
	// logic directly via doWithRetry.
	_, err := doWithRetry(
		context.Background(),
		func(ctx context.Context) (
			CompletionResponse, int, error,
		) {
			attempts := 0
			_ = attempts
			return CompletionResponse{},
				429,
				fmt.Errorf(
					"openai API error (status 429)" +
						": You exceeded your " +
						"current quota")
		},
	)
	if err == nil {
		t.Fatal("expected error for 429 quota")
	}

	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Errorf(
			"expected QuotaExceededError, got %T: %v",
			err, err)
	}
}
```

**Step 2: Run the tests to verify failure**

```bash
go test ./internal/llm/... \
    -run "TestRetryOn402|TestRetryOn429Quota" -v
```

Expected: FAIL (QuotaExceededError does not exist).

**Step 3: Add the QuotaExceededError type**

Add the following to `internal/llm/client.go` after
the `CompletionResponse` struct.

```go
// QuotaExceededError indicates the LLM provider has
// rejected the request due to an exhausted API quota
// (as opposed to a temporary rate limit).
type QuotaExceededError struct {
	Provider string
	Message  string
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf(
		"%s quota exceeded: %s",
		e.Provider, e.Message,
	)
}
```

**Step 4: Update doWithRetry to detect quota errors**

Replace the `doWithRetry` function in
`internal/llm/retry.go` with the following
implementation.

```go
// isQuotaError checks whether the HTTP status code
// and error message indicate an API quota exhaustion
// rather than a temporary rate limit.
func isQuotaError(statusCode int, err error) bool {
	// Anthropic returns 402 for quota exceeded.
	if statusCode == 402 {
		return true
	}
	// OpenAI returns 429 with "exceeded your current
	// quota" in the message body.
	if statusCode == 429 && err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "quota") ||
			strings.Contains(
				msg, "exceeded") {
			return true
		}
	}
	return false
}

// doWithRetry calls fn up to maxRetries times with
// exponential backoff when the HTTP status code is 429
// (rate limited) or 503 (service unavailable). Quota
// errors (402 or 429-with-quota-body) fail immediately
// without retrying.
func doWithRetry(
	ctx context.Context,
	fn func(ctx context.Context) (
		CompletionResponse, int, error),
) (CompletionResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, statusCode, err := fn(ctx)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Quota errors fail immediately.
		if isQuotaError(statusCode, err) {
			return CompletionResponse{},
				&QuotaExceededError{
					Provider: "llm",
					Message:  err.Error(),
				}
		}

		// Only retry on 429 (rate limited) or
		// 503 (service unavailable).
		if statusCode != 429 && statusCode != 503 {
			return CompletionResponse{}, err
		}

		if attempt < maxRetries {
			backoff := time.Duration(
				math.Pow(2,
					float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return CompletionResponse{},
					ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return CompletionResponse{}, lastErr
}
```

Add `"strings"` to the import block in
`internal/llm/retry.go`.

**Step 5: Run all LLM tests to verify**

```bash
go test ./internal/llm/... -v
```

Expected: all tests pass, including the new quota
detection tests and the existing `TestRetryOn429`
retry test.

**Commit message:**
`feat: detect LLM quota errors and fail immediately`

---

### Task A3: Store failure_reason in Handlers

This task updates the content analysis handler to
store a `failure_reason` when enrichment fails,
using the new `SetJobFailureReason` helper and
detecting `QuotaExceededError`.

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go`

**Step 1: Update RunContentEnrichment error handling**

In `RunContentEnrichment` (around line 1588),
replace the pipeline failure block with logic that
detects quota errors and stores the reason.

```go
enrichItems, err := pipeline.Run(
	bgCtx, provider, input)
if err != nil {
	log.Printf(
		"Content-enrich: pipeline run failed "+
			"for job %d: %v",
		jobID, err)

	reason := "Enrichment encountered an error"
	var qe *llm.QuotaExceededError
	if errors.As(err, &qe) {
		reason = "API quota exceeded"
	} else if strings.Contains(
		err.Error(), "rate limit") {
		reason = "Rate limited after retries"
	}

	_ = h.db.SetJobFailureReason(
		context.Background(), jobID, reason)
	return
}
```

**Step 2: Update TryAutoEnrich error handling**

Apply the same pattern in `TryAutoEnrich` (around
line 1777), replacing the existing failure block.

```go
enrichItems, err := pipeline.Run(
	bgCtx, provider, input)
if err != nil {
	log.Printf(
		"Auto-enrich: pipeline run failed "+
			"for job %d: %v",
		jobID, err)

	reason := "Enrichment encountered an error"
	var qe *llm.QuotaExceededError
	if errors.As(err, &qe) {
		reason = "API quota exceeded"
	} else if strings.Contains(
		err.Error(), "rate limit") {
		reason = "Rate limited after retries"
	}

	_ = h.db.SetJobFailureReason(
		context.Background(), jobID, reason)
	return
}
```

**Step 3: Add missing imports**

Add `"github.com/antonypegg/imagineer/internal/llm"`
to the import block of the handler file if the import
is not already present.

**Step 4: Run the full test suite to verify**

```bash
go test ./internal/api/... -v
```

**Commit message:**
`feat: store failure_reason on enrichment errors`

---

### Task A4: Populate PipelineInput.Relationships

This task fixes the bug where `PipelineInput` never
receives relationship data. Both `RunContentEnrichment`
and `TryAutoEnrich` now load relationships from the
database and pass them to the pipeline.

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go`

**Step 1: Load relationships in RunContentEnrichment**

In `RunContentEnrichment`, after the line that
fetches entities (or after building the `input`
struct around line 1577), add a call to load
relationships.

```go
// Load relationships for the enrichment
// pipeline so the graph expert receives
// complete data.
relationships, relErr :=
	h.db.ListRelationshipsByCampaign(
		bgCtx, campaignID)
if relErr != nil {
	log.Printf(
		"Content-enrich: failed to load "+
			"relationships for job %d: %v",
		jobID, relErr)
	// Continue without relationships rather
	// than blocking enrichment.
}
```

Then set the `Relationships` field on the `input`
struct.

```go
input := enrichment.PipelineInput{
	CampaignID:    campaignID,
	JobID:         jobID,
	SourceTable:   job.SourceTable,
	SourceID:      job.SourceID,
	SourceScope:   enrichment.ScopeFromSourceTable(
		job.SourceTable),
	Content:       content,
	Relationships: relationships,
	GameSystemID:  gameSystemID,
	Context:       ragCtx,
}
```

**Step 2: Load relationships in TryAutoEnrich**

Apply the same pattern in `TryAutoEnrich`, after the
entity loading block (around line 1765).

```go
relationships, relErr :=
	h.db.ListRelationshipsByCampaign(
		bgCtx, job.CampaignID)
if relErr != nil {
	log.Printf(
		"Auto-enrich: failed to load "+
			"relationships for job %d: %v",
		jobID, relErr)
}

input := enrichment.PipelineInput{
	CampaignID:    job.CampaignID,
	JobID:         jobID,
	SourceTable:   job.SourceTable,
	SourceID:      job.SourceID,
	SourceScope:   enrichment.ScopeFromSourceTable(
		job.SourceTable),
	Content:       content,
	Entities:      entities,
	Relationships: relationships,
	GameSystemID:  gameSystemID,
	Context:       ragCtx,
}
```

**Step 3: Run the full test suite to verify**

```bash
go test ./internal/api/... -v
```

**Commit message:**
`fix: populate PipelineInput.Relationships for enrichment`

---

### Task A5: Frontend Type Updates

This task updates the frontend TypeScript types and
API client to include `failureReason` on the job
interface.

**Files:**

- Modify: `client/src/api/contentAnalysis.ts`
- Test: `client/src/api/contentAnalysis.test.ts`

**Step 1: Add failureReason to the interface**

In `client/src/api/contentAnalysis.ts`, add
`failureReason` to the `ContentAnalysisJob`
interface after the `currentPhase` field.

```typescript
export interface ContentAnalysisJob {
    id: number;
    campaignId: number;
    sourceTable: string;
    sourceId: number;
    sourceField: string;
    status: string;
    totalItems: number;
    resolvedItems: number;
    enrichmentTotal: number;
    enrichmentResolved: number;
    phases: string[];
    currentPhase: string | null;
    failureReason?: string;
    createdAt: string;
    updatedAt: string;
}
```

**Step 2: Write a type assertion test**

Create or update
`client/src/api/contentAnalysis.test.ts` with a
type assertion verifying the field exists.

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import type {
    ContentAnalysisJob,
} from './contentAnalysis';

describe('ContentAnalysisJob type', () => {
    it('includes failureReason field', () => {
        const job: ContentAnalysisJob = {
            id: 1,
            campaignId: 1,
            sourceTable: 'chapters',
            sourceId: 1,
            sourceField: 'overview',
            status: 'failed',
            totalItems: 5,
            resolvedItems: 0,
            enrichmentTotal: 0,
            enrichmentResolved: 0,
            phases: ['identification'],
            currentPhase: 'identification',
            failureReason: 'API quota exceeded',
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
        };
        expect(job.failureReason).toBe(
            'API quota exceeded',
        );
    });

    it('allows undefined failureReason', () => {
        const job: ContentAnalysisJob = {
            id: 1,
            campaignId: 1,
            sourceTable: 'chapters',
            sourceId: 1,
            sourceField: 'overview',
            status: 'completed',
            totalItems: 5,
            resolvedItems: 5,
            enrichmentTotal: 0,
            enrichmentResolved: 0,
            phases: ['identification'],
            currentPhase: null,
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
        };
        expect(job.failureReason).toBeUndefined();
    });
});
```

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/api/contentAnalysis.test.ts
```

**Commit message:**
`feat: add failureReason to frontend ContentAnalysisJob type`

---

## Phase B: Wizard Shell and Routing

Phase B delivers the AnalysisWizard shell component,
nested routing, the `useAnalysisWizard` hook, context
provider, phase stepper, auto-advance logic, error
banner, and navigation buttons. After Phase B ships,
the wizard shell renders but the phase pages are
placeholder components.

### Task B1: useAnalysisWizard Hook

This task creates the custom hook that manages wizard
state: job data, items grouped by phase, navigation,
and auto-advancement.

**Files:**

- Create: `client/src/hooks/useAnalysisWizard.ts`
- Test: `client/src/hooks/useAnalysisWizard.test.ts`

**Step 1: Write the test**

Create
`client/src/hooks/useAnalysisWizard.test.ts`
with tests for phase computation and filtering.

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import {
    getPhaseItems,
    DETECTION_GROUPS,
    ANALYSIS_GROUPS,
    ENRICHMENT_GROUPS,
} from './useAnalysisWizard';
import type {
    ContentAnalysisItem,
} from '../api/contentAnalysis';

function makeItem(
    detectionType: string,
    phase: string,
    resolution = 'pending',
): ContentAnalysisItem {
    return {
        id: Math.random() * 1000,
        jobId: 1,
        detectionType,
        matchedText: 'test',
        resolution,
        phase,
        createdAt: '2026-01-01T00:00:00Z',
    } as ContentAnalysisItem;
}

describe('getPhaseItems', () => {
    it('filters identification items', () => {
        const items = [
            makeItem(
                'wiki_link_resolved',
                'identification',
            ),
            makeItem(
                'untagged_mention',
                'identification',
            ),
            makeItem(
                'description_update',
                'enrichment',
            ),
            makeItem(
                'analysis_report',
                'analysis',
            ),
        ];
        const result = getPhaseItems(
            items, 'identification',
        );
        expect(result).toHaveLength(2);
        expect(
            result.every(
                (i) =>
                    DETECTION_GROUPS.includes(
                        i.detectionType,
                    ),
            ),
        ).toBe(true);
    });

    it('filters analysis items', () => {
        const items = [
            makeItem(
                'analysis_report', 'analysis',
            ),
            makeItem(
                'canon_contradiction', 'analysis',
            ),
            makeItem(
                'wiki_link_resolved',
                'identification',
            ),
        ];
        const result = getPhaseItems(
            items, 'analysis',
        );
        expect(result).toHaveLength(2);
        expect(
            result.every(
                (i) =>
                    ANALYSIS_GROUPS.includes(
                        i.detectionType,
                    ),
            ),
        ).toBe(true);
    });

    it('filters enrichment items', () => {
        const items = [
            makeItem(
                'description_update', 'enrichment',
            ),
            makeItem(
                'log_entry', 'enrichment',
            ),
            makeItem(
                'analysis_report', 'analysis',
            ),
        ];
        const result = getPhaseItems(
            items, 'enrichment',
        );
        expect(result).toHaveLength(2);
        expect(
            result.every(
                (i) =>
                    ENRICHMENT_GROUPS.includes(
                        i.detectionType,
                    ),
            ),
        ).toBe(true);
    });
});
```

**Step 2: Implement the hook**

Create `client/src/hooks/useAnalysisWizard.ts`
with the following key exports and logic.

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import {
    useMemo,
    useCallback,
    useState,
    useEffect,
} from 'react';
import { useNavigate } from 'react-router-dom';
import {
    useAnalysisJob,
    useAnalysisItems,
} from './useContentAnalysis';
import type {
    ContentAnalysisItem,
} from '../api/contentAnalysis';

/** Detection types for the Identify phase. */
export const DETECTION_GROUPS = [
    'wiki_link_resolved',
    'wiki_link_unresolved',
    'untagged_mention',
    'potential_alias',
    'misspelling',
] as const;

/** Detection types for the Revise phase. */
export const ANALYSIS_GROUPS = [
    'analysis_report',
    'content_suggestion',
    'mechanics_warning',
    'investigation_gap',
    'pacing_note',
    'canon_contradiction',
    'temporal_inconsistency',
    'character_inconsistency',
] as const;

/** Detection types for the Enrich phase. */
export const ENRICHMENT_GROUPS = [
    'description_update',
    'log_entry',
    'relationship_suggestion',
    'new_entity_suggestion',
    'graph_warning',
    'redundant_edge',
    'invalid_type_pair',
    'orphan_warning',
] as const;

/** Phase key to route segment mapping. */
export const PHASE_ROUTES: Record<
    string, string
> = {
    identification: 'identify',
    analysis: 'revise',
    enrichment: 'enrich',
};

/**
 * Filter items by phase using detection type
 * group membership.
 */
export function getPhaseItems(
    items: ContentAnalysisItem[],
    phase: string,
): ContentAnalysisItem[] {
    let types: readonly string[];
    switch (phase) {
        case 'identification':
            types = DETECTION_GROUPS;
            break;
        case 'analysis':
            types = ANALYSIS_GROUPS;
            break;
        case 'enrichment':
            types = ENRICHMENT_GROUPS;
            break;
        default:
            return [];
    }
    return items.filter((item) =>
        types.includes(item.detectionType),
    );
}

/** Return type for the useAnalysisWizard hook. */
export interface AnalysisWizardState {
    // -- data --
    job: ReturnType<
        typeof useAnalysisJob
    >['data'];
    items: ContentAnalysisItem[];
    isLoading: boolean;
    error: Error | null;

    // -- phase navigation --
    phases: string[];
    currentPhase: string | null;
    currentPhaseIndex: number;
    phaseItems: ContentAnalysisItem[];
    pendingCount: number;
    canAdvance: boolean;
    canGoBack: boolean;
    nextPhaseLabel: string | null;

    // -- actions --
    goToPhase: (phase: string) => void;
    goToNextPhase: () => void;
}

/**
 * The useAnalysisWizard hook provides job data,
 * items, phase navigation state, and auto-advance
 * logic for the AnalysisWizard shell.
 */
export function useAnalysisWizard(
    campaignId: number,
    jobId: number,
    currentRoutePhase?: string,
): AnalysisWizardState {
    const navigate = useNavigate();
    const jobQuery = useAnalysisJob(
        campaignId, jobId,
    );
    const itemsQuery = useAnalysisItems(
        campaignId, jobId,
    );
    const [autoAdvanced, setAutoAdvanced] =
        useState(false);

    const job = jobQuery.data;
    const items = itemsQuery.data ?? [];
    const phases = job?.phases ?? [];

    // Map the route segment back to the phase key.
    const routeToPhase: Record<string, string> = {
        identify: 'identification',
        revise: 'analysis',
        enrich: 'enrichment',
    };

    const currentPhase = currentRoutePhase
        ? routeToPhase[currentRoutePhase] ??
          currentRoutePhase
        : job?.currentPhase ?? phases[0] ?? null;

    const currentPhaseIndex = currentPhase
        ? phases.indexOf(currentPhase)
        : -1;

    const phaseItems = useMemo(
        () =>
            currentPhase
                ? getPhaseItems(items, currentPhase)
                : [],
        [items, currentPhase],
    );

    const pendingCount = useMemo(
        () =>
            phaseItems.filter(
                (i) => i.resolution === 'pending',
            ).length,
        [phaseItems],
    );

    const canAdvance =
        currentPhaseIndex < phases.length - 1;
    const canGoBack = currentPhaseIndex > 0;

    const nextPhaseLabel = canAdvance
        ? PHASE_ROUTES[phases[
              currentPhaseIndex + 1
          ]] ?? null
        : null;

    const goToPhase = useCallback(
        (phase: string) => {
            const segment =
                PHASE_ROUTES[phase] ?? phase;
            navigate(
                `/campaigns/${campaignId}` +
                    `/analysis/${jobId}` +
                    `/${segment}`,
            );
        },
        [navigate, campaignId, jobId],
    );

    const goToNextPhase = useCallback(() => {
        if (canAdvance) {
            goToPhase(
                phases[currentPhaseIndex + 1],
            );
        }
    }, [
        canAdvance,
        goToPhase,
        phases,
        currentPhaseIndex,
    ]);

    // Auto-advance when all items are resolved.
    useEffect(() => {
        if (
            pendingCount === 0 &&
            phaseItems.length > 0 &&
            canAdvance &&
            !autoAdvanced
        ) {
            setAutoAdvanced(true);
            const timer = setTimeout(() => {
                goToNextPhase();
                setAutoAdvanced(false);
            }, 1500);
            return () => clearTimeout(timer);
        }
        return undefined;
    }, [
        pendingCount,
        phaseItems.length,
        canAdvance,
        autoAdvanced,
        goToNextPhase,
    ]);

    return {
        job,
        items,
        isLoading:
            jobQuery.isLoading ||
            itemsQuery.isLoading,
        error:
            jobQuery.error ?? itemsQuery.error,
        phases,
        currentPhase,
        currentPhaseIndex,
        phaseItems,
        pendingCount,
        canAdvance,
        canGoBack,
        nextPhaseLabel,
        goToPhase,
        goToNextPhase,
    };
}
```

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/useAnalysisWizard.test.ts
```

**Commit message:**
`feat: add useAnalysisWizard hook with phase filtering`

---

### Task B2: AnalysisWizardContext

This task creates a React context that the wizard
shell provides and the phase pages consume. The
context exposes the wizard state so phase pages do
not need to call the hook directly.

**Files:**

- Create:
  `client/src/contexts/AnalysisWizardContext.tsx`

**Step 1: Create the context**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import {
    createContext,
    useContext,
} from 'react';
import type {
    AnalysisWizardState,
} from '../hooks/useAnalysisWizard';

const AnalysisWizardContext = createContext<
    AnalysisWizardState | null
>(null);

export function AnalysisWizardProvider({
    value,
    children,
}: {
    value: AnalysisWizardState;
    children: React.ReactNode;
}) {
    return (
        <AnalysisWizardContext.Provider
            value={value}
        >
            {children}
        </AnalysisWizardContext.Provider>
    );
}

export function useWizardContext():
    AnalysisWizardState {
    const ctx = useContext(
        AnalysisWizardContext,
    );
    if (!ctx) {
        throw new Error(
            'useWizardContext must be used ' +
                'within AnalysisWizardProvider',
        );
    }
    return ctx;
}

export default AnalysisWizardContext;
```

**Commit message:**
`feat: add AnalysisWizardContext for phase pages`

---

### Task B3: AnalysisWizard Shell Component

This task creates the AnalysisWizard shell component
with the MUI Stepper, error banner, navigation
buttons, and nested route outlet.

**Files:**

- Create: `client/src/pages/AnalysisWizard.tsx`
- Test: `client/src/pages/AnalysisWizard.test.tsx`

**Step 1: Write a smoke test**

Create `client/src/pages/AnalysisWizard.test.tsx`
to verify the shell renders a stepper and an outlet.

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from
    '@testing-library/react';
import { MemoryRouter, Route, Routes } from
    'react-router-dom';
import { QueryClient, QueryClientProvider } from
    '@tanstack/react-query';

// Minimal smoke test: verify the wizard shell
// renders without crashing when given a valid
// route. Full integration tests run in Phases
// C-E.

describe('AnalysisWizard', () => {
    it('renders loading state', async () => {
        // The component loads job data via React
        // Query, so it starts in a loading state.
        // A full test requires mocking the API;
        // this test verifies the component mounts
        // without errors.
        const queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false },
            },
        });

        const { container } = render(
            <QueryClientProvider
                client={queryClient}
            >
                <MemoryRouter
                    initialEntries={[
                        '/campaigns/1/analysis' +
                            '/1/identify',
                    ]}
                >
                    <Routes>
                        <Route
                            path={
                                '/campaigns' +
                                '/:campaignId' +
                                '/analysis' +
                                '/:jobId/*'
                            }
                            element={
                                <div>
                                    Wizard Shell
                                </div>
                            }
                        />
                    </Routes>
                </MemoryRouter>
            </QueryClientProvider>,
        );
        expect(
            container.textContent,
        ).toContain('Wizard Shell');
    });
});
```

**Step 2: Create the AnalysisWizard component**

Create `client/src/pages/AnalysisWizard.tsx` with
the following structure. The component uses the
`useAnalysisWizard` hook, wraps children in
`AnalysisWizardProvider`, renders a Stepper, an
error Alert, and navigation buttons.

Key structural elements:

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { useParams, Outlet, useLocation } from
    'react-router-dom';
import {
    Box,
    Stepper,
    Step,
    StepLabel,
    Button,
    Alert,
    Snackbar,
    Typography,
    CircularProgress,
} from '@mui/material';
import { useState } from 'react';
import {
    useAnalysisWizard,
    PHASE_ROUTES,
} from '../hooks/useAnalysisWizard';
import {
    AnalysisWizardProvider,
} from '../contexts/AnalysisWizardContext';

/** Map phase keys to display labels. */
const PHASE_LABELS: Record<string, string> = {
    identification: 'Identify',
    analysis: 'Revise',
    enrichment: 'Enrich',
};

export default function AnalysisWizard() {
    const { campaignId, jobId } = useParams<{
        campaignId: string;
        jobId: string;
    }>();
    const location = useLocation();

    // Extract the current route phase segment
    // from the URL pathname.
    const pathSegments =
        location.pathname.split('/');
    const currentRoutePhase =
        pathSegments[pathSegments.length - 1];

    const cId = Number(campaignId);
    const jId = Number(jobId);

    const wizard = useAnalysisWizard(
        cId, jId, currentRoutePhase,
    );

    const [snackOpen, setSnackOpen] =
        useState(false);

    // ... Component renders:
    // 1. Loading spinner when wizard.isLoading
    // 2. Header with job title / source info
    // 3. MUI Stepper with wizard.phases
    // 4. Error Alert when job.status === 'failed'
    //    and job.failureReason is set
    // 5. AnalysisWizardProvider wrapping Outlet
    // 6. Navigation buttons (Back to Overview,
    //    Continue to [next phase])
    // 7. Auto-advance Snackbar

    // Error banner logic:
    // if (wizard.job?.status === 'failed' &&
    //     wizard.job?.failureReason) {
    //     render Alert with the failure reason
    // }

    // The Outlet renders the active phase page
    // (IdentifyPhasePage, RevisePhasePage, or
    // EnrichPhasePage) based on nested routing.

    return (
        <AnalysisWizardProvider value={wizard}>
            <Box sx={{ p: 3, maxWidth: 1400,
                       mx: 'auto' }}>
                {/* Header */}
                {/* Stepper */}
                {/* Error banner */}
                {/* Phase content via Outlet */}
                <Outlet />
                {/* Navigation */}
            </Box>
        </AnalysisWizardProvider>
    );
}
```

The implementer builds out the full JSX using MUI
components. The critical elements are the Stepper
mapping over `wizard.phases`, the error Alert
checking `wizard.job?.failureReason`, and the
`AnalysisWizardProvider` wrapping the `Outlet`.

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/AnalysisWizard.test.tsx
```

**Commit message:**
`feat: add AnalysisWizard shell with stepper and error banner`

---

### Task B4: Update Routing in App.tsx

This task replaces the single AnalysisTriagePage
route with nested routes under the AnalysisWizard
shell. The base route redirects to the first active
phase. Placeholder components stand in for the phase
pages until Phases C-E deliver them.

**Files:**

- Modify: `client/src/App.tsx`

**Step 1: Update the imports**

Replace the `AnalysisTriagePage` import with imports
for `AnalysisWizard` and placeholder phase pages.

```typescript
import AnalysisWizard from
    './pages/AnalysisWizard';
// Placeholder phase pages (replaced in Phases
// C-E). Use a simple component that reads from
// context and renders placeholder text.
import IdentifyPhasePage from
    './pages/IdentifyPhasePage';
import RevisePhasePage from
    './pages/RevisePhasePage';
import EnrichPhasePage from
    './pages/EnrichPhasePage';
```

**Step 2: Replace the route definition**

In the `FullScreenWrapper` route group, replace
the single AnalysisTriagePage route with nested
routes.

```tsx
{/* Analysis wizard with phase screens */}
<Route
    path={
        '/campaigns/:campaignId' +
        '/analysis/:jobId'
    }
    element={<AnalysisWizard />}
>
    <Route
        index
        element={
            <Navigate
                to="identify"
                replace
            />
        }
    />
    <Route
        path="identify"
        element={
            <IdentifyPhasePage />
        }
    />
    <Route
        path="revise"
        element={
            <RevisePhasePage />
        }
    />
    <Route
        path="enrich"
        element={
            <EnrichPhasePage />
        }
    />
</Route>
```

**Step 3: Create placeholder phase pages**

Create minimal placeholder components for each
phase page. Each placeholder imports the wizard
context and renders a "Coming soon" message with
the item count. These placeholders ship with
Phase B and are replaced in Phases C-E.

For example,
`client/src/pages/IdentifyPhasePage.tsx`:

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { Box, Typography } from '@mui/material';
import {
    useWizardContext,
} from '../contexts/AnalysisWizardContext';

export default function IdentifyPhasePage() {
    const { phaseItems, pendingCount } =
        useWizardContext();

    return (
        <Box sx={{ p: 2 }}>
            <Typography variant="h5">
                Identify Phase
            </Typography>
            <Typography>
                {phaseItems.length} items
                ({pendingCount} pending)
            </Typography>
        </Box>
    );
}
```

Create similar files for `RevisePhasePage.tsx` and
`EnrichPhasePage.tsx`, changing the title and phase
name accordingly.

**Step 4: Run the test suite**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run
```

**Commit message:**
`feat: add nested wizard routing with placeholder phase pages`

---

## Phase C: Identify Phase Page

Phase C delivers the IdentifyPhasePage component,
extracting the identification UI from the monolithic
AnalysisTriagePage. The page displays items grouped
by detection type, a detail panel, and supports
Accept, Reassign, Create Entity, Dismiss, and
Accept All actions.

### Task C1: IdentifyPhasePage Core UI

This task replaces the placeholder IdentifyPhasePage
with the full identification UI: a left panel with
grouped items and a right panel with detail view.

**Files:**

- Modify: `client/src/pages/IdentifyPhasePage.tsx`
- Test:
  `client/src/pages/IdentifyPhasePage.test.tsx`

**Step 1: Write component tests**

Create
`client/src/pages/IdentifyPhasePage.test.tsx`
with tests that verify item grouping and rendering.

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from
    '@testing-library/react';
import { QueryClient, QueryClientProvider } from
    '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { AnalysisWizardProvider } from
    '../contexts/AnalysisWizardContext';
import type { AnalysisWizardState } from
    '../hooks/useAnalysisWizard';
import type { ContentAnalysisItem } from
    '../api/contentAnalysis';

// The test provides a mock wizard context so
// the page renders without API calls.

function makeItem(
    id: number,
    detectionType: string,
    matchedText: string,
    resolution = 'pending',
): ContentAnalysisItem {
    return {
        id,
        jobId: 1,
        detectionType,
        matchedText,
        resolution,
        phase: 'identification',
        createdAt: '2026-01-01T00:00:00Z',
    } as ContentAnalysisItem;
}

function makeMockWizard(
    overrides: Partial<AnalysisWizardState> = {},
): AnalysisWizardState {
    return {
        job: {
            id: 1,
            campaignId: 1,
            sourceTable: 'chapters',
            sourceId: 1,
            sourceField: 'overview',
            status: 'pending',
            totalItems: 3,
            resolvedItems: 0,
            enrichmentTotal: 0,
            enrichmentResolved: 0,
            phases: ['identification'],
            currentPhase: 'identification',
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
        },
        items: [],
        isLoading: false,
        error: null,
        phases: ['identification'],
        currentPhase: 'identification',
        currentPhaseIndex: 0,
        phaseItems: [],
        pendingCount: 0,
        canAdvance: false,
        canGoBack: false,
        nextPhaseLabel: null,
        goToPhase: vi.fn(),
        goToNextPhase: vi.fn(),
        ...overrides,
    };
}

describe('IdentifyPhasePage', () => {
    it('groups items by detection type', async () =>
    {
        // This test verifies that items appear
        // under their detection type headings.
        const items = [
            makeItem(
                1,
                'wiki_link_resolved',
                'Viktor',
            ),
            makeItem(
                2,
                'wiki_link_resolved',
                'Elara',
            ),
            makeItem(
                3,
                'untagged_mention',
                'tavern',
            ),
        ];

        const wizard = makeMockWizard({
            phaseItems: items,
            pendingCount: 3,
        });

        // Import dynamically to allow the
        // placeholder to be replaced first.
        const { default: IdentifyPhasePage } =
            await import(
                './IdentifyPhasePage'
            );

        const queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false },
            },
        });

        render(
            <QueryClientProvider
                client={queryClient}
            >
                <MemoryRouter>
                    <AnalysisWizardProvider
                        value={wizard}
                    >
                        <IdentifyPhasePage />
                    </AnalysisWizardProvider>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        // The page should render the items.
        // Exact assertions depend on the final
        // UI, but at minimum the matched text
        // should appear.
        expect(
            screen.getByText('Viktor'),
        ).toBeDefined();
        expect(
            screen.getByText('tavern'),
        ).toBeDefined();
    });
});
```

**Step 2: Implement IdentifyPhasePage**

Replace the placeholder in
`client/src/pages/IdentifyPhasePage.tsx` with the
full component. The component structure:

- Reads `phaseItems` from `useWizardContext()`.
- Groups items by `detectionType` using
  `useMemo`.
- Left panel: renders each group as a collapsible
  section with a count badge. Each item row shows
  `matchedText`, entity chip (if `entityName` is
  set), and similarity score.
- Right panel: detail view for the selected item
  showing context snippet, entity match, and
  action buttons.
- State: `selectedItemId` tracked via `useState`.

Key interfaces for the component:

```typescript
interface DetectionGroup {
    type: string;
    items: ContentAnalysisItem[];
    pendingCount: number;
}
```

Actions use the existing hooks:

- `useResolveItem(campaignId)` for Accept,
  Dismiss, and Create Entity.
- `useBatchResolve(campaignId)` for Accept All.
- `useRevertItem(campaignId)` for Undo.

The Reassign action calls `useResolveItem` with
a `suggestedContentOverride` containing the
reassigned entity ID.

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/IdentifyPhasePage.test.tsx
```

**Commit message:**
`feat: implement IdentifyPhasePage with grouped items and detail panel`

---

### Task C2: Entity Reassign Autocomplete

This task adds an `EntityAutocomplete` component
for the Reassign action, allowing the GM to redirect
a match to a different entity.

**Files:**

- Create:
  `client/src/components/EntityAutocomplete.tsx`
- Test:
  `client/src/components/EntityAutocomplete.test.tsx`

**Step 1: Write the test**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from
    '@testing-library/react';
import { QueryClient, QueryClientProvider } from
    '@tanstack/react-query';

describe('EntityAutocomplete', () => {
    it('renders an autocomplete input', async () =>
    {
        const queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false },
            },
        });
        const { default: EntityAutocomplete } =
            await import(
                './EntityAutocomplete'
            );

        render(
            <QueryClientProvider
                client={queryClient}
            >
                <EntityAutocomplete
                    campaignId={1}
                    onSelect={vi.fn()}
                />
            </QueryClientProvider>,
        );

        expect(
            screen.getByRole('combobox'),
        ).toBeDefined();
    });
});
```

**Step 2: Implement the component**

The `EntityAutocomplete` component uses the MUI
`Autocomplete` with an entity search API call.
Props:

```typescript
interface EntityAutocompleteProps {
    campaignId: number;
    onSelect: (entity: {
        id: number;
        name: string;
        entityType: string;
    }) => void;
    excludeEntityId?: number;
    label?: string;
}
```

The component debounces input (300ms), queries
the entity list API filtered by name prefix,
and renders matching entities with type chips.

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/EntityAutocomplete.test.tsx
```

**Commit message:**
`feat: add EntityAutocomplete for reassign action`

---

### Task C3: Accept All Batch Action

This task adds the "Accept All" button to the
IdentifyPhasePage for batch-resolving all pending
items of a given detection type.

**Files:**

- Modify: `client/src/pages/IdentifyPhasePage.tsx`

**Step 1: Add the Accept All button**

In each detection type group header on the
IdentifyPhasePage, add an "Accept All" button
that calls `useBatchResolve` with the detection
type and `resolution: 'accepted'`.

The button should:

- Appear only for groups where all items have
  a matched entity (entity ID is set).
- Show a confirmation count ("Accept all 5
  resolved links").
- Disable while the mutation is in progress.

```typescript
const batchResolve = useBatchResolve(campaignId);

const handleAcceptAll = (
    detectionType: string,
) => {
    if (!wizard.job) return;
    batchResolve.mutate({
        jobId: wizard.job.id,
        detectionType,
        resolution: 'accepted',
    });
};
```

**Step 2: Run the test suite**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/IdentifyPhasePage.test.tsx
```

**Commit message:**
`feat: add Accept All batch action to IdentifyPhasePage`

---

## Phase D: Revise Phase Page

Phase D delivers the RevisePhasePage with analysis
findings, the revision workflow, iterative cycles,
auto re-identification after apply, and a new
mentions section.

### Task D1: RevisePhasePage Core UI

This task replaces the placeholder RevisePhasePage
with analysis findings grouped by detection type and
a detail panel.

**Files:**

- Modify: `client/src/pages/RevisePhasePage.tsx`
- Test:
  `client/src/pages/RevisePhasePage.test.tsx`

**Step 1: Write the test**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from
    '@testing-library/react';
import { QueryClient, QueryClientProvider } from
    '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { AnalysisWizardProvider } from
    '../contexts/AnalysisWizardContext';
import type { AnalysisWizardState } from
    '../hooks/useAnalysisWizard';
import type { ContentAnalysisItem } from
    '../api/contentAnalysis';

function makeItem(
    id: number,
    detectionType: string,
    matchedText: string,
): ContentAnalysisItem {
    return {
        id,
        jobId: 1,
        detectionType,
        matchedText,
        resolution: 'pending',
        phase: 'analysis',
        createdAt: '2026-01-01T00:00:00Z',
    } as ContentAnalysisItem;
}

describe('RevisePhasePage', () => {
    it('renders analysis findings', async () => {
        const items = [
            makeItem(
                1,
                'analysis_report',
                'The chapter lacks detail.',
            ),
            makeItem(
                2,
                'canon_contradiction',
                'Viktor died in Chapter 2.',
            ),
        ];

        const wizard = {
            job: {
                id: 1,
                campaignId: 1,
                sourceTable: 'chapters',
                sourceId: 1,
                sourceField: 'overview',
                status: 'pending',
                totalItems: 2,
                resolvedItems: 0,
                enrichmentTotal: 0,
                enrichmentResolved: 0,
                phases: [
                    'identification',
                    'analysis',
                ],
                currentPhase: 'analysis',
                createdAt:
                    '2026-01-01T00:00:00Z',
                updatedAt:
                    '2026-01-01T00:00:00Z',
            },
            items: [],
            isLoading: false,
            error: null,
            phases: [
                'identification', 'analysis',
            ],
            currentPhase: 'analysis',
            currentPhaseIndex: 1,
            phaseItems: items,
            pendingCount: 2,
            canAdvance: false,
            canGoBack: true,
            nextPhaseLabel: null,
            goToPhase: vi.fn(),
            goToNextPhase: vi.fn(),
        } as unknown as AnalysisWizardState;

        const queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false },
            },
        });

        const { default: RevisePhasePage } =
            await import('./RevisePhasePage');

        render(
            <QueryClientProvider
                client={queryClient}
            >
                <MemoryRouter>
                    <AnalysisWizardProvider
                        value={wizard}
                    >
                        <RevisePhasePage />
                    </AnalysisWizardProvider>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(
            screen.getByText(
                /lacks detail/i,
            ),
        ).toBeDefined();
    });
});
```

**Step 2: Implement RevisePhasePage**

Replace the placeholder with the full component.
The component structure:

- Left panel, section 1: analysis findings grouped
  by detection type, each with severity indicator
  and Acknowledge/Dismiss buttons.
- Left panel, section 2: new mentions that appear
  after applying a revision (items from the
  identification phase that were created after the
  last revision apply). These use the same grouping
  and actions as the Identify phase.
- Right panel: detail view for the selected finding
  or new mention.
- Revision workflow at the top: "Generate Revision"
  button, diff view, edit mode, "Apply Revision"
  button, iteration counter.

Key state:

```typescript
const [selectedItemId, setSelectedItemId] =
    useState<number | null>(null);
const [revisionContent, setRevisionContent] =
    useState<string | null>(null);
const [isEditing, setIsEditing] = useState(false);
const [revisionCount, setRevisionCount] =
    useState(0);
```

The component uses:

- `useGenerateRevision(campaignId)` for
  generating revisions.
- `useApplyRevision(campaignId)` for applying
  revisions (which triggers re-identification
  on the backend).
- `useResolveItem(campaignId)` for acknowledging
  or dismissing findings.

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/RevisePhasePage.test.tsx
```

**Commit message:**
`feat: implement RevisePhasePage with findings and revision workflow`

---

### Task D2: Revision Diff View

This task adds a side-by-side diff view component
used by the RevisePhasePage to compare original and
revised content.

**Files:**

- Create:
  `client/src/components/DiffView.tsx`
- Test:
  `client/src/components/DiffView.test.tsx`

**Step 1: Write the test**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from
    '@testing-library/react';

describe('DiffView', () => {
    it('renders original and revised content',
        async () => {
        const { default: DiffView } =
            await import('./DiffView');

        render(
            <DiffView
                original="Hello world"
                revised="Hello new world"
            />,
        );

        expect(
            screen.getByText(/original/i),
        ).toBeDefined();
        expect(
            screen.getByText(/revised/i),
        ).toBeDefined();
    });
});
```

**Step 2: Implement the component**

The DiffView component displays two side-by-side
panels (original on the left, revised on the right)
using MUI Grid. The component highlights changed
lines. An optional `onEdit` prop enables inline
editing of the revised content.

Props:

```typescript
interface DiffViewProps {
    original: string;
    revised: string;
    onEdit?: (newContent: string) => void;
    editable?: boolean;
}
```

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/DiffView.test.tsx
```

**Commit message:**
`feat: add DiffView component for revision comparison`

---

## Phase E: Enrich Phase Page

Phase E delivers the EnrichPhasePage with two-pass
enrichment, edit-before-accept, description
auto-tagging, graph health checks, per-entity error
handling, and cleanup of the old AnalysisTriagePage.

### Task E1: EnrichPhasePage Core UI

This task replaces the placeholder EnrichPhasePage
with the two-pass enrichment UI: Pass 1 entity scan
checklist and Pass 2 per-entity enrichment results.

**Files:**

- Modify: `client/src/pages/EnrichPhasePage.tsx`
- Test:
  `client/src/pages/EnrichPhasePage.test.tsx`

**Step 1: Write the test**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from
    '@testing-library/react';
import { QueryClient, QueryClientProvider } from
    '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { AnalysisWizardProvider } from
    '../contexts/AnalysisWizardContext';
import type { AnalysisWizardState } from
    '../hooks/useAnalysisWizard';
import type { ContentAnalysisItem } from
    '../api/contentAnalysis';

function makeItem(
    id: number,
    detectionType: string,
    matchedText: string,
): ContentAnalysisItem {
    return {
        id,
        jobId: 1,
        detectionType,
        matchedText,
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2026-01-01T00:00:00Z',
        suggestedContent: {},
    } as ContentAnalysisItem;
}

describe('EnrichPhasePage', () => {
    it('renders enrichment items', async () => {
        const items = [
            makeItem(
                1,
                'description_update',
                'Viktor',
            ),
            makeItem(
                2,
                'relationship_suggestion',
                'Viktor -> Elara',
            ),
        ];

        const wizard = {
            job: {
                id: 1,
                campaignId: 1,
                sourceTable: 'chapters',
                sourceId: 1,
                sourceField: 'overview',
                status: 'enriching',
                totalItems: 0,
                resolvedItems: 0,
                enrichmentTotal: 2,
                enrichmentResolved: 0,
                phases: [
                    'identification',
                    'analysis',
                    'enrichment',
                ],
                currentPhase: 'enrichment',
                createdAt:
                    '2026-01-01T00:00:00Z',
                updatedAt:
                    '2026-01-01T00:00:00Z',
            },
            items: items,
            isLoading: false,
            error: null,
            phases: [
                'identification',
                'analysis',
                'enrichment',
            ],
            currentPhase: 'enrichment',
            currentPhaseIndex: 2,
            phaseItems: items,
            pendingCount: 2,
            canAdvance: false,
            canGoBack: true,
            nextPhaseLabel: null,
            goToPhase: vi.fn(),
            goToNextPhase: vi.fn(),
        } as unknown as AnalysisWizardState;

        const queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false },
            },
        });

        const { default: EnrichPhasePage } =
            await import('./EnrichPhasePage');

        render(
            <QueryClientProvider
                client={queryClient}
            >
                <MemoryRouter>
                    <AnalysisWizardProvider
                        value={wizard}
                    >
                        <EnrichPhasePage />
                    </AnalysisWizardProvider>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(
            screen.getByText(/Viktor/),
        ).toBeDefined();
    });
});
```

**Step 2: Implement EnrichPhasePage**

Replace the placeholder with the full component.
The component structure:

Pass 1 (entity scan checklist):

- The left panel shows a checklist of entities
  that could benefit from enrichment.
- Existing entities show name, what can improve,
  and a one-line reason.
- New entity suggestions show name, type, and
  reason.
- Each entity has a checkbox for selection.
- An "Enrich Selected" button triggers Pass 2
  for checked entities.
- Individual "Enrich" buttons per entity are also
  available.

Pass 2 (per-entity enrichment results):

- Results stream in via `useEnrichmentStream`.
- Items are grouped under each entity name.
- Each item type has a specific detail view in
  the right panel:
  - `description_update`: diff view with
    edit-before-accept.
  - `relationship_suggestion`: visual diagram
    with type autocomplete.
  - `log_entry`: content and timestamp display.
  - `new_entity_suggestion`: editable name, type,
    and description. Auto-tags descriptions with
    wiki-links for exact entity name matches on
    accept.
  - Graph advisory items (`graph_warning`,
    `redundant_edge`, `invalid_type_pair`,
    `orphan_warning`): description and
    recommendation.

Per-entity error handling:

- Failed entities display a red error badge
  with the failure reason.
- A "Retry" button appears for failed entities.
- Failures on one entity do not block other
  entities.

Key state:

```typescript
const [selectedEntities, setSelectedEntities] =
    useState<Set<number>>(new Set());
const [selectedItemId, setSelectedItemId] =
    useState<number | null>(null);
const [editingContent, setEditingContent] =
    useState<Record<number, unknown>>({});
```

The component uses:

- `useTriggerEnrichment(campaignId)` for
  triggering enrichment.
- `useEnrichmentStream(campaignId, jobId, ...)
  for polling results.
- `useCancelEnrichment(campaignId)` for
  cancellation.
- `useResolveItem(campaignId)` for accepting,
  dismissing, and creating entities.

**Step 3: Run the test**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/EnrichPhasePage.test.tsx
```

**Commit message:**
`feat: implement EnrichPhasePage with two-pass enrichment`

---

### Task E2: Edit-Before-Accept for Descriptions

This task adds inline editing to description update
items in the EnrichPhasePage, allowing the GM to
modify the suggested description before accepting
the suggestion.

**Files:**

- Modify: `client/src/pages/EnrichPhasePage.tsx`

**Step 1: Add edit mode to description items**

In the right panel detail view for
`description_update` items, add an "Edit" toggle
that replaces the diff view with an editable
textarea. The edited content is stored in
`editingContent` state keyed by item ID.

When the GM clicks "Accept", the component sends
the edited content as a
`suggestedContentOverride` in the resolve request.

```typescript
const handleAcceptDescription = (
    item: ContentAnalysisItem,
) => {
    const override = editingContent[item.id];
    resolveItem.mutate({
        itemId: item.id,
        req: {
            resolution: 'accepted',
            suggestedContentOverride: override
                ? {
                      suggestedDescription:
                          override,
                  }
                : undefined,
        },
    });
};
```

**Step 2: Run the test suite**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/EnrichPhasePage.test.tsx
```

**Commit message:**
`feat: add edit-before-accept for description updates`

---

### Task E3: Description Auto-Tagging

This task adds wiki-link auto-tagging to new entity
descriptions when the GM accepts a
`new_entity_suggestion`. The system scans the
description for exact entity name matches and wraps
matching text in `[[entity_name]]` syntax.

**Files:**

- Create:
  `client/src/utils/autoTagWikiLinks.ts`
- Test:
  `client/src/utils/autoTagWikiLinks.test.ts`

**Step 1: Write the tests**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import { autoTagWikiLinks } from
    './autoTagWikiLinks';

describe('autoTagWikiLinks', () => {
    it('wraps exact name matches', () => {
        const result = autoTagWikiLinks(
            'Viktor met Elara at the tavern.',
            ['Viktor', 'Elara'],
        );
        expect(result).toBe(
            '[[Viktor]] met [[Elara]] at the ' +
                'tavern.',
        );
    });

    it('skips already tagged names', () => {
        const result = autoTagWikiLinks(
            '[[Viktor]] met Elara.',
            ['Viktor', 'Elara'],
        );
        expect(result).toBe(
            '[[Viktor]] met [[Elara]].',
        );
    });

    it('handles case-insensitive match', () => {
        const result = autoTagWikiLinks(
            'viktor arrived.',
            ['Viktor'],
        );
        expect(result).toBe(
            '[[viktor]] arrived.',
        );
    });

    it('handles no matches', () => {
        const result = autoTagWikiLinks(
            'Nothing to tag here.',
            ['Viktor'],
        );
        expect(result).toBe(
            'Nothing to tag here.',
        );
    });

    it('handles empty entity list', () => {
        const result = autoTagWikiLinks(
            'Some text.',
            [],
        );
        expect(result).toBe('Some text.');
    });
});
```

**Step 2: Implement the utility**

```typescript
/*--------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *--------------------------------------------------
 */

/**
 * Wraps untagged entity name occurrences in
 * wiki-link syntax ([[name]]). Names that already
 * appear inside wiki-links are skipped.
 *
 * @param text - The text to scan for entity names.
 * @param entityNames - The entity names to match.
 * @returns The text with matching names wrapped in
 *          wiki-links.
 */
export function autoTagWikiLinks(
    text: string,
    entityNames: string[],
): string {
    if (entityNames.length === 0) return text;

    let result = text;
    for (const name of entityNames) {
        // Match the name when it is NOT already
        // inside [[ ]].
        const escaped = name.replace(
            /[.*+?^${}()|[\]\\]/g, '\\$&',
        );
        const pattern = new RegExp(
            `(?<!\\[\\[)\\b(${escaped})\\b(?!\\]\\])`,
            'gi',
        );
        result = result.replace(
            pattern, '[[$1]]',
        );
    }
    return result;
}
```

**Step 3: Run the tests**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/utils/autoTagWikiLinks.test.ts
```

**Commit message:**
`feat: add autoTagWikiLinks utility for description tagging`

---

### Task E4: Graph Health Section

This task adds a graph health section to the
EnrichPhasePage that displays graph advisory items
(orphan warnings, redundant edges, invalid type
pairs) after enrichment completes.

**Files:**

- Modify: `client/src/pages/EnrichPhasePage.tsx`

**Step 1: Add graph health section**

After the per-entity enrichment results, add a
"Graph Health" section that filters items by the
graph-related detection types: `graph_warning`,
`redundant_edge`, `invalid_type_pair`, and
`orphan_warning`.

Each graph advisory item displays:

- A description of the issue.
- A recommendation for resolution.
- Acknowledge and Dismiss buttons.

The section only appears when graph health items
exist.

```typescript
const graphItems = useMemo(
    () =>
        phaseItems.filter((item) =>
            [
                'graph_warning',
                'redundant_edge',
                'invalid_type_pair',
                'orphan_warning',
            ].includes(item.detectionType),
        ),
    [phaseItems],
);
```

**Step 2: Run the test suite**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/EnrichPhasePage.test.tsx
```

**Commit message:**
`feat: add graph health section to EnrichPhasePage`

---

### Task E5: Delete Old AnalysisTriagePage

This task removes the old monolithic
AnalysisTriagePage and verifies the application
compiles and all tests pass without the old
component.

**Files:**

- Delete: `client/src/pages/AnalysisTriagePage.tsx`
- Modify: `client/src/App.tsx` (remove any
  remaining import)

**Step 1: Remove the old file**

Delete
`client/src/pages/AnalysisTriagePage.tsx`.

**Step 2: Clean up imports**

In `client/src/App.tsx`, remove the
`AnalysisTriagePage` import if any reference
remains. The route was already replaced in
Task B4.

**Step 3: Run the full test suite**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run
```

Verify the TypeScript compilation also passes.

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx tsc --noEmit
```

**Step 4: Run the Go test suite**

```bash
go test ./... -v
```

**Commit message:**
`refactor: remove monolithic AnalysisTriagePage`

---

## Task Summary

The following table summarizes all tasks across
delivery phases.

| Phase | Task | Description                       |
|-------|------|-----------------------------------|
| A     | A1   | Database migration                |
| A     | A2   | LLM quota detection               |
| A     | A3   | failure_reason in handlers        |
| A     | A4   | Populate Relationships            |
| A     | A5   | Frontend type updates             |
| B     | B1   | useAnalysisWizard hook            |
| B     | B2   | AnalysisWizardContext             |
| B     | B3   | AnalysisWizard shell              |
| B     | B4   | Nested routing in App.tsx         |
| C     | C1   | IdentifyPhasePage core UI         |
| C     | C2   | EntityAutocomplete                |
| C     | C3   | Accept All batch action           |
| D     | D1   | RevisePhasePage core UI           |
| D     | D2   | Revision diff view                |
| E     | E1   | EnrichPhasePage core UI           |
| E     | E2   | Edit-before-accept                |
| E     | E3   | Description auto-tagging          |
| E     | E4   | Graph health section              |
| E     | E5   | Delete old AnalysisTriagePage     |

Total: 19 tasks across 5 delivery phases.
