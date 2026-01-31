import { Box, Typography } from '@mui/material';
import { useParams } from 'react-router-dom';

export default function Entities() {
    const { id } = useParams();

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Entities
            </Typography>
            <Typography variant="body1" color="text.secondary">
                Campaign ID: {id}
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
                Entity management coming soon. This page will allow you to:
            </Typography>
            <ul>
                <li>View and manage NPCs, locations, items, and factions</li>
                <li>Define relationships between entities</li>
                <li>Track clues and their discovery paths</li>
                <li>Add keeper notes (GM-only)</li>
            </ul>
        </Box>
    );
}
