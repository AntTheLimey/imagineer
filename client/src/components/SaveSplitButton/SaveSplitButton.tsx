// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useEffect, useRef, useState } from 'react';
import {
    Button,
    ButtonGroup,
    ClickAwayListener,
    CircularProgress,
    Grow,
    MenuItem,
    MenuList,
    Paper,
    Popper,
} from '@mui/material';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';

/**
 * The save mode determines which pipeline stages run after saving.
 *
 * - "save" saves content without any automated analysis.
 * - "analyze" saves and then runs phase 1 analysis.
 * - "enrich" saves and then runs both phase 1 analysis and phase 2
 *   enrichment.
 */
export type SaveMode = 'save' | 'analyze' | 'enrich';

/**
 * Props for the SaveSplitButton component.
 */
export interface SaveSplitButtonProps {
    /** Callback fired when the user clicks the primary button. */
    onSave: (mode: SaveMode) => void;
    /** Whether the editor content has unsaved changes. */
    isDirty: boolean;
    /** Whether a save operation is currently in progress. */
    isSaving: boolean;
    /** Optional default save mode to pre-select instead of "Save & Analyze". */
    defaultMode?: SaveMode;
}

/**
 * A menu option linking a display label to a save mode.
 */
interface SaveOption {
    label: string;
    mode: SaveMode;
}

const options: SaveOption[] = [
    { label: 'Save', mode: 'save' },
    { label: 'Save & Analyze', mode: 'analyze' },
    { label: 'Save, Analyze & Enrich', mode: 'enrich' },
];

/**
 * The default selected index. Index 1 corresponds to "Save & Analyze",
 * which matches the current auto-analyze behavior.
 */
const DEFAULT_SELECTED_INDEX = 1;

/**
 * A split button offering three save modes: save only, save and analyze,
 * or save, analyze, and enrich. The primary button reflects the most
 * recently selected option and persists for the session. A dropdown arrow
 * reveals a menu with all three options.
 *
 * The button is disabled when there are no unsaved changes or when a save
 * operation is in progress. A spinner replaces the label while saving.
 *
 * @param props - The component props.
 * @returns A React element containing the split button group and dropdown.
 *
 * @example
 * ```tsx
 * <SaveSplitButton
 *     onSave={(mode) => handleSave(mode)}
 *     isDirty={hasUnsavedChanges}
 *     isSaving={mutation.isLoading}
 * />
 * ```
 */
export default function SaveSplitButton({
    onSave,
    isDirty,
    isSaving,
    defaultMode,
}: SaveSplitButtonProps) {
    const [open, setOpen] = useState(false);
    const initialIndex = defaultMode
        ? options.findIndex(o => o.mode === defaultMode)
        : DEFAULT_SELECTED_INDEX;
    const [selectedIndex, setSelectedIndex] = useState(
        initialIndex >= 0 ? initialIndex : DEFAULT_SELECTED_INDEX
    );
    useEffect(() => {
        const targetIndex = defaultMode
            ? options.findIndex(o => o.mode === defaultMode)
            : DEFAULT_SELECTED_INDEX;
        setSelectedIndex(targetIndex >= 0 ? targetIndex : DEFAULT_SELECTED_INDEX);
    }, [defaultMode]);
    const anchorRef = useRef<HTMLDivElement>(null);

    const selectedOption = options[selectedIndex];
    const isDisabled = !isDirty || isSaving;

    const handlePrimaryClick = () => {
        onSave(selectedOption.mode);
    };

    const handleToggle = () => {
        setOpen((prev) => !prev);
    };

    const handleClose = (event: Event) => {
        if (
            anchorRef.current &&
            anchorRef.current.contains(event.target as HTMLElement)
        ) {
            return;
        }
        setOpen(false);
    };

    const handleMenuItemClick = (index: number) => {
        setSelectedIndex(index);
        setOpen(false);
    };

    return (
        <>
            <ButtonGroup
                variant="contained"
                color="primary"
                ref={anchorRef}
                aria-label="save options"
            >
                <Button
                    onClick={handlePrimaryClick}
                    disabled={isDisabled}
                    startIcon={
                        isSaving ? (
                            <CircularProgress size={20} color="inherit" />
                        ) : null
                    }
                >
                    {selectedOption.label}
                </Button>
                <Button
                    size="small"
                    onClick={handleToggle}
                    disabled={isSaving}
                    aria-controls={open ? 'save-split-menu' : undefined}
                    aria-expanded={open ? 'true' : undefined}
                    aria-label="select save option"
                    aria-haspopup="menu"
                >
                    <ArrowDropDownIcon />
                </Button>
            </ButtonGroup>
            <Popper
                open={open}
                anchorEl={anchorRef.current}
                placement="bottom-end"
                transition
                sx={{ zIndex: 1 }}
            >
                {({ TransitionProps }) => (
                    <Grow {...TransitionProps}>
                        <Paper>
                            <ClickAwayListener onClickAway={handleClose}>
                                <MenuList
                                    id="save-split-menu"
                                    autoFocusItem
                                >
                                    {options.map((option, index) => (
                                        <MenuItem
                                            key={option.mode}
                                            selected={
                                                index === selectedIndex
                                            }
                                            onClick={() =>
                                                handleMenuItemClick(index)
                                            }
                                        >
                                            {option.label}
                                        </MenuItem>
                                    ))}
                                </MenuList>
                            </ClickAwayListener>
                        </Paper>
                    </Grow>
                )}
            </Popper>
        </>
    );
}
