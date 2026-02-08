# Markdown Editor (Phase 1) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the HTML-based RichTextEditor with a
Markdown-native editor and renderer, eliminating HTML from
the content pipeline entirely.

**Architecture:** TipTap remains the editor framework but
serializes to Markdown via the `tiptap-markdown` extension.
A `MarkdownRenderer` component wraps `react-markdown` for
read-mode display. All content flows as Markdown: editor,
API, database, vectorizer.

**Tech Stack:** TipTap 2.x, tiptap-markdown, react-markdown,
React 18, MUI 5, TypeScript, Vitest

---

### Task 1: Create feature branch and update dependencies

**Files:**

- Modify: `client/package.json`

**Step 1: Create branch**

```bash
git checkout -b feature/markdown-editor
```

**Step 2: Install new dependencies**

```bash
cd client
npm install tiptap-markdown react-markdown
```

**Step 3: Uninstall old dependencies**

```bash
cd client
npm uninstall dompurify @types/dompurify
```

**Step 4: Verify package.json**

Run:
`grep -E "tiptap-markdown|react-markdown|dompurify" client/package.json`

Expected: `tiptap-markdown` and `react-markdown` in
dependencies. No `dompurify` entries.

**Step 5: Commit**

```bash
git add client/package.json client/package-lock.json
git commit -m "chore: add tiptap-markdown and react-markdown, remove dompurify"
```

---

### Task 2: Create MarkdownEditor component

**Files:**

- Create:
  `client/src/components/MarkdownEditor/MarkdownEditor.tsx`
- Create:
  `client/src/components/MarkdownEditor/EditorToolbar.tsx`
- Create:
  `client/src/components/MarkdownEditor/index.ts`

**Step 1: Write MarkdownEditor.tsx**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

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
import { Markdown } from 'tiptap-markdown';
import EditorToolbar from './EditorToolbar';

