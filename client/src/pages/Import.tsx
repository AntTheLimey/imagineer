import { useState } from 'react';
import {
    Box,
    Button,
    Card,
    CardContent,
    Grid,
    Tab,
    Tabs,
    TextField,
    Typography,
} from '@mui/material';
import {
    CloudUpload as UploadIcon,
    Description as EvernoteIcon,
    Article as GoogleDocsIcon,
} from '@mui/icons-material';

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

export default function Import() {
    const [tab, setTab] = useState(0);
    const [evernoteFile, setEvernoteFile] = useState<File | null>(null);
    const [googleDocsUrl, setGoogleDocsUrl] = useState('');

    const handleEvernoteUpload = () => {
        if (evernoteFile) {
            // TODO: Call Evernote import API
            console.log('Importing Evernote file:', evernoteFile.name);
        }
    };

    const handleGoogleDocsImport = () => {
        if (googleDocsUrl) {
            // TODO: Call Google Docs import API
            console.log('Importing from Google Docs:', googleDocsUrl);
        }
    };

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Import Content
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
                Import campaign content from external sources. The importer will
                extract entities, relationships, and timeline events.
            </Typography>

            <Card>
                <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                    <Tabs value={tab} onChange={(_, v) => setTab(v)}>
                        <Tab icon={<EvernoteIcon />} label="Evernote" />
                        <Tab icon={<GoogleDocsIcon />} label="Google Docs" />
                        <Tab icon={<UploadIcon />} label="File Upload" />
                    </Tabs>
                </Box>

                <CardContent>
                    <TabPanel value={tab} index={0}>
                        <Typography variant="h6" gutterBottom>
                            Import from Evernote
                        </Typography>
                        <Typography variant="body2" color="text.secondary" paragraph>
                            Export your notes from Evernote as .enex files and upload them here.
                        </Typography>
                        <Grid container spacing={2} alignItems="center">
                            <Grid item>
                                <Button
                                    variant="outlined"
                                    component="label"
                                    startIcon={<UploadIcon />}
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
                                        <Button variant="contained" onClick={handleEvernoteUpload}>
                                            Import
                                        </Button>
                                    </Grid>
                                </>
                            )}
                        </Grid>
                    </TabPanel>

                    <TabPanel value={tab} index={1}>
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
                                />
                            </Grid>
                            <Grid item>
                                <Button
                                    variant="contained"
                                    onClick={handleGoogleDocsImport}
                                    disabled={!googleDocsUrl}
                                >
                                    Import
                                </Button>
                            </Grid>
                        </Grid>
                    </TabPanel>

                    <TabPanel value={tab} index={2}>
                        <Typography variant="h6" gutterBottom>
                            Upload Files
                        </Typography>
                        <Typography variant="body2" color="text.secondary" paragraph>
                            Upload text files, markdown, or other document formats.
                        </Typography>
                        <Button
                            variant="outlined"
                            component="label"
                            startIcon={<UploadIcon />}
                        >
                            Select Files
                            <input
                                type="file"
                                hidden
                                multiple
                                accept=".txt,.md,.docx,.pdf"
                            />
                        </Button>
                    </TabPanel>
                </CardContent>
            </Card>
        </Box>
    );
}
