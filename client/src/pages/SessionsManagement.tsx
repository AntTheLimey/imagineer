// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionsManagement - Page for managing chapters and sessions.
 *
 * Provides a two-column layout with chapters on the left and sessions
 * on the right. Accessed via the main left navigation.
 */

import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Alert, Box, Typography } from '@mui/material';
import { ChapterList } from '../components/Chapters';
import { SessionList } from '../components/Sessions';
import type { Chapter, Session } from '../types';

/**
 * Render the Sessions Management page with chapters and sessions panels.
 *
 * @returns The Sessions Management page component
 */
export default function SessionsManagement() {
    const { id } = useParams<{ id: string }>();
    const campaignId = id ? Number(id) : undefined;
    const navigate = useNavigate();

    // Selection state
    const [selectedChapterId, setSelectedChapterId] = useState<number | undefined>();
    const [selectedSessionId, setSelectedSessionId] = useState<number | undefined>();

    if (!campaignId) {
        return (
            <Alert severity="error">
                No campaign selected. Please select a campaign first.
            </Alert>
        );
    }

    const handleCreateChapter = () => {
        navigate(`/campaigns/${campaignId}/chapters/new`);
    };

    const handleEditChapter = (chapter: Chapter) => {
        navigate(`/campaigns/${campaignId}/chapters/${chapter.id}/edit`);
    };

    const handleCreateSession = () => {
        navigate(`/campaigns/${campaignId}/sessions/new`);
    };

    const handleEditSession = (session: Session) => {
        navigate(`/campaigns/${campaignId}/sessions/${session.id}/edit`);
    };

    return (
        <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            {/* Page Header */}
            <Box sx={{ mb: 3 }}>
                <Typography variant="h5" component="h1" gutterBottom>
                    Sessions
                </Typography>
                <Typography variant="body2" color="text.secondary">
                    Organize your campaign into chapters and track individual game sessions.
                </Typography>
            </Box>

            {/* Two-column layout */}
            <Box sx={{ display: 'flex', gap: 3, flex: 1, minHeight: 0 }}>
                {/* Chapters Panel */}
                <Box
                    sx={{
                        width: 320,
                        flexShrink: 0,
                        borderRadius: 1,
                        border: 1,
                        borderColor: 'divider',
                        bgcolor: 'background.paper',
                        overflow: 'hidden',
                        display: 'flex',
                        flexDirection: 'column',
                    }}
                >
                    <ChapterList
                        campaignId={campaignId}
                        selectedChapterId={selectedChapterId}
                        onSelectChapter={setSelectedChapterId}
                        onCreateChapter={handleCreateChapter}
                        onEditChapter={handleEditChapter}
                    />
                </Box>

                {/* Sessions Panel */}
                <Box
                    sx={{
                        flex: 1,
                        minWidth: 0,
                        borderRadius: 1,
                        border: 1,
                        borderColor: 'divider',
                        bgcolor: 'background.paper',
                        overflow: 'hidden',
                        display: 'flex',
                        flexDirection: 'column',
                    }}
                >
                    {selectedChapterId ? (
                        <SessionList
                            campaignId={campaignId}
                            chapterId={selectedChapterId}
                            selectedSessionId={selectedSessionId}
                            onSelectSession={setSelectedSessionId}
                            onCreateSession={handleCreateSession}
                            onEditSession={handleEditSession}
                        />
                    ) : (
                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                height: '100%',
                                p: 3,
                            }}
                        >
                            <Alert severity="info" sx={{ maxWidth: 400 }}>
                                Select a chapter to view its sessions, or create a new
                                chapter to get started.
                            </Alert>
                        </Box>
                    )}
                </Box>
            </Box>

        </Box>
    );
}