export interface MarkdownEditorProps {
    value: string;
    onChange: (markdown: string) => void;
    placeholder?: string;
    label?: string;
    error?: boolean;
    helperText?: string;
    minHeight?: number;
    maxHeight?: number;
    disabled?: boolean;
}

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
        onUpdate: ({ editor: e }) => {
            const md =
                e.storage.markdown.getMarkdown();
            onChange(md);
        },
    });

    useEffect(() => {
        if (!editor) return;
        const current =
            editor.storage.markdown.getMarkdown();
        if (value !== current) {
            editor.commands.setContent(value);
        }
    }, [value, editor]);

    useEffect(() => {
        if (!editor) return;
        editor.setEditable(!disabled);
    }, [disabled, editor]);

    return (
        <Box>
            {label && (
                <InputLabel
                    error={error}
                    sx={{ mb: 1 }}
                >
                    {label}
                </InputLabel>
            )}
            <Paper
                variant="outlined"
                sx={{
                    border: error
                        ? '1px solid error.main'
                        : undefined,
                    '& .ProseMirror': {
                        minHeight,
                        maxHeight,
                        overflowY: 'auto',
                        padding: 2,
                        outline: 'none',
                        '& p.is-editor-empty:first-of-type::before': {
                            color: 'text.disabled',
                            content: 'attr(data-placeholder)',
                            float: 'left',
                            height: 0,
                            pointerEvents: 'none',
                        },
                    },
                }}
            >
                <EditorToolbar
                    editor={editor}
                    disabled={disabled}
                />
                <EditorContent editor={editor} />
            </Paper>
            {helperText && (
                <FormHelperText error={error}>
                    {helperText}
                </FormHelperText>
            )}
        </Box>
    );
}
```

**Step 2: Write EditorToolbar.tsx**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

import {
    Box,
    Divider,
    IconButton,
    Tooltip,
} from '@mui/material';
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
import type { Editor } from '@tiptap/react';

interface EditorToolbarProps {
    editor: Editor | null;
    disabled?: boolean;
}

export default function EditorToolbar({
    editor,
    disabled = false,
}: EditorToolbarProps) {
    if (!editor) return null;

    const btn = (
        label: string,
        icon: React.ReactNode,
        action: () => void,
        isActive?: boolean,
    ) => (
        <Tooltip title={label}>
            <span>
                <IconButton
                    size="small"
                    onClick={action}
                    disabled={disabled}
                    color={
                        isActive
                            ? 'primary'
                            : 'default'
                    }
                >
                    {icon}
                </IconButton>
            </span>
        </Tooltip>
    );

    return (
        <Box
            sx={{
                display: 'flex',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: 0.5,
                px: 1,
                py: 0.5,
                borderBottom: 1,
                borderColor: 'divider',
            }}
        >
            {btn(
                'Bold',
                <FormatBold />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleBold()
                        .run(),
                editor.isActive('bold'),
            )}
            {btn(
                'Italic',
                <FormatItalic />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleItalic()
                        .run(),
                editor.isActive('italic'),
            )}
            {btn(
                'Strikethrough',
                <FormatStrikethrough />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleStrike()
                        .run(),
                editor.isActive('strike'),
            )}
            <Divider
                orientation="vertical"
                flexItem
            />
            {btn(
                'Heading 2',
                <Box
                    component="span"
                    sx={{
                        fontSize: '0.85rem',
                        fontWeight: 'bold',
                    }}
                >
                    H2
                </Box>,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleHeading({ level: 2 })
                        .run(),
                editor.isActive('heading', {
                    level: 2,
                }),
            )}
            {btn(
                'Heading 3',
                <Box
                    component="span"
                    sx={{
                        fontSize: '0.8rem',
                        fontWeight: 'bold',
                    }}
                >
                    H3
                </Box>,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleHeading({ level: 3 })
                        .run(),
                editor.isActive('heading', {
                    level: 3,
                }),
            )}
            <Divider
                orientation="vertical"
                flexItem
            />
            {btn(
                'Bullet List',
                <FormatListBulleted />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleBulletList()
                        .run(),
                editor.isActive('bulletList'),
            )}
            {btn(
                'Ordered List',
                <FormatListNumbered />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleOrderedList()
                        .run(),
                editor.isActive('orderedList'),
            )}
            <Divider
                orientation="vertical"
                flexItem
            />
            {btn(
                'Blockquote',
                <FormatQuote />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .toggleBlockquote()
                        .run(),
                editor.isActive('blockquote'),
            )}
            {btn(
                'Horizontal Rule',
                <HorizontalRule />,
                () =>
                    editor
                        .chain()
                        .focus()
                        .setHorizontalRule()
                        .run(),
            )}
            <Divider
                orientation="vertical"
                flexItem
            />
            {btn('Undo', <Undo />, () =>
                editor
                    .chain()
                    .focus()
                    .undo()
                    .run(),
            )}
            {btn('Redo', <Redo />, () =>
                editor
                    .chain()
                    .focus()
                    .redo()
                    .run(),
            )}
        </Box>
    );
}
```

**Step 3: Write index.ts**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

export { default } from './MarkdownEditor';
export { default as MarkdownEditor } from './MarkdownEditor';
export type { MarkdownEditorProps } from './MarkdownEditor';
```

**Step 4: Verify TypeScript compiles**

Run: `cd client && npx tsc --noEmit 2>&1 | head -20`

Expected: No errors related to MarkdownEditor files.
(Pre-existing errors in ChapterEditorPage.tsx are OK.)

**Step 5: Commit**

```bash
git add client/src/components/MarkdownEditor/
git commit -m "feat: add MarkdownEditor component with TipTap Markdown serialization"
```

---

### Task 3: Create MarkdownRenderer component

**Files:**

- Create:
  `client/src/components/MarkdownRenderer/MarkdownRenderer.tsx`
- Create:
  `client/src/components/MarkdownRenderer/index.ts`

**Step 1: Write MarkdownRenderer.tsx**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

import { Box } from '@mui/material';
import ReactMarkdown from 'react-markdown';

interface MarkdownRendererProps {
    content: string | null | undefined;
    maxLines?: number;
}

export default function MarkdownRenderer({
    content,
    maxLines,
}: MarkdownRendererProps) {
    if (!content) return null;

    return (
        <Box
            sx={{
                '& h1': { fontSize: '1.5rem', mt: 2, mb: 1 },
                '& h2': { fontSize: '1.25rem', mt: 2, mb: 1 },
                '& h3': { fontSize: '1.1rem', mt: 1.5, mb: 0.5 },
                '& p': { mt: 0, mb: 1 },
                '& ul, & ol': { mt: 0, mb: 1, pl: 3 },
                '& blockquote': {
                    borderLeft: 3,
                    borderColor: 'divider',
                    pl: 2,
                    ml: 0,
                    color: 'text.secondary',
                    fontStyle: 'italic',
                },
                '& hr': { my: 2, borderColor: 'divider' },
                '& code': {
                    backgroundColor: 'action.hover',
                    px: 0.5,
                    borderRadius: 0.5,
                    fontSize: '0.875em',
                },
                ...(maxLines
                    ? {
                          display: '-webkit-box',
                          WebkitLineClamp: maxLines,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                      }
                    : {}),
            }}
        >
            <ReactMarkdown
                allowedElements={[
                    'p', 'strong', 'em', 'del',
                    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
                    'ul', 'ol', 'li',
                    'blockquote', 'hr', 'br',
                    'code', 'pre',
                ]}
                unwrapDisallowed
            >
                {content}
            </ReactMarkdown>
        </Box>
    );
}
```

