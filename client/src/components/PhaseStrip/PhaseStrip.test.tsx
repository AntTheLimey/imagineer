// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Tests for the PhaseStrip component.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import PhaseStrip from './PhaseStrip';

describe('PhaseStrip', () => {
    const defaultProps = {
        onSave: vi.fn(),
        isDirty: true,
        isSaving: false,
    };

    beforeEach(() => {
        vi.clearAllMocks();
        localStorage.clear();
    });

    it('renders three phase checkboxes', () => {
        render(<PhaseStrip {...defaultProps} />);

        expect(screen.getByLabelText(/Identify/)).toBeInTheDocument();
        expect(screen.getByLabelText(/Revise/)).toBeInTheDocument();
        expect(screen.getByLabelText(/Enrich/)).toBeInTheDocument();
    });

    it('loads default state from localStorage', () => {
        localStorage.setItem(
            'imagineer:phaseSelection',
            JSON.stringify({ identify: false, revise: true, enrich: true }),
        );

        render(<PhaseStrip {...defaultProps} />);

        const checkboxes = screen.getAllByRole('checkbox');
        // Order: Identify, Revise, Enrich
        expect(checkboxes[0]).not.toBeChecked(); // identify
        expect(checkboxes[1]).toBeChecked();     // revise
        expect(checkboxes[2]).toBeChecked();     // enrich
    });

    it('uses defaults when localStorage is empty', () => {
        render(<PhaseStrip {...defaultProps} />);

        const checkboxes = screen.getAllByRole('checkbox');
        // Defaults: identify=true, revise=false, enrich=false
        expect(checkboxes[0]).toBeChecked();     // identify
        expect(checkboxes[1]).not.toBeChecked(); // revise
        expect(checkboxes[2]).not.toBeChecked(); // enrich
    });

    it('toggles a checkbox when clicked', async () => {
        const user = userEvent.setup();
        render(<PhaseStrip {...defaultProps} />);

        const identifyCheckbox = screen.getAllByRole('checkbox')[0];
        expect(identifyCheckbox).toBeChecked();

        await user.click(identifyCheckbox);
        expect(identifyCheckbox).not.toBeChecked();

        await user.click(identifyCheckbox);
        expect(identifyCheckbox).toBeChecked();
    });

    it('shows "Save" when no phases are checked', async () => {
        const user = userEvent.setup();
        // Default has identify checked; uncheck it
        render(<PhaseStrip {...defaultProps} />);

        const identifyCheckbox = screen.getAllByRole('checkbox')[0];
        await user.click(identifyCheckbox);

        expect(
            screen.getByRole('button', { name: /^Save$/i }),
        ).toBeInTheDocument();
    });

    it('shows "Save & Go" when any phase is checked', () => {
        // Default has identify=true
        render(<PhaseStrip {...defaultProps} />);

        expect(
            screen.getByRole('button', { name: /Save & Go/i }),
        ).toBeInTheDocument();
    });

    it('disables the button when not dirty', () => {
        render(<PhaseStrip {...defaultProps} isDirty={false} />);

        expect(screen.getByRole('button', { name: /Save/i })).toBeDisabled();
    });

    it('disables the button when saving', () => {
        render(<PhaseStrip {...defaultProps} isSaving={true} />);

        expect(
            screen.getByRole('button', { name: /Save/i }),
        ).toBeDisabled();
    });

    it('calls onSave with current phase selection when clicked', async () => {
        const user = userEvent.setup();
        const onSave = vi.fn();

        localStorage.setItem(
            'imagineer:phaseSelection',
            JSON.stringify({ identify: true, revise: false, enrich: true }),
        );

        render(<PhaseStrip {...defaultProps} onSave={onSave} />);

        await user.click(screen.getByRole('button', { name: /Save & Go/i }));

        expect(onSave).toHaveBeenCalledTimes(1);
        expect(onSave).toHaveBeenCalledWith({
            identify: true,
            revise: false,
            enrich: true,
        });
    });

    it('persists phase selection to localStorage on change', async () => {
        const user = userEvent.setup();
        render(<PhaseStrip {...defaultProps} />);

        // Toggle revise on
        const reviseCheckbox = screen.getAllByRole('checkbox')[1];
        await user.click(reviseCheckbox);

        const stored = JSON.parse(
            localStorage.getItem('imagineer:phaseSelection')!,
        );
        expect(stored).toEqual({
            identify: true,
            revise: true,
            enrich: false,
        });
    });

    it('renders subtitle text for each phase', () => {
        render(<PhaseStrip {...defaultProps} />);

        expect(screen.getByText('Pattern detection')).toBeInTheDocument();
        expect(screen.getByText('AI advisory')).toBeInTheDocument();
        expect(screen.getByText('Structural')).toBeInTheDocument();
    });

    it('shows a spinner on the button while saving', () => {
        render(<PhaseStrip {...defaultProps} isSaving={true} />);

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
});
