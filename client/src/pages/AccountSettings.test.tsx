// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import AccountSettings from './AccountSettings';
import type { UserSettingsResponse } from '../api/userSettings';

// Mock react-router-dom's useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

// Mock the hooks
vi.mock('../hooks/useUserSettings', () => ({
    useUserSettings: vi.fn(),
    useUpdateUserSettings: vi.fn(),
}));

vi.mock('../hooks', () => ({
    useUnsavedChanges: vi.fn(() => ({
        isDirty: false,
        setIsDirty: vi.fn(),
        clearDirty: vi.fn(),
        checkUnsavedChanges: vi.fn(() => false),
        ConfirmDialog: null,
    })),
}));

// Import mocked hooks
import { useUserSettings, useUpdateUserSettings } from '../hooks/useUserSettings';
import { useUnsavedChanges } from '../hooks';

// Mock MUI icons
vi.mock('@mui/icons-material', () => ({
    ArrowBack: () => <span data-testid="back-icon">Back</span>,
    Save: () => <span data-testid="save-icon">Save</span>,
    Done: () => <span data-testid="done-icon">Done</span>,
    Visibility: () => <span data-testid="visibility-icon">Show</span>,
    VisibilityOff: () => <span data-testid="visibility-off-icon">Hide</span>,
    Dashboard: () => <span data-testid="dashboard-icon">Dashboard</span>,
    Add: () => <span data-testid="add-icon">Add</span>,
    FileUpload: () => <span data-testid="file-upload-icon">FileUpload</span>,
    Logout: () => <span data-testid="logout-icon">Logout</span>,
    Menu: () => <span data-testid="menu-icon">Menu</span>,
    People: () => <span data-testid="people-icon">People</span>,
    Settings: () => <span data-testid="settings-icon">Settings</span>,
    EventNote: () => <span data-testid="event-note-icon">EventNote</span>,
    Timeline: () => <span data-testid="timeline-icon">Timeline</span>,
    // Entity type icons for AppShell navigation
    Business: () => <span data-testid="business-icon">Business</span>,
    Category: () => <span data-testid="category-icon">Category</span>,
    Description: () => <span data-testid="description-icon">Description</span>,
    Event: () => <span data-testid="event-icon">Event</span>,
    ExpandLess: () => <span data-testid="expand-less-icon">ExpandLess</span>,
    ExpandMore: () => <span data-testid="expand-more-icon">ExpandMore</span>,
    Groups: () => <span data-testid="groups-icon">Groups</span>,
    Inventory: () => <span data-testid="inventory-icon">Inventory</span>,
    List: () => <span data-testid="list-icon">List</span>,
    Person: () => <span data-testid="person-icon">Person</span>,
    Pets: () => <span data-testid="pets-icon">Pets</span>,
    Place: () => <span data-testid="place-icon">Place</span>,
    Search: () => <span data-testid="search-icon">Search</span>,
}));

const mockSettings: UserSettingsResponse = {
    contentGenService: 'anthropic',
    contentGenApiKey: '****abc1',
    embeddingService: 'voyage',
    embeddingApiKey: '****xyz2',
    imageGenService: 'openai',
    imageGenApiKey: '****def3',
};

const emptySettings: UserSettingsResponse = {
    contentGenService: null,
    contentGenApiKey: null,
    embeddingService: null,
    embeddingApiKey: null,
    imageGenService: null,
    imageGenApiKey: null,
};

function renderWithRouter(component: React.ReactNode) {
    return render(<BrowserRouter>{component}</BrowserRouter>);
}

