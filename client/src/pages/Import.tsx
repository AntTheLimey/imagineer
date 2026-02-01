// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState } from 'react';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Collapse,
    FormControl,
    Grid,
    InputLabel,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    MenuItem,
    Paper,
    Select,
    Tab,
    Tabs,
    TextField,
    Typography,
} from '@mui/material';
import {
    CheckCircle as SuccessIcon,
    CloudUpload as UploadIcon,
    Description as EvernoteIcon,
    DesktopMac as LocalIcon,
    Article as GoogleDocsIcon,
    Error as ErrorIcon,
    Warning as WarningIcon,
} from '@mui/icons-material';
import {
    useCampaigns,
    useImportEvernote,
    useImportGoogleDocs,
    useImportFiles,
} from '../hooks';
import type { ImportResult } from '../types';
import { EvernoteImport } from '../components/EvernoteImport';

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function TabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;
    return (
        <div
            role="tabpanel"
            hidden={value !== index}
            {...other}
        >
            {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
        </div>
    );
}

/**
 * Component to display import results summary.
 */
function ImportResultSummary({ result, onDismiss }: { result: ImportResult; onDismiss: () => void }) {
    const hasWarnings = result.warnings.length > 0;
    const hasErrors = result.errors.length > 0;

    return (
        <Paper
            sx={{
                p: 2,
                mb: 2,
                bgcolor: result.success ? 'success.light' : 'error.light',
            }}
        >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                {result.success ? (
                    <SuccessIcon color="success" />
                ) : (
                    <ErrorIcon color="error" />
                )}
                <Typography variant="h6">
                    {result.success ? 'Import Successful' : 'Import Failed'}
                </Typography>
            </Box>

            <Typography variant="body1" sx={{ mb: 2 }}>
                Created {result.entitiesCreated} entities and {result.relationshipsCreated} relationships.
            </Typography>

            {hasWarnings && (
                <Box sx={{ mb: 2 }}>
                    <Typography variant="subtitle2" sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
                        <WarningIcon fontSize="small" color="warning" />
                        Warnings ({result.warnings.length})
                    </Typography>
                    <List dense>
                        {result.warnings.map((warning, index) => (
                            <ListItem key={index} sx={{ py: 0 }}>
                                <ListItemIcon sx={{ minWidth: 32 }}>
                                    <WarningIcon fontSize="small" color="warning" />
                                </ListItemIcon>
                                <ListItemText primary={warning} />
                            </ListItem>
                        ))}
                    </List>
                </Box>
            )}

            {hasErrors && (
                <Box sx={{ mb: 2 }}>
                    <Typography variant="subtitle2" sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
                        <ErrorIcon fontSize="small" color="error" />
                        Errors ({result.errors.length})
                    </Typography>
                    <List dense>
                        {result.errors.map((error, index) => (
                            <ListItem key={index} sx={{ py: 0 }}>
                                <ListItemIcon sx={{ minWidth: 32 }}>
                                    <ErrorIcon fontSize="small" color="error" />
                                </ListItemIcon>
                                <ListItemText primary={error} />
                            </ListItem>
                        ))}
                    </List>
                </Box>
            )}

            <Button variant="outlined" size="small" onClick={onDismiss}>
                Dismiss
            </Button>
        </Paper>
    );
}

/**
 * Aggregates multiple import results into a single summary.
 */
function aggregateResults(results: ImportResult[]): ImportResult {
    return {
        success: results.every(r => r.success),
        entitiesCreated: results.reduce((sum, r) => sum + r.entitiesCreated, 0),
        relationshipsCreated: results.reduce((sum, r) => sum + r.relationshipsCreated, 0),
        errors: results.flatMap(r => r.errors),
        warnings: results.flatMap(r => r.warnings),
    };
}