**Step 2: Write index.ts**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

export { default } from './MarkdownRenderer';
export { default as MarkdownRenderer } from './MarkdownRenderer';
```

**Step 3: Verify TypeScript compiles**

Run: `cd client && npx tsc --noEmit 2>&1 | head -20`

Expected: No new errors.

**Step 4: Commit**

```bash
git add client/src/components/MarkdownRenderer/
git commit -m "feat: add MarkdownRenderer component for Markdown display"
```

---

### Task 4: Update pages to use MarkdownEditor

**Files:**

- Modify: `client/src/pages/CampaignOverview.tsx`
- Modify: `client/src/pages/EntityEditor.tsx`
- Modify: `client/src/pages/ChapterEditorPage.tsx`
- Modify: `client/src/pages/CreateCampaign.tsx`
- Modify:
  `client/src/components/CampaignSettings/CampaignSettings.tsx`

For each file, make two changes:

1. Replace the import:
   - Old:
     `import { RichTextEditor } from '../components/RichTextEditor';`
   - New:
     `import { MarkdownEditor } from '../components/MarkdownEditor';`

2. Replace the component usage:
   - Old: `<RichTextEditor` ...
     `onChange={(html) => updateField('description', html)}`
     ... `/>`
   - New: `<MarkdownEditor` ...
     `onChange={(md) => updateField('description', md)}`
     ... `/>`

The prop interface is identical, so only the component
name and import path change. The `html` callback parameter
should be renamed to `md` for clarity.

**Step 1: Update CampaignOverview.tsx**

Change line 44 import from:

```typescript
import { RichTextEditor } from '../components/RichTextEditor';
```

to:

```typescript
import { MarkdownEditor } from '../components/MarkdownEditor';
```

Change the component at lines 510-517 from
`<RichTextEditor` to `<MarkdownEditor`, rename the
onChange param from `html` to `md`:

```tsx
<MarkdownEditor
    label=""
    value={formData.description}
    onChange={(md) => updateField('description', md)}
    placeholder="Describe your campaign setting, themes, and background..."
    minHeight={150}
    maxHeight={300}
/>
```

**Step 2: Update EntityEditor.tsx**

Change line 34 import from:

```typescript
import { RichTextEditor } from '../components/RichTextEditor';
```

to:

```typescript
import { MarkdownEditor } from '../components/MarkdownEditor';
```

Change the component at lines 592-599 from
`<RichTextEditor` to `<MarkdownEditor`, rename `html`
to `md`:

```tsx
<MarkdownEditor
    label="Description"
    value={formData.description}
    onChange={(md) => updateField('description', md)}
    placeholder="Describe this entity..."
    error={!!formErrors.description}
    helperText={formErrors.description}
/>
```

**Step 3: Update ChapterEditorPage.tsx**

Change line 32 import from:

```typescript
import { RichTextEditor } from '../components/RichTextEditor';
```

to:

```typescript
import { MarkdownEditor } from '../components/MarkdownEditor';
```

Change the component at lines 526-531 from
`<RichTextEditor` to `<MarkdownEditor`:

```tsx
<MarkdownEditor
    value={formData.overview}
    onChange={(md) => updateField('overview', md)}
    placeholder="Describe this chapter's story arc, themes, and key events..."
    minHeight={300}
