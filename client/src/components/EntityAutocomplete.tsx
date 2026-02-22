/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * EntityAutocomplete - Autocomplete component for searching and selecting
 * entities within a campaign. Used by the Reassign action in the Identify
 * phase and reusable in Revise and Enrich phases.
 */

import { useState, useEffect, useMemo } from 'react';
import Autocomplete from '@mui/material/Autocomplete';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { useEntities } from '../hooks/useEntities';
import {
    entityTypeColors,
    formatEntityType,
} from './EntitySelector/entityConstants';
import type { Entity } from '../types';

/**
 * Props for EntityAutocomplete.
 */
export interface EntityAutocompleteProps {
    /** The campaign to search entities within. */
    campaignId: number;
    /** Callback when the user selects an entity. */
    onSelect: (entity: {
        id: number;
        name: string;
        entityType: string;
    }) => void;
    /** Entity ID to exclude from results (e.g., the currently matched entity). */
    excludeEntityId?: number;
    /** Label for the text field. Defaults to "Search entities..." */
    label?: string;
}

/**
 * Debounce delay in milliseconds for the search input.
 */
const DEBOUNCE_MS = 300;

/**
 * Maximum number of entities to fetch per search query.
 */
const PAGE_SIZE = 20;

/**
 * An autocomplete input that searches campaign entities with debounced
 * input and displays results with entity type chips.
 */
export default function EntityAutocomplete({
    campaignId,
    onSelect,
    excludeEntityId,
    label,
}: EntityAutocompleteProps) {
    const [inputValue, setInputValue] = useState('');
    const [debouncedInput, setDebouncedInput] = useState('');

    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedInput(inputValue);
        }, DEBOUNCE_MS);
        return () => clearTimeout(timer);
    }, [inputValue]);

    const { data: entities, isLoading } = useEntities({
        campaignId,
        searchTerm: debouncedInput || undefined,
        pageSize: PAGE_SIZE,
    });

    const filteredOptions = useMemo(() => {
        const list = entities ?? [];
        if (excludeEntityId != null) {
            return list.filter((e) => e.id !== excludeEntityId);
        }
        return list;
    }, [entities, excludeEntityId]);

    return (
        <Autocomplete<Entity, false, false, false>
            options={filteredOptions}
            getOptionLabel={(option) => option.name}
            isOptionEqualToValue={(option, value) => option.id === value.id}
            inputValue={inputValue}
            onInputChange={(_, value) => setInputValue(value)}
            onChange={(_, value) => {
                if (value) {
                    onSelect({
                        id: value.id,
                        name: value.name,
                        entityType: value.entityType,
                    });
                }
            }}
            renderInput={(params) => (
                <TextField
                    {...params}
                    label={label ?? 'Search entities...'}
                    size="small"
                    InputProps={{
                        ...params.InputProps,
                        endAdornment: (
                            <>
                                {isLoading ? (
                                    <CircularProgress
                                        color="inherit"
                                        size={20}
                                    />
                                ) : null}
                                {params.InputProps.endAdornment}
                            </>
                        ),
                    }}
                />
            )}
            renderOption={(props, option) => {
                const { key, ...rest } = props;
                return (
                    <li key={key} {...rest}>
                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 1,
                            }}
                        >
                            <Typography>{option.name}</Typography>
                            <Chip
                                label={formatEntityType(option.entityType)}
                                size="small"
                                color={
                                    entityTypeColors[option.entityType] ??
                                    'default'
                                }
                            />
                        </Box>
                    </li>
                );
            }}
            loading={isLoading}
            noOptionsText={
                inputValue ? 'No entities found' : 'Type to search...'
            }
            filterOptions={(x) => x}
        />
    );
}