export default function Import() {
    const [tab, setTab] = useState(0);
    const [selectedCampaignId, setSelectedCampaignId] = useState('');
    const [evernoteFile, setEvernoteFile] = useState<File | null>(null);
    const [googleDocsUrl, setGoogleDocsUrl] = useState('');
    const [uploadFiles, setUploadFiles] = useState<File[]>([]);
    const [importResult, setImportResult] = useState<ImportResult | null>(null);

    // Fetch campaigns for the selector
    const { data: campaigns, isLoading: campaignsLoading } = useCampaigns();

    // Import mutations
    const importEvernote = useImportEvernote();
    const importGoogleDocs = useImportGoogleDocs();
    const importFiles = useImportFiles();

    // Check if any import is in progress
    const isImporting = importEvernote.isPending || importGoogleDocs.isPending || importFiles.isPending;

    // Handle Evernote import
    const handleEvernoteUpload = async () => {
        if (!evernoteFile || !selectedCampaignId) return;

        try {
            const result = await importEvernote.mutateAsync({
                campaignId: selectedCampaignId,
                file: evernoteFile,
            });
            setImportResult(result);
            setEvernoteFile(null);
        } catch (error) {
            // Error is handled by the mutation's error state
            console.error('Evernote import failed:', error);
        }
    };

    // Handle Google Docs import
    const handleGoogleDocsImport = async () => {
        if (!googleDocsUrl || !selectedCampaignId) return;

        try {
            const result = await importGoogleDocs.mutateAsync({
                campaignId: selectedCampaignId,
                documentUrl: googleDocsUrl,
            });
            setImportResult(result);
            setGoogleDocsUrl('');
        } catch (error) {
            // Error is handled by the mutation's error state
            console.error('Google Docs import failed:', error);
        }
    };

    // Handle file upload import
    const handleFileUpload = async () => {
        if (uploadFiles.length === 0 || !selectedCampaignId) return;

        try {
            const results = await importFiles.mutateAsync({
                campaignId: selectedCampaignId,
                files: uploadFiles,
            });
            // Aggregate multiple results into one summary
            setImportResult(aggregateResults(results));
            setUploadFiles([]);
        } catch (error) {
            // Error is handled by the mutation's error state
            console.error('File upload import failed:', error);
        }
    };

    // Handle file selection for upload tab
    const handleFilesSelected = (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = event.target.files;
        if (files) {
            setUploadFiles(Array.from(files));
        }
    };

    // Clear the import result
    const dismissResult = () => {
        setImportResult(null);
    };

    // Check if campaign is selected
    const noCampaignSelected = !selectedCampaignId;

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Import Content
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
                Import campaign content from external sources. The importer will
                extract entities, relationships, and timeline events.
            </Typography>

            {/* Campaign Selector */}
            <Paper sx={{ p: 2, mb: 3 }}>
                <FormControl fullWidth disabled={campaignsLoading || isImporting}>
                    <InputLabel id="campaign-select-label">Select Campaign</InputLabel>
                    <Select
                        labelId="campaign-select-label"
                        value={selectedCampaignId}
                        label="Select Campaign"
                        onChange={(e) => setSelectedCampaignId(e.target.value)}
                    >
                        {campaigns?.map((campaign) => (
                            <MenuItem key={campaign.id} value={campaign.id}>
                                {campaign.name}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>
                {noCampaignSelected && (
                    <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                        Please select a campaign before importing content.
                    </Typography>
                )}
            </Paper>

            {/* Import Result Display (for non-Evernote Local tabs) */}
            {tab !== 0 && (
                <Collapse in={!!importResult}>
                    {importResult && (
                        <ImportResultSummary result={importResult} onDismiss={dismissResult} />
                    )}
                </Collapse>
            )}

            {/* Error Alerts */}
            {importEvernote.error && (
                <Alert severity="error" sx={{ mb: 2 }} onClose={() => importEvernote.reset()}>
                    Failed to import Evernote file. Please check the file format and try again.
                </Alert>
            )}
            {importGoogleDocs.error && (
                <Alert severity="error" sx={{ mb: 2 }} onClose={() => importGoogleDocs.reset()}>
                    Failed to import from Google Docs. Please check the URL and ensure the document is accessible.
                </Alert>
            )}
            {importFiles.error && (
                <Alert severity="error" sx={{ mb: 2 }} onClose={() => importFiles.reset()}>
                    Failed to import files. Please check the file formats and try again.
                </Alert>
            )}

            <Card>
                <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                    <Tabs value={tab} onChange={(_, v) => setTab(v)}>
                        <Tab icon={<LocalIcon />} label="Evernote (Local)" />
                        <Tab icon={<EvernoteIcon />} label="Evernote File" />
                        <Tab icon={<GoogleDocsIcon />} label="Google Docs" />
                        <Tab icon={<UploadIcon />} label="File Upload" />
                    </Tabs>
                </Box>

                <CardContent>
                    {/* Evernote Local Tab */}
                    <TabPanel value={tab} index={0}>
                        {noCampaignSelected ? (
                            <Alert severity="info">
                                Please select a campaign above to import from Evernote.
                            </Alert>
                        ) : (
                            <EvernoteImport campaignId={selectedCampaignId} />
                        )}
                    </TabPanel>

                    {/* Evernote File Tab */}
                    <TabPanel value={tab} index={1}>
                        <Typography variant="h6" gutterBottom>
                            Import from Evernote File
                        </Typography>
                        <Typography variant="body2" color="text.secondary" paragraph>
                            Export your notes from Evernote as .enex files and upload them here.
                            This option works on all platforms.
                        </Typography>
                        <Grid container spacing={2} alignItems="center">
                            <Grid item>
                                <Button
                                    variant="outlined"
                                    component="label"
                                    startIcon={<UploadIcon />}
                                    disabled={noCampaignSelected || isImporting}
                                >
                                    Select .enex File
                                    <input
                                        type="file"
                                        hidden
                                        accept=".enex"
                                        onChange={(e) => setEvernoteFile(e.target.files?.[0] || null)}
                                    />
                                </Button>
                            </Grid>
                            {evernoteFile && (
                                <>
                                    <Grid item>
                                        <Typography>{evernoteFile.name}</Typography>
                                    </Grid>
                                    <Grid item>
                                        <Button
                                            variant="contained"
                                            onClick={handleEvernoteUpload}
                                            disabled={noCampaignSelected || importEvernote.isPending}
                                            startIcon={importEvernote.isPending ? <CircularProgress size={20} /> : undefined}
                                        >
                                            {importEvernote.isPending ? 'Importing...' : 'Import'}
                                        </Button>
                                    </Grid>
                                </>
                            )}
                        </Grid>
                    </TabPanel>

                    {/* Google Docs Tab */}
                    <TabPanel value={tab} index={2}>
                        <Typography variant="h6" gutterBottom>
                            Import from Google Docs
                        </Typography>
                        <Typography variant="body2" color="text.secondary" paragraph>
                            Paste a Google Docs URL to import content. The document must be
                            publicly accessible or shared with the application.
                        </Typography>
                        <Grid container spacing={2} alignItems="center">
                            <Grid item xs={12} md={8}>
                                <TextField
                                    fullWidth
                                    label="Google Docs URL"
                                    placeholder="https://docs.google.com/document/d/..."
                                    value={googleDocsUrl}
                                    onChange={(e) => setGoogleDocsUrl(e.target.value)}
                                    disabled={noCampaignSelected || isImporting}
                                />
                            </Grid>
                            <Grid item>
                                <Button
                                    variant="contained"
                                    onClick={handleGoogleDocsImport}
                                    disabled={!googleDocsUrl || noCampaignSelected || importGoogleDocs.isPending}
                                    startIcon={importGoogleDocs.isPending ? <CircularProgress size={20} /> : undefined}
                                >
                                    {importGoogleDocs.isPending ? 'Importing...' : 'Import'}
                                </Button>
                            </Grid>
                        </Grid>
                    </TabPanel>

                    {/* File Upload Tab */}
                    <TabPanel value={tab} index={3}>
                        <Typography variant="h6" gutterBottom>
                            Upload Files
                        </Typography>
                        <Typography variant="body2" color="text.secondary" paragraph>
                            Upload text files, markdown, or other document formats.
                        </Typography>
                        <Grid container spacing={2} alignItems="center">
                            <Grid item>
                                <Button
                                    variant="outlined"
                                    component="label"
                                    startIcon={<UploadIcon />}
                                    disabled={noCampaignSelected || isImporting}
                                >
                                    Select Files
                                    <input
                                        type="file"
                                        hidden
                                        multiple
                                        accept=".txt,.md,.docx,.pdf"
                                        onChange={handleFilesSelected}
                                    />
                                </Button>
                            </Grid>
                            {uploadFiles.length > 0 && (
                                <>
                                    <Grid item>
                                        <Typography>
                                            {uploadFiles.length === 1
                                                ? uploadFiles[0].name
                                                : `${uploadFiles.length} files selected`}
                                        </Typography>
                                    </Grid>
                                    <Grid item>
                                        <Button
                                            variant="contained"
                                            onClick={handleFileUpload}
                                            disabled={noCampaignSelected || importFiles.isPending}
                                            startIcon={importFiles.isPending ? <CircularProgress size={20} /> : undefined}
                                        >
                                            {importFiles.isPending ? 'Importing...' : 'Import'}
                                        </Button>
                                    </Grid>
                                </>
                            )}
                        </Grid>
                        {uploadFiles.length > 1 && (
                            <Box sx={{ mt: 2 }}>
                                <Typography variant="subtitle2" gutterBottom>
                                    Selected files:
                                </Typography>
                                <List dense>
                                    {uploadFiles.map((file, index) => (
                                        <ListItem key={index} sx={{ py: 0 }}>
                                            <ListItemText
                                                primary={file.name}
                                                secondary={`${(file.size / 1024).toFixed(1)} KB`}
                                            />
                                        </ListItem>
                                    ))}
                                </List>
                            </Box>
                        )}
                    </TabPanel>
                </CardContent>
            </Card>
        </Box>
    );
}