/>
```

**Step 4: Update CreateCampaign.tsx**

Change line 34 import from:

```typescript
import { RichTextEditor } from '../components/RichTextEditor';
```

to:

```typescript
import { MarkdownEditor } from '../components/MarkdownEditor';
```

Change the component at lines 293-300 from
`<RichTextEditor` to `<MarkdownEditor`, rename `html`
to `md`:

```tsx
<MarkdownEditor
    label="Description"
    value={formData.description}
    onChange={(md) => updateField('description', md)}
    placeholder="Describe your campaign setting, themes, and background..."
    minHeight={150}
    maxHeight={300}
/>
```

**Step 5: Update CampaignSettings.tsx**

Change line 31 import from:

```typescript
import { RichTextEditor } from '../RichTextEditor';
```

to:

```typescript
import { MarkdownEditor } from '../MarkdownEditor';
```

Change the component at lines 263-270 from
`<RichTextEditor` to `<MarkdownEditor`, rename `html`
to `md`:

```tsx
<MarkdownEditor
    label="Description"
    value={formData.description}
    onChange={(md) => updateField('description', md)}
    placeholder="Describe your campaign setting, themes, and background..."
    minHeight={150}
    maxHeight={300}
/>
```

**Step 6: Verify TypeScript compiles**

Run: `cd client && npx tsc --noEmit 2>&1 | head -20`

Expected: No new errors related to editor imports.

**Step 7: Commit**

```bash
git add client/src/pages/CampaignOverview.tsx \
        client/src/pages/EntityEditor.tsx \
        client/src/pages/ChapterEditorPage.tsx \
        client/src/pages/CreateCampaign.tsx \
        client/src/components/CampaignSettings/CampaignSettings.tsx
git commit -m "refactor: replace RichTextEditor with MarkdownEditor across all pages"
```

---

### Task 5: Replace sanitizeHtml and dangerouslySetInnerHTML with MarkdownRenderer

**Files:**

- Modify: `client/src/pages/CampaignOverview.tsx`
- Modify: `client/src/pages/Entities.tsx`
- Modify:
  `client/src/components/EntityPreviewPanel/EntityPreviewPanel.tsx`

**Step 1: Update CampaignOverview.tsx**

Remove the sanitizeHtml import (line 53):

```typescript
import { sanitizeHtml } from '../utils';
```

Add MarkdownRenderer import:

```typescript
import { MarkdownRenderer } from '../components/MarkdownRenderer';
```

Replace the dangerouslySetInnerHTML block (lines 545-547):

```tsx
dangerouslySetInnerHTML={{
    __html: sanitizeHtml(campaign.description),
}}
```

with:

```tsx
<MarkdownRenderer content={campaign.description} />
```

Note: the parent element that had
`dangerouslySetInnerHTML` can no longer have that prop.
Change the element from a self-closing tag to a container
with `MarkdownRenderer` as a child.

**Step 2: Update Entities.tsx**

Remove the sanitizeHtml/stripHtml import (line 58):

```typescript
import { sanitizeHtml, stripHtml } from '../utils';
```

Add MarkdownRenderer import:

```typescript
import { MarkdownRenderer } from '../components/MarkdownRenderer';
```

Replace the stripHtml usage (line 584):

```tsx
{entity.description ? stripHtml(entity.description) : '-'}
```

with (Markdown is already readable as plain text, but
truncate for table display):

```tsx
{entity.description
    ? entity.description.slice(0, 100) +
      (entity.description.length > 100 ? '...' : '')
    : '-'}
```

Replace the dangerouslySetInnerHTML block
(lines 905-907):

```tsx
dangerouslySetInnerHTML={{
    __html: sanitizeHtml(dialogEntity.description),
}}
```

with MarkdownRenderer as a child element:

```tsx
<MarkdownRenderer
    content={dialogEntity.description}
