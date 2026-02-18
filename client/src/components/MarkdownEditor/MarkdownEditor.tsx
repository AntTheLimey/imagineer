/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useEffect, useMemo, useRef } from 'react';
import {
    Box,
    FormHelperText,
    InputLabel,
    Paper,
} from '@mui/material';
import { Extension } from '@tiptap/core';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import Suggestion from '@tiptap/suggestion';
import { Markdown } from 'tiptap-markdown';
import EditorToolbar from './EditorToolbar';
import WikiLinkNode from './WikiLinkNode';
import { buildWikiLinkSuggestion } from './WikiLinkSuggestion';

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
    /** Campaign ID for wiki link entity resolution. */
    campaignId?: number;
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
    campaignId,
}: MarkdownEditorProps) {
    // Keep a ref to the latest onChange callback so the onUpdate
    // closure always invokes the current version (Bug 2 fix).
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;

    // Track the last value reported by onUpdate so the value-sync
    // effect can distinguish external value changes (from the parent)
    // from internal changes (from the editor itself). This replaces
    // the previous isInternalUpdate boolean flag, which was fragile
    // and could become desynchronized during async rendering.
    const lastEditorValueRef = useRef(value);

    // Build the extensions array, conditionally including the wiki
    // link suggestion plugin when a campaignId is available.
    const extensions = useMemo(() => {
        const base = [
            StarterKit,
            Placeholder.configure({
                placeholder: placeholder ?? '',
            }),
            Markdown.configure({
                html: false,
                transformCopiedText: true,
                transformPastedText: true,
            }),
            WikiLinkNode,
        ];

        if (campaignId) {
            const suggestionConfig =
                buildWikiLinkSuggestion(campaignId);

            const WikiLinkSuggestionExtension = Extension.create({
                name: 'wikiLinkSuggestion',

                addProseMirrorPlugins() {
                    return [
                        Suggestion({
                            ...suggestionConfig,
                            editor: this.editor,
                        }),
                    ];
                },
            });

            base.push(WikiLinkSuggestionExtension);
        }

        return base;
    }, [campaignId, placeholder]);

    const editor = useEditor({
        extensions,
        content: value,
        editable: !disabled,
        onUpdate: ({ editor }) => {
            const storage = editor.storage as unknown as MarkdownStorage;
            const markdown = storage.markdown.getMarkdown();
            lastEditorValueRef.current = markdown;
            onChangeRef.current(markdown);
        },
    });

    // Update editor content when value prop changes externally.
    // Compares the incoming value against what the editor last
    // reported to avoid re-setting content the editor already has
    // (which would reset the cursor position).
    useEffect(() => {
        if (!editor) return;

        // If the parent is reflecting back the same value the editor
        // last produced, there is nothing to do.
        if (value === lastEditorValueRef.current) return;

        lastEditorValueRef.current = value;
        editor.commands.setContent(value || '');
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
                    campaignId={campaignId}
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
