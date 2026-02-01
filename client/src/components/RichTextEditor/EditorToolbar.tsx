// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { Box, Divider, IconButton, ToggleButton } from '@mui/material';
import {
    FormatBold,
    FormatItalic,
    FormatStrikethrough,
    FormatListBulleted,
    FormatListNumbered,
    Undo,
    Redo,
} from '@mui/icons-material';
import { Editor } from '@tiptap/react';

/**
 * Props for the EditorToolbar component.
 */
interface EditorToolbarProps {
    /** The TipTap editor instance. */
    editor: Editor | null;
    /** Whether the toolbar should be disabled. */
    disabled?: boolean;
}

/**
 * A MUI-styled toolbar for the RichTextEditor component.
 *
 * Provides formatting controls including bold, italic, strikethrough,
 * headings (H1, H2, H3), bullet and ordered lists, and undo/redo.
 * Buttons reflect the active state of the current selection.
 *
 * @param props - The component props.
 * @returns A React element containing the editor toolbar.
 */
export default function EditorToolbar({
    editor,
    disabled = false,
}: EditorToolbarProps) {
    if (!editor) {
        return null;
    }

    return (
        <Box
            sx={{
                display: 'flex',
                flexWrap: 'wrap',
                alignItems: 'center',
                gap: 0.5,
                p: 1,
                borderBottom: 1,
                borderColor: 'divider',
            }}
        >
            {/* Text formatting */}
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().toggleBold().run()}
                disabled={disabled || !editor.can().chain().focus().toggleBold().run()}
                color={editor.isActive('bold') ? 'primary' : 'default'}
                aria-label="Bold"
            >
                <FormatBold fontSize="small" />
            </IconButton>
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().toggleItalic().run()}
                disabled={disabled || !editor.can().chain().focus().toggleItalic().run()}
                color={editor.isActive('italic') ? 'primary' : 'default'}
                aria-label="Italic"
            >
                <FormatItalic fontSize="small" />
            </IconButton>
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().toggleStrike().run()}
                disabled={disabled || !editor.can().chain().focus().toggleStrike().run()}
                color={editor.isActive('strike') ? 'primary' : 'default'}
                aria-label="Strikethrough"
            >
                <FormatStrikethrough fontSize="small" />
            </IconButton>

            <Divider orientation="vertical" flexItem sx={{ mx: 0.5 }} />

            {/* Headings */}
            <ToggleButton
                size="small"
                value="h1"
                selected={editor.isActive('heading', { level: 1 })}
                onChange={() =>
                    editor.chain().focus().toggleHeading({ level: 1 }).run()
                }
                disabled={disabled}
                sx={{ px: 1, minWidth: 'auto' }}
                aria-label="Heading 1"
            >
                H1
            </ToggleButton>
            <ToggleButton
                size="small"
                value="h2"
                selected={editor.isActive('heading', { level: 2 })}
                onChange={() =>
                    editor.chain().focus().toggleHeading({ level: 2 }).run()
                }
                disabled={disabled}
                sx={{ px: 1, minWidth: 'auto' }}
                aria-label="Heading 2"
            >
                H2
            </ToggleButton>
            <ToggleButton
                size="small"
                value="h3"
                selected={editor.isActive('heading', { level: 3 })}
                onChange={() =>
                    editor.chain().focus().toggleHeading({ level: 3 }).run()
                }
                disabled={disabled}
                sx={{ px: 1, minWidth: 'auto' }}
                aria-label="Heading 3"
            >
                H3
            </ToggleButton>

            <Divider orientation="vertical" flexItem sx={{ mx: 0.5 }} />

            {/* Lists */}
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().toggleBulletList().run()}
                disabled={disabled}
                color={editor.isActive('bulletList') ? 'primary' : 'default'}
                aria-label="Bullet list"
            >
                <FormatListBulleted fontSize="small" />
            </IconButton>
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().toggleOrderedList().run()}
                disabled={disabled}
                color={editor.isActive('orderedList') ? 'primary' : 'default'}
                aria-label="Ordered list"
            >
                <FormatListNumbered fontSize="small" />
            </IconButton>

            <Divider orientation="vertical" flexItem sx={{ mx: 0.5 }} />

            {/* History */}
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().undo().run()}
                disabled={disabled || !editor.can().chain().focus().undo().run()}
                aria-label="Undo"
            >
                <Undo fontSize="small" />
            </IconButton>
            <IconButton
                size="small"
                onClick={() => editor.chain().focus().redo().run()}
                disabled={disabled || !editor.can().chain().focus().redo().run()}
                aria-label="Redo"
            >
                <Redo fontSize="small" />
            </IconButton>
        </Box>
    );
}
