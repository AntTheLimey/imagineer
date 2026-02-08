/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { Box, Divider, IconButton, Tooltip } from '@mui/material';
import {
    FormatBold,
    FormatItalic,
    FormatStrikethrough,
    FormatListBulleted,
    FormatListNumbered,
    FormatQuote,
    HorizontalRule,
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
 * A MUI-styled toolbar for the MarkdownEditor component.
 *
 * Provides formatting controls including bold, italic, strikethrough,
 * headings (H2, H3), bullet and ordered lists, blockquote, horizontal
 * rule, and undo/redo. Buttons reflect the active state of the current
 * selection via primary color highlighting.
 *
 * @param props - The component props.
 * @returns A React element containing the editor toolbar, or null if
 *     no editor instance is available.
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
            <Tooltip title="Bold">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor.chain().focus().toggleBold().run()
                    }
                    disabled={
                        disabled ||
                        !editor
                            .can()
                            .chain()
                            .focus()
                            .toggleBold()
                            .run()
                    }
                    color={
                        editor.isActive('bold') ? 'primary' : 'default'
                    }
                    aria-label="Bold"
                >
                    <FormatBold fontSize="small" />
                </IconButton>
            </Tooltip>
            <Tooltip title="Italic">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor.chain().focus().toggleItalic().run()
                    }
                    disabled={
                        disabled ||
                        !editor
                            .can()
                            .chain()
                            .focus()
                            .toggleItalic()
                            .run()
                    }
                    color={
                        editor.isActive('italic')
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Italic"
                >
                    <FormatItalic fontSize="small" />
                </IconButton>
            </Tooltip>
            <Tooltip title="Strikethrough">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor.chain().focus().toggleStrike().run()
                    }
                    disabled={
                        disabled ||
                        !editor
                            .can()
                            .chain()
                            .focus()
                            .toggleStrike()
                            .run()
                    }
                    color={
                        editor.isActive('strike')
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Strikethrough"
                >
                    <FormatStrikethrough fontSize="small" />
                </IconButton>
            </Tooltip>

            <Divider
                orientation="vertical"
                flexItem
                sx={{ mx: 0.5 }}
            />

            {/* Headings */}
            <Tooltip title="Heading 2">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .toggleHeading({ level: 2 })
                            .run()
                    }
                    disabled={disabled}
                    color={
                        editor.isActive('heading', { level: 2 })
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Heading 2"
                    sx={{
                        fontWeight: 'bold',
                        fontSize: '0.8rem',
                        minWidth: 32,
                    }}
                >
                    H2
                </IconButton>
            </Tooltip>
            <Tooltip title="Heading 3">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .toggleHeading({ level: 3 })
                            .run()
                    }
                    disabled={disabled}
                    color={
                        editor.isActive('heading', { level: 3 })
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Heading 3"
                    sx={{
                        fontWeight: 'bold',
                        fontSize: '0.8rem',
                        minWidth: 32,
                    }}
                >
                    H3
                </IconButton>
            </Tooltip>

            <Divider
                orientation="vertical"
                flexItem
                sx={{ mx: 0.5 }}
            />

            {/* Lists */}
            <Tooltip title="Bullet List">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .toggleBulletList()
                            .run()
                    }
                    disabled={disabled}
                    color={
                        editor.isActive('bulletList')
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Bullet list"
                >
                    <FormatListBulleted fontSize="small" />
                </IconButton>
            </Tooltip>
            <Tooltip title="Ordered List">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .toggleOrderedList()
                            .run()
                    }
                    disabled={disabled}
                    color={
                        editor.isActive('orderedList')
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Ordered list"
                >
                    <FormatListNumbered fontSize="small" />
                </IconButton>
            </Tooltip>

            <Divider
                orientation="vertical"
                flexItem
                sx={{ mx: 0.5 }}
            />

            {/* Block elements */}
            <Tooltip title="Blockquote">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .toggleBlockquote()
                            .run()
                    }
                    disabled={disabled}
                    color={
                        editor.isActive('blockquote')
                            ? 'primary'
                            : 'default'
                    }
                    aria-label="Blockquote"
                >
                    <FormatQuote fontSize="small" />
                </IconButton>
            </Tooltip>
            <Tooltip title="Horizontal Rule">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor
                            .chain()
                            .focus()
                            .setHorizontalRule()
                            .run()
                    }
                    disabled={disabled}
                    aria-label="Horizontal rule"
                >
                    <HorizontalRule fontSize="small" />
                </IconButton>
            </Tooltip>

            <Divider
                orientation="vertical"
                flexItem
                sx={{ mx: 0.5 }}
            />

            {/* History */}
            <Tooltip title="Undo">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor.chain().focus().undo().run()
                    }
                    disabled={
                        disabled ||
                        !editor
                            .can()
                            .chain()
                            .focus()
                            .undo()
                            .run()
                    }
                    aria-label="Undo"
                >
                    <Undo fontSize="small" />
                </IconButton>
            </Tooltip>
            <Tooltip title="Redo">
                <IconButton
                    size="small"
                    onClick={() =>
                        editor.chain().focus().redo().run()
                    }
                    disabled={
                        disabled ||
                        !editor
                            .can()
                            .chain()
                            .focus()
                            .redo()
                            .run()
                    }
                    aria-label="Redo"
                >
                    <Redo fontSize="small" />
                </IconButton>
            </Tooltip>
        </Box>
    );
}
