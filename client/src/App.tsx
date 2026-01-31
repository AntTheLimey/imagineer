import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Box } from '@mui/material';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Campaigns from './pages/Campaigns';
import Entities from './pages/Entities';
import Timeline from './pages/Timeline';
import Import from './pages/Import';

function App() {
    return (
        <BrowserRouter>
            <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                <Layout>
                    <Routes>
                        <Route path="/" element={<Dashboard />} />
                        <Route path="/campaigns" element={<Campaigns />} />
                        <Route path="/campaigns/:id/entities" element={<Entities />} />
                        <Route path="/campaigns/:id/timeline" element={<Timeline />} />
                        <Route path="/import" element={<Import />} />
                    </Routes>
                </Layout>
            </Box>
        </BrowserRouter>
    );
}

export default App;