/>
```

**Step 3: Update EntityPreviewPanel.tsx**

Remove the sanitizeHtml import (line 23):

```typescript
import { sanitizeHtml } from '../../utils';
```

Add MarkdownRenderer import:

```typescript
import { MarkdownRenderer } from '../MarkdownRenderer';
```

Replace the dangerouslySetInnerHTML block
(lines 208-210):

```tsx
dangerouslySetInnerHTML={{
    __html: sanitizeHtml(entity.description),
}}
```

with MarkdownRenderer as a child, using the maxLines prop
for line clamping (replaces WebkitLineClamp):

```tsx
<MarkdownRenderer
    content={entity.description}
    maxLines={6}
/>
```

**Step 4: Verify TypeScript compiles**

Run: `cd client && npx tsc --noEmit 2>&1 | head -20`

Expected: No new errors. No remaining references to
sanitizeHtml or dangerouslySetInnerHTML.

**Step 5: Verify no remaining references**

Run:
`grep -r "sanitizeHtml\|stripHtml\|dangerouslySetInnerHTML" client/src/ --include="*.tsx" --include="*.ts" | grep -v node_modules`

Expected: Only hits in `utils/sanitizeHtml.ts` itself
(to be removed in the next task).

**Step 6: Commit**

```bash
git add client/src/pages/CampaignOverview.tsx \
        client/src/pages/Entities.tsx \
        client/src/components/EntityPreviewPanel/EntityPreviewPanel.tsx
git commit -m "refactor: replace sanitizeHtml/dangerouslySetInnerHTML with MarkdownRenderer"
```

---

### Task 6: Remove old RichTextEditor and sanitizeHtml

**Files:**

- Delete:
  `client/src/components/RichTextEditor/RichTextEditor.tsx`
- Delete:
  `client/src/components/RichTextEditor/EditorToolbar.tsx`
- Delete:
  `client/src/components/RichTextEditor/index.ts`
- Delete: `client/src/utils/sanitizeHtml.ts`
- Modify: `client/src/utils/index.ts`
  (remove sanitizeHtml export)

**Step 1: Delete RichTextEditor directory**

```bash
rm -rf client/src/components/RichTextEditor/
```

**Step 2: Delete sanitizeHtml.ts**

```bash
rm client/src/utils/sanitizeHtml.ts
```

**Step 3: Update utils/index.ts**

Remove line 14:

```typescript
export { sanitizeHtml, stripHtml } from './sanitizeHtml';
```

**Step 4: Verify no remaining imports of old code**

Run:
`grep -r "RichTextEditor\|sanitizeHtml\|stripHtml\|dompurify" client/src/ --include="*.tsx" --include="*.ts" | grep -v node_modules`

Expected: No results.

**Step 5: Verify TypeScript compiles**

Run: `cd client && npx tsc --noEmit 2>&1 | head -20`

Expected: No new errors.

**Step 6: Commit**

```bash
git add -A client/src/components/RichTextEditor/ \
        client/src/utils/sanitizeHtml.ts \
        client/src/utils/index.ts
git commit -m "chore: remove RichTextEditor, sanitizeHtml, and dompurify"
```

---

### Task 7: Write MarkdownEditor tests

**Files:**

- Create:
  `client/src/components/MarkdownEditor/MarkdownEditor.test.tsx`

**Step 1: Write test file**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import MarkdownEditor from './MarkdownEditor';

describe('MarkdownEditor', () => {
    it('renders without crashing', () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
            />,
        );
        expect(
            document.querySelector('.ProseMirror'),
        ).toBeTruthy();
    });

    it('renders with initial Markdown value', () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value="**bold text**"
                onChange={onChange}
            />,
        );
        const editor =
            document.querySelector('.ProseMirror');
        expect(editor?.innerHTML).toContain(
            '<strong>',
        );
    });

    it('displays label when provided', () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                label="Description"
            />,
        );
        expect(
            screen.getByText('Description'),
        ).toBeTruthy();
    });

    it('displays placeholder when empty', () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                placeholder="Enter text..."
            />,
        );
        const placeholder =
            document.querySelector(
                '.is-editor-empty',
            );
        expect(placeholder).toBeTruthy();
    });

    it('displays helper text when provided', () => {
        const onChange = vi.fn();
        render(
            <MarkdownEditor
                value=""
                onChange={onChange}
                helperText="Required field"
            />,
        );
        expect(
            screen.getByText('Required field'),
        ).toBeTruthy();
    });

    it('handles null-ish value gracefully', () => {
        const onChange = vi.fn();
        expect(() =>
            render(
                <MarkdownEditor
                    value=""
                    onChange={onChange}
                />,
            ),
        ).not.toThrow();
    });
});
```

