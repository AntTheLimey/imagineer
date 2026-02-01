// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useEffect } from 'react';
import {
    Box,
    FormHelperText,
    InputLabel,
    Paper,
} from '@mui/material';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import EditorToolbar from './EditorToolbar';

/**
 * Props for the RichTextEditor component.
 */
export interface RichTextEditorProps {
    /** The HTML content of the editor. */
    value: string;
    /** Callback fired when the content changes. */
    onChange: (html: string) => void;
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
 * A MUI-styled rich text editor component wrapping TipTap.
 *
 * Provides a WYSIWYG editing experience with formatting options including
 * bold, italic, strikethrough, headings, lists, and undo/redo. The editor
 * is styled to match MUI TextField aesthetics with support for labels,
 * error states, and helper text.
 *
 * @param props - The component props.
 * @returns A React element containing the rich text editor.
 *
 * @example
 * ```tsx
 * const [content, setContent] = useState('<p>Initial content</p>');
 *
 * <RichTextEditor
 *     value={content}
 *     onChange={setContent}
 *     label="Description"
 *     placeholder="Enter a description..."
 * />
 * ```
 */
export default function RichTextEditor({
    value,
    onChange,
    placeholder,
    label,
    error = false,
    helperText,
    minHeight = 200,
    maxHeight = 400,
    disabled = false,
}: RichTextEditorProps) {
    const editor = useEditor({
        extensions: [
            StarterKit,
            Placeholder.configure({
                placeholder: placeholder ?? '',
            }),
        ],
        content: value,
        editable: !disabled,
        onUpdate: ({ editor }) => {
            onChange(editor.getHTML());
        },
    });

    // Update editor content when value prop changes externally
    useEffect(() => {
        if (editor && value !== editor.getHTML()) {
            editor.commands.setContent(value);
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
                        color: error ? 'error.main' : 'text.secondary',
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
                        borderColor: error ? 'error.main' : 'primary.main',
                        borderWidth: 2,
                    },
                    opacity: disabled ? 0.5 : 1,
                }}
            >
                <EditorToolbar editor={editor} disabled={disabled} />
                <Box
                    sx={{
                        p: 2,
                        minHeight,
                        maxHeight,
                        overflowY: 'auto',
                        '& .tiptap': {
                            outline: 'none',
                            minHeight: minHeight - 32, // Account for padding
                            '& p.is-editor-empty:first-child::before': {
                                content: 'attr(data-placeholder)',
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
