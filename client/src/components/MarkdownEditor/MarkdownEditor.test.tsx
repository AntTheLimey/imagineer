/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import MarkdownEditor from './MarkdownEditor';

describe('MarkdownEditor', () => {
    it('renders without crashing', async () => {
        const onChange = vi.fn();
        const { container } = render(
            <MarkdownEditor value="" onChange={onChange} />
        );

        await waitFor(() => {
            expect(
                container.querySelector('.ProseMirror')
            ).toBeInTheDocument();
        });
    });

    it('renders with initial Markdown value', async () => {
        const onChange = vi.fn();
        const { container } = render(
            <MarkdownEditor
                value="**bold text**"
                onChange={onChange}
            />
        );

        await waitFor(() => {
            const proseMirror = container.querySelector('.ProseMirror');
            expect(proseMirror).toBeInTheDocument();
            const strong = proseMirror?.querySelector('strong');
            expect(strong).toBeInTheDocument();
            expect(strong?.textContent).toBe('bold text');
        });
    });

    it('displays label when provided', async () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                label="Description"
            />
        );

        expect(screen.getByText('Description')).toBeInTheDocument();
    });

    it('displays placeholder when empty', async () => {
        const onChange = vi.fn();
        const { container } = render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                placeholder="Enter text..."
            />
        );

        await waitFor(() => {
            expect(
                container.querySelector('.is-editor-empty')
            ).toBeInTheDocument();
        });
    });

    it('displays helper text when provided', async () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                helperText="Required field"
            />
        );

        expect(
            screen.getByText('Required field')
        ).toBeInTheDocument();
    });

    it('handles empty string value gracefully', async () => {
        const onChange = vi.fn();

        expect(() => {
            render(
                <MarkdownEditor value="" onChange={onChange} />
            );
        }).not.toThrow();
    });
});