**Step 2: Run tests**

Run:
`cd client && npx vitest run src/components/MarkdownEditor/MarkdownEditor.test.tsx`

Expected: All tests pass.

**Step 3: Commit**

```bash
git add client/src/components/MarkdownEditor/MarkdownEditor.test.tsx
git commit -m "test: add MarkdownEditor component tests"
```

---

### Task 8: Write MarkdownRenderer tests

**Files:**

- Create:
  `client/src/components/MarkdownRenderer/MarkdownRenderer.test.tsx`

**Step 1: Write test file**

```tsx
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MarkdownRenderer from './MarkdownRenderer';

describe('MarkdownRenderer', () => {
    it('renders null content without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content={null} />,
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders undefined content without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content={undefined} />,
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders empty string without crashing', () => {
        const { container } = render(
            <MarkdownRenderer content="" />,
        );
        expect(container.innerHTML).toBe('');
    });

    it('renders bold text', () => {
        render(
            <MarkdownRenderer
                content="**bold text**"
            />,
        );
        const strong =
            document.querySelector('strong');
        expect(strong?.textContent).toBe('bold text');
    });

    it('renders italic text', () => {
        render(
            <MarkdownRenderer
                content="*italic text*"
            />,
        );
        const em = document.querySelector('em');
        expect(em?.textContent).toBe('italic text');
    });

    it('renders headings', () => {
        render(
            <MarkdownRenderer
                content="## Section Title"
            />,
        );
        const h2 = document.querySelector('h2');
        expect(h2?.textContent).toBe('Section Title');
    });

    it('renders bullet lists', () => {
        render(
            <MarkdownRenderer
                content="- item one\n- item two"
            />,
        );
        const items =
            document.querySelectorAll('li');
        expect(items.length).toBe(2);
    });

    it('renders blockquotes', () => {
        render(
            <MarkdownRenderer
                content="> quoted text"
            />,
        );
        const bq =
            document.querySelector('blockquote');
        expect(bq?.textContent).toContain(
            'quoted text',
        );
    });

    it('renders horizontal rules', () => {
        render(
            <MarkdownRenderer
                content="above\n\n---\n\nbelow"
            />,
        );
        const hr = document.querySelector('hr');
        expect(hr).toBeTruthy();
    });

    it('blocks raw HTML tags', () => {
        render(
            <MarkdownRenderer
                content='<script>alert("xss")</script>'
            />,
        );
        const script =
            document.querySelector('script');
        expect(script).toBeNull();
    });
});
```

**Step 2: Run tests**

Run:
`cd client && npx vitest run src/components/MarkdownRenderer/MarkdownRenderer.test.tsx`

Expected: All tests pass.

**Step 3: Commit**

```bash
git add client/src/components/MarkdownRenderer/MarkdownRenderer.test.tsx
git commit -m "test: add MarkdownRenderer component tests"
```

---

### Task 9: Run full test suite and verify

**Step 1: Run client tests**

Run: `cd client && npx vitest run`

Expected: All tests pass. No regressions.

**Step 2: Run TypeScript check**

Run: `cd client && npx tsc --noEmit`

Expected: No new errors (pre-existing ChapterEditorPage
errors are acceptable).

**Step 3: Verify no HTML artifacts remain**

Run:
`grep -r "dangerouslySetInnerHTML\|sanitizeHtml\|stripHtml\|dompurify\|DOMPurify\|RichTextEditor" client/src/ --include="*.ts" --include="*.tsx" | grep -v node_modules | grep -v ".test."`

Expected: No results.

**Step 4: Run Go tests to verify no backend changes
broke anything**

Run:
`cd /Users/antonypegg/PROJECTS/imagineer && go test ./...`

Expected: All tests pass (no backend changes in this
phase).

---

END OF PLAN.