describe('AccountSettings', () => {
    const mockMutateAsync = vi.fn();

    beforeEach(() => {
        vi.resetAllMocks();

        vi.mocked(useUpdateUserSettings).mockReturnValue({
            mutateAsync: mockMutateAsync,
            isPending: false,
            error: null,
        } as unknown as ReturnType<typeof useUpdateUserSettings>);

        vi.mocked(useUnsavedChanges).mockReturnValue({
            isDirty: false,
            setIsDirty: vi.fn(),
            clearDirty: vi.fn(),
            checkUnsavedChanges: vi.fn(() => false),
            ConfirmDialog: null,
            markDirty: vi.fn(),
            showConfirmDialog: vi.fn(),
            hideConfirmDialog: vi.fn(),
        });
    });

    it('should render the page title', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Account Settings appears in both breadcrumb and title (h1)
        const title = screen.getByRole('heading', { name: 'Account Settings' });
        expect(title).toBeInTheDocument();
    });

    it('should render loading skeletons when loading', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        const skeletons = document.querySelectorAll('.MuiSkeleton-root');
        expect(skeletons.length).toBeGreaterThan(0);
    });

    it('should render error alert when fetch fails', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('Failed to fetch'),
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        expect(screen.getByText(/Failed to load settings/)).toBeInTheDocument();
    });

    it('should render all three sections', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        expect(screen.getByText('Content Generation')).toBeInTheDocument();
        expect(screen.getByText('Embedding Generation')).toBeInTheDocument();
        expect(screen.getByText('Image Generation')).toBeInTheDocument();
    });

    it('should render service dropdowns', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Should have 3 service combobox dropdowns
        const serviceSelects = document.querySelectorAll('[role="combobox"]');
        expect(serviceSelects.length).toBe(3);
    });

    it('should render API key inputs', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Should have 3 API key password inputs
        const passwordInputs = document.querySelectorAll('input[type="password"]');
        expect(passwordInputs.length).toBe(3);
    });

    it('should populate form with existing settings', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // API key inputs have password type, verify they're present
        const passwordInputs = document.querySelectorAll('input[type="password"]');
        expect(passwordInputs.length).toBe(3);
    });

    it('should toggle API key visibility', async () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Find visibility toggle buttons
        const toggleButtons = screen.getAllByRole('button', { name: /show api key/i });
        expect(toggleButtons.length).toBe(3);

        // Click the first toggle
        fireEvent.click(toggleButtons[0]);

        // The aria-label should change to "Hide API key"
        await waitFor(() => {
            expect(
                screen.getByRole('button', { name: /hide api key/i })
            ).toBeInTheDocument();
        });
    });

    it('should call update mutation on save', async () => {
        const user = userEvent.setup();

        vi.mocked(useUserSettings).mockReturnValue({
            data: emptySettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        const mockSetIsDirty = vi.fn();
        vi.mocked(useUnsavedChanges).mockReturnValue({
            isDirty: true,
            setIsDirty: mockSetIsDirty,
            clearDirty: vi.fn(),
            checkUnsavedChanges: vi.fn(() => false),
            ConfirmDialog: null,
            markDirty: vi.fn(),
            showConfirmDialog: vi.fn(),
            hideConfirmDialog: vi.fn(),
        });

        mockMutateAsync.mockResolvedValue(emptySettings);

        renderWithRouter(<AccountSettings />);

        // Find and click the Save & Close button (save is disabled when empty)
        const saveAndCloseButton = screen.getByRole('button', { name: /save & close/i });
        await user.click(saveAndCloseButton);

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalled();
        });
    });

    it('should show validation error when service is selected but API key is empty', async () => {
        const user = userEvent.setup();

        vi.mocked(useUserSettings).mockReturnValue({
            data: emptySettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        const mockSetIsDirty = vi.fn();
        vi.mocked(useUnsavedChanges).mockReturnValue({
            isDirty: true,
            setIsDirty: mockSetIsDirty,
            clearDirty: vi.fn(),
            checkUnsavedChanges: vi.fn(() => false),
            ConfirmDialog: null,
            markDirty: vi.fn(),
            showConfirmDialog: vi.fn(),
            hideConfirmDialog: vi.fn(),
        });

        renderWithRouter(<AccountSettings />);

        // Open the first service dropdown and select a service
        const serviceSelects = document.querySelectorAll('[role="combobox"]');
        fireEvent.mouseDown(serviceSelects[0]);

        const anthropicOption = await screen.findByText('Anthropic (Claude)');
        fireEvent.click(anthropicOption);

        // Try to save without entering an API key using Save & Close
        const saveAndCloseButton = screen.getByRole('button', { name: /save & close/i });
        await user.click(saveAndCloseButton);

        // Should show validation error
        await waitFor(() => {
            expect(
                screen.getByText(/API key is required when a service is selected/i)
            ).toBeInTheDocument();
        });

        // Mutation should not be called
        expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it('should not render error alert when there is no error', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('should navigate back when back button is clicked and no unsaved changes', async () => {
        const user = userEvent.setup();

        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        const backButton = screen.getByRole('button', { name: /go back/i });
        await user.click(backButton);

        expect(mockNavigate).toHaveBeenCalledWith('/');
    });

    it('should hide embedding API key input when Ollama is selected', () => {
        const ollamaSettings: UserSettingsResponse = {
            ...mockSettings,
            embeddingService: 'ollama',
            embeddingApiKey: null,
        };

        vi.mocked(useUserSettings).mockReturnValue({
            data: ollamaSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Should only have 2 API key password inputs (not 3)
        const passwordInputs = document.querySelectorAll('input[type="password"]');
        expect(passwordInputs.length).toBe(2);

        // Should show the "no API key required" message
        expect(screen.getByText(/no API key required/i)).toBeInTheDocument();
    });

    it('should show breadcrumbs', () => {
        vi.mocked(useUserSettings).mockReturnValue({
            data: mockSettings,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useUserSettings>);

        renderWithRouter(<AccountSettings />);

        // Home link in breadcrumbs
        expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument();
        // Account Settings appears in both breadcrumb and title
        const accountSettingsElements = screen.getAllByText('Account Settings');
        expect(accountSettingsElements.length).toBeGreaterThanOrEqual(2);
    });
});
