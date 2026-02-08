/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MarkdownRenderer from './MarkdownRenderer';

describe('MarkdownRenderer', () => {
    it('renders null content without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content={null} />
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders undefined content without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content={undefined} />
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders empty string without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content="" />
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders bold text', () => {
        render(<MarkdownRenderer content="**bold text**" />);
        const strong = screen.getByText('bold text');
        expect(strong.tagName).toBe('STRONG');
    });

    it('renders italic text', () => {
        render(<MarkdownRenderer content="*italic text*" />);
        const em = screen.getByText('italic text');
        expect(em.tagName).toBe('EM');
    });

    it('renders headings', () => {
        render(<MarkdownRenderer content="## Section Title" />);
        const heading = screen.getByRole('heading', { level: 2 });
        expect(heading).toBeInTheDocument();
        expect(heading).toHaveTextContent('Section Title');
    });

    it('renders bullet lists', () => {
        render(
            <MarkdownRenderer content={"- item one\n- item two"} />
        );
        const items = screen.getAllByRole('listitem');
        expect(items).toHaveLength(2);
    });

    it('renders blockquotes', () => {
        render(<MarkdownRenderer content="> quoted text" />);
        const blockquote = document.querySelector('blockquote');
        expect(blockquote).toBeInTheDocument();
        expect(blockquote).toHaveTextContent('quoted text');
    });

    it('renders horizontal rules', () => {
        render(
            <MarkdownRenderer
                content={"above\n\n---\n\nbelow"}
            />
        );
        const hr = document.querySelector('hr');
        expect(hr).toBeInTheDocument();
    });

    it('blocks raw HTML tags', () => {
        render(
            <MarkdownRenderer
                content={'<script>alert("xss")</script>'}
            />
        );
        const script = document.querySelector('script');
        expect(script).not.toBeInTheDocument();
    });
});
