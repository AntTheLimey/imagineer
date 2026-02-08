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
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

    describe('wiki links', () => {
        it('renders [[Entity]] as a clickable element', () => {
            render(
                <MarkdownRenderer
                    content="Meet [[Inspector Legrasse]] at the docks"
                />
            );

            const wikiLink = screen.getByTestId('wiki-link');
            expect(wikiLink).toBeInTheDocument();
            expect(wikiLink).toHaveTextContent('Inspector Legrasse');
            expect(wikiLink).toHaveAttribute(
                'data-entity-name',
                'Inspector Legrasse'
            );
        });

        it('renders [[Entity|alias]] showing alias text', () => {
            render(
                <MarkdownRenderer
                    content="Talk to [[Inspector Legrasse|the Inspector]]"
                />
            );

            const wikiLink = screen.getByTestId('wiki-link');
            expect(wikiLink).toBeInTheDocument();
            expect(wikiLink).toHaveTextContent('the Inspector');
            expect(wikiLink).toHaveAttribute(
                'data-entity-name',
                'Inspector Legrasse'
            );
        });

        it('calls onEntityClick when clicking a wiki link', async () => {
            const user = userEvent.setup();
            const handleClick = vi.fn();

            render(
                <MarkdownRenderer
                    content="Meet [[Inspector Legrasse]] here"
                    onEntityClick={handleClick}
                />
            );

            const wikiLink = screen.getByTestId('wiki-link');
            await user.click(wikiLink);

            expect(handleClick).toHaveBeenCalledTimes(1);
            expect(handleClick).toHaveBeenCalledWith(
                'Inspector Legrasse'
            );
        });

        it('renders multiple wiki links in one paragraph', () => {
            render(
                <MarkdownRenderer
                    content="[[Alpha]] met [[Beta]] at [[Gamma]]"
                />
            );

            const wikiLinks = screen.getAllByTestId('wiki-link');
            expect(wikiLinks).toHaveLength(3);
            expect(wikiLinks[0]).toHaveTextContent('Alpha');
            expect(wikiLinks[1]).toHaveTextContent('Beta');
            expect(wikiLinks[2]).toHaveTextContent('Gamma');
        });

        it('does not render wiki links inside inline code', () => {
            render(
                <MarkdownRenderer
                    content="Use `[[Entity]]` syntax"
                />
            );

            const wikiLinks = screen.queryAllByTestId('wiki-link');
            expect(wikiLinks).toHaveLength(0);
            expect(screen.getByText('[[Entity]]')).toBeInTheDocument();
        });

        it('renders wiki links inside bold text', () => {
            render(
                <MarkdownRenderer
                    content="**bold [[Entity]] text**"
                />
            );

            const wikiLink = screen.getByTestId('wiki-link');
            expect(wikiLink).toHaveTextContent('Entity');
        });

        it('does not call onEntityClick when callback is not provided', async () => {
            const user = userEvent.setup();

            render(
                <MarkdownRenderer
                    content="Meet [[Inspector Legrasse]] here"
                />
            );

            const wikiLink = screen.getByTestId('wiki-link');
            // Clicking should not throw when no callback is provided.
            await user.click(wikiLink);
            expect(wikiLink).toBeInTheDocument();
        });
    });
});
