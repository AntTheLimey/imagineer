// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useEffect, useMemo } from 'react';
import {
    Autocomplete,
    Box,
    Chip,
    CircularProgress,
    TextField,
} from '@mui/material';
import { useEntities } from '../../hooks/useEntities';
import type { Entity, EntityType } from '../../types';
import { entityTypeColors, formatEntityType } from './entityConstants';

/**
 * Props for the EntitySelector component.
 */
export interface EntitySelectorProps {
    /** The campaign ID to search entities within. */
    campaignId: number;
    /** The currently selected entity. */
    value: Entity | null;
    /** Callback fired when the selected entity changes. */
    onChange: (entity: Entity | null) => void;
    /** Entity IDs to exclude from the options (e.g., the current entity). */
    excludeIds?: number[];
    /** Optional entity type filter. */
    filterType?: EntityType;
    /** Label displayed above the input. */
    label?: string;
    /** Placeholder text displayed when the input is empty. */
    placeholder?: string;
    /** If true, the input displays in an error state. */
    error?: boolean;
    /** Helper text displayed below the input. */
    helperText?: string;
    /** If true, the input is disabled. */
    disabled?: boolean;
}

/**
 * An autocomplete component for selecting campaign entities.
 *
 * Provides async search functionality with debouncing, displays entity names
 * with type indicators, and supports filtering by entity type. Excluded
 * entity IDs (such as the current entity being edited) are filtered out
 * of the options.
 *
 * @param props - The component props.
 * @returns A React element containing the entity selector autocomplete.
 *
 * @example
 * ```tsx
 * const [selectedEntity, setSelectedEntity] = useState<Entity | null>(null);
 *
 * <EntitySelector
 *     campaignId={campaignId}
 *     value={selectedEntity}
 *     onChange={setSelectedEntity}
 *     excludeIds={[currentEntityId]}
 *     label="Target Entity"
 *     placeholder="Search for an entity..."
 * />
 * ```
 */
export default function EntitySelector({
    campaignId,
    value,
    onChange,
    excludeIds = [],
    filterType,
    label = 'Entity',
    placeholder = 'Search for an entity...',
    error = false,
    helperText,
    disabled = false,
}: EntitySelectorProps) {
    const [inputValue, setInputValue] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');

    // Debounce the search input
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearch(inputValue);
        }, 300);

        return () => clearTimeout(timer);
    }, [inputValue]);

    // Fetch entities based on search term
    const { data, isLoading } = useEntities({
        campaignId,
        searchTerm: debouncedSearch || undefined,
        entityType: filterType,
        pageSize: 20,
    });

    // Filter out excluded entities
    const options = useMemo(() => {
        if (!data) return [];
        return data.filter((entity: Entity) => !excludeIds.includes(entity.id));
    }, [data, excludeIds]);

    return (
        <Autocomplete
            value={value}
            onChange={(_event, newValue) => onChange(newValue)}
            inputValue={inputValue}
            onInputChange={(_event, newInputValue) => setInputValue(newInputValue)}
            options={options}
            getOptionLabel={(option) => option.name}
            isOptionEqualToValue={(option, val) => option.id === val.id}
            loading={isLoading}
            disabled={disabled}
            filterOptions={(x) => x} // Disable client-side filtering; server handles it
            renderInput={(params) => (
                <TextField
                    {...params}
                    label={label}
                    placeholder={placeholder}
                    error={error}
                    helperText={helperText}
                    InputProps={{
                        ...params.InputProps,
                        endAdornment: (
                            <>
                                {isLoading ? (
                                    <CircularProgress color="inherit" size={20} />
                                ) : null}
                                {params.InputProps.endAdornment}
                            </>
                        ),
                    }}
                />
            )}
            renderOption={(props, option) => {
                // Extract key from props to pass separately
                const { key, ...otherProps } = props;
                return (
                    <Box
                        component="li"
                        key={key}
                        {...otherProps}
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                        }}
                    >
                        <Box sx={{ flexGrow: 1 }}>{option.name}</Box>
                        <Chip
                            label={formatEntityType(option.entityType)}
                            size="small"
                            color={entityTypeColors[option.entityType]}
                            variant="outlined"
                        />
                    </Box>
                );
            }}
            renderTags={(tagValue, getTagProps) =>
                tagValue.map((option, index) => {
                    const tagProps = getTagProps({ index });
                    return (
                        <Chip
                            {...tagProps}
                            key={option.id}
                            label={option.name}
                            size="small"
                            color={entityTypeColors[option.entityType]}
                        />
                    );
                })
            }
        />
    );
}
