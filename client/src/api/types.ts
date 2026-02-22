// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Options for API calls that can trigger content analysis.
 */
export interface AnalysisOptions {
    analyze?: boolean;
    enrich?: boolean;
    phases?: string[];
}
