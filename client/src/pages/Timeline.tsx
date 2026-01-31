import { Box, Typography } from '@mui/material';
import { useParams } from 'react-router-dom';

export default function Timeline() {
    const { id } = useParams();

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Timeline
            </Typography>
            <Typography variant="body1" color="text.secondary">
                Campaign ID: {id}
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
                Timeline view coming soon. This page will allow you to:
            </Typography>
            <ul>
                <li>View chronological in-game events</li>
                <li>Track what players know vs. hidden events</li>
                <li>Identify timeline inconsistencies</li>
                <li>Plan future events and sessions</li>
            </ul>
        </Box>
    );
}
