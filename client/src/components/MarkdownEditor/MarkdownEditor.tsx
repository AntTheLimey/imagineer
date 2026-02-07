/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useEffect, useRef } from 'react';
import {
    Box,
    FormHelperText,
    InputLabel,
    Paper,
} from '@mui/material';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import { Markdown } from 'tiptap-markdown';
import EditorToolbar from './EditorToolbar';

/**
 * Shape of the tiptap-markdown storage object.
 *
 * The tiptap-markdown package does not ship TypeScript declarations
 * for its editor storage, so this interface provides type safety for
 * accessing `editor.storage.markdown`.
 */
interface MarkdownStorage {
    markdown: {
        getMarkdown: () => string;
    };
}

/**
 * Props for the MarkdownEditor component.
 */
export interface MarkdownEditorProps {
    /** The Markdown content of the editor. */
    value: string;
    /** Callback fired when the content changes, receiving Markdown. */
    onChange: (markdown: string) => void;
    /** Placeholder text displayed when the editor is empty. */
    placeholder?: string;
    /** Label displayed above the editor. */
    label?: string;
    /** If true, the editor displays in an error state. */
    error?: boolean;
    /** Helper text displayed below the editor. */
    helperText?: string;
    /** Minimum height of the editor content area in pixels. */
    minHeight?: number;
    /** Maximum height of the editor content area in pixels. */
    maxHeight?: number;
    /** If true, the editor is disabled. */
    disabled?: boolean;
}

/**
 * A MUI-styled Markdown editor component wrapping TipTap.
 *
 * Provides a WYSIWYG editing experience with Markdown serialization
 * via the tiptap-markdown extension. Content is stored and emitted as
 * Markdown rather than HTML. Supports formatting options including
 * bold, italic, strikethrough, headings, lists, blockquotes,
 * horizontal rules, and undo/redo. The editor is styled to match MUI
 * TextField aesthetics with support for labels, error states, and
 * helper text.
 *
 * @param props - The component props.
 * @returns A React element containing the Markdown editor.
 *
 * @example
 * ```tsx
 * const [content, setContent] = useState('# Hello World');
 *
 * <MarkdownEditor
 *     value={content}
 *     onChange={setContent}
 *     label="Description"
 *     placeholder="Enter a description..."
 * />
 * ```
 */
export default function MarkdownEditor({
    value,
    onChange,
    placeholder,
    label,
    error = false,
    helperText,
    minHeight = 200,
    maxHeight = 500,
    disabled = false,
}: MarkdownEditorProps) {
    // Keep a ref to the latest onChange callback so the onUpdate
    // closure always invokes the current version (Bug 2 fix).
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;

    // Track whether a content change originated from the editor
    // itself so the value-sync useEffect can skip re-setting content
    // that the editor already knows about (Bug 1 fix).
    const isInternalUpdate = useRef(false);

    const editor = useEditor({
        extensions: [
            StarterKit,
            Placeholder.configure({
                placeholder: placeholder ?? '',
            }),
            Markdown.configure({
                html: false,
                transformCopiedText: true,
                transformPastedText: true,
            }),
        ],
        content: value,
        editable: !disabled,
        onUpdate: ({ editor }) => {
            const storage = editor.storage as unknown as MarkdownStorage;
            const markdown = storage.markdown.getMarkdown();
            isInternalUpdate.current = true;
            onChangeRef.current(markdown);
        },
    });

    // Update editor content when value prop changes externally
    useEffect(() => {
        if (!editor) return;

        // Skip if this render was triggered by the editor's own
        // onUpdate callback to avoid an infinite loop and cursor
        // jumping.
        if (isInternalUpdate.current) {
            isInternalUpdate.current = false;
            return;
        }

        const storage = editor.storage as unknown as MarkdownStorage;
        const currentMarkdown = storage.markdown.getMarkdown();

        if (value !== currentMarkdown) {
            editor.commands.setContent(value || '');
        }
    }, [editor, value]);

    // Update editable state when disabled prop changes
    useEffect(() => {
        if (editor) {
            editor.setEditable(!disabled);
        }
    }, [editor, disabled]);

    return (
        <Box>
            {label && (
                <InputLabel
                    error={error}
                    sx={{
                        mb: 0.5,
                        fontSize: '0.875rem',
                        color: error
                            ? 'error.main'
                            : 'text.secondary',
                    }}
                >
                    {label}
                </InputLabel>
            )}
            <Paper
                variant="outlined"
                sx={{
                    borderColor: error ? 'error.main' : 'divider',
                    '&:hover': {
                        borderColor: disabled
                            ? 'divider'
                            : error
                              ? 'error.main'
                              : 'text.primary',
                    },
                    '&:focus-within': {
                        borderColor: error
                            ? 'error.main'
                            : 'primary.main',
                        borderWidth: 2,
                    },
                    opacity: disabled ? 0.5 : 1,
                }}
            >
                <EditorToolbar
                    editor={editor}
                    disabled={disabled}
                />
                <Box
                    sx={{
                        p: 2,
                        minHeight,
                        maxHeight,
                        overflowY: 'auto',
                        '& .tiptap': {
                            outline: 'none',
                            minHeight: minHeight - 32,
                            '& p.is-editor-empty:first-child::before':
                                {
                                    content:
                                        'attr(data-placeholder)',
                                    color: 'text.disabled',
                                    pointerEvents: 'none',
                                    float: 'left',
                                    height: 0,
                                },
                            '& p': {
                                margin: 0,
                                marginBottom: 1,
                            },
                            '& h1, & h2, & h3': {
                                marginTop: 1,
                                marginBottom: 0.5,
                            },
                            '& ul, & ol': {
                                paddingLeft: 3,
                                marginBottom: 1,
                            },
                        },
                    }}
                >
                    <EditorContent editor={editor} />
                </Box>
            </Paper>
            {helperText && (
                <FormHelperText error={error} sx={{ mx: 1.75 }}>
                    {helperText}
                </FormHelperText>
            )}
        </Box>
    );
}
