// @vitest-environment jsdom
/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * Tests for the DiffView component.
 */

import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DiffView, diffLines } from './DiffView';

describe('diffLines', () => {
    it('returns a single unchanged entry when both strings are empty', () => {
        const result = diffLines('', '');

        expect(result).toHaveLength(1);
        expect(result[0]).toEqual({
            origLine: '',
            revLine: '',
            changed: false,
        });
    });

    it('marks identical lines as unchanged', () => {
        const text = 'line one\nline two';
        const result = diffLines(text, text);

        expect(result).toHaveLength(2);
        expect(result.every((l) => !l.changed)).toBe(true);
    });

    it('marks differing lines as changed', () => {
        const result = diffLines('alpha\nbeta', 'alpha\ngamma');

        expect(result[0].changed).toBe(false);
        expect(result[1].changed).toBe(true);
        expect(result[1].origLine).toBe('beta');
        expect(result[1].revLine).toBe('gamma');
    });

    it('pads shorter text with empty strings', () => {
        const result = diffLines('one\ntwo\nthree', 'one');

        expect(result).toHaveLength(3);
        expect(result[1]).toEqual({
            origLine: 'two',
            revLine: '',
            changed: true,
        });
        expect(result[2]).toEqual({
            origLine: 'three',
            revLine: '',
            changed: true,
        });
    });
});

describe('DiffView', () => {
    afterEach(() => {
        cleanup();
    });

    it('renders "Original" and "Revised" headers', () => {
        render(<DiffView original="hello" revised="hello" />);

        expect(screen.getByText('Original')).toBeInTheDocument();
        expect(screen.getByText('Revised')).toBeInTheDocument();
    });

    it('renders original and revised content', () => {
        render(
            <DiffView original="original text" revised="revised text" />,
        );

        expect(screen.getByText('original text')).toBeInTheDocument();
        expect(screen.getByText('revised text')).toBeInTheDocument();
    });

    it('highlights changed lines with different background test IDs', () => {
        render(
            <DiffView
                original={'same\nremoved'}
                revised={'same\nadded'}
            />,
        );

        // Changed lines on the original side get the "removed" test ID.
        const removedLines = screen.getAllByTestId('diff-line-removed');
        expect(removedLines).toHaveLength(1);
        expect(removedLines[0]).toHaveTextContent('removed');

        // Changed lines on the revised side get the "added" test ID.
        const addedLines = screen.getAllByTestId('diff-line-added');
        expect(addedLines).toHaveLength(1);
        expect(addedLines[0]).toHaveTextContent('added');
    });

    it('does not mark unchanged lines as changed', () => {
        render(<DiffView original="same" revised="same" />);

        const unchangedLines = screen.getAllByTestId('diff-line-unchanged');

        // Both panels render one unchanged line each.
        expect(unchangedLines).toHaveLength(2);

        // No changed lines should exist.
        expect(screen.queryByTestId('diff-line-removed')).not.toBeInTheDocument();
        expect(screen.queryByTestId('diff-line-added')).not.toBeInTheDocument();
    });

    it('shows a TextField when editable is true and onEdit is provided', () => {
        const onEdit = vi.fn();
        render(
            <DiffView
                original="original"
                revised="revised"
                editable={true}
                onEdit={onEdit}
            />,
        );

        const textArea = screen.getByLabelText('Edit revised content');
        expect(textArea).toBeInTheDocument();
        // The aria-label is placed on the underlying textarea element.
        expect(textArea.tagName).toBe('TEXTAREA');
    });

    it('does not show a TextField when editable is true but onEdit is missing', () => {
        render(
            <DiffView
                original="original"
                revised="revised"
                editable={true}
            />,
        );

        expect(
            screen.queryByLabelText('Edit revised content'),
        ).not.toBeInTheDocument();
    });

    it('calls onEdit when the user types in the editable TextField', async () => {
        const user = userEvent.setup();
        const onEdit = vi.fn();

        render(
            <DiffView
                original="original"
                revised="revised"
                editable={true}
                onEdit={onEdit}
            />,
        );

        const textArea = screen.getByLabelText('Edit revised content');
        await user.type(textArea, ' more');

        // onEdit should have been called for each keystroke.
        expect(onEdit).toHaveBeenCalled();
        // The last call should contain the revised text with the
        // appended characters.
        const lastCall = onEdit.mock.calls[onEdit.mock.calls.length - 1];
        expect(lastCall[0]).toContain('revised');
    });

    it('handles empty strings gracefully', () => {
        render(<DiffView original="" revised="" />);

        expect(screen.getByText('Original')).toBeInTheDocument();
        expect(screen.getByText('Revised')).toBeInTheDocument();

        // Both panels should render at least one unchanged line.
        const unchangedLines = screen.getAllByTestId('diff-line-unchanged');
        expect(unchangedLines.length).toBeGreaterThanOrEqual(2);
    });

    it('handles revised being longer than original', () => {
        render(
            <DiffView
                original="line one"
                revised={'line one\nline two\nline three'}
            />,
        );

        // "line two" and "line three" are only on the revised side.
        expect(screen.getByText('line two')).toBeInTheDocument();
        expect(screen.getByText('line three')).toBeInTheDocument();
    });

    it('handles original being longer than revised', () => {
        render(
            <DiffView
                original={'line one\nline two\nline three'}
                revised="line one"
            />,
        );

        // "line one" appears in both panels; use getAllByText.
        const lineOneElements = screen.getAllByText('line one');
        expect(lineOneElements).toHaveLength(2);

        // "line two" and "line three" appear only in the original panel.
        expect(screen.getByText('line two')).toBeInTheDocument();
        expect(screen.getByText('line three')).toBeInTheDocument();
    });
});
