# Imagineer Task Tracker

## In Progress
- [ ] Implement API endpoints (connect React frontend to Go backend)
- [ ] First agent: consistency-checker

## Backlog
- [ ] MCP custom tool definitions for Imagineer operations
- [ ] Entity CRUD operations in API
- [ ] Campaign CRUD operations in API
- [ ] Session management features
- [ ] Relationship mapping visualization
- [ ] Timeline view implementation
- [ ] Canon conflict detection and resolution UI
- [ ] User authentication system
- [ ] Import API endpoints for Evernote/Google Docs

## Completed
- [x] Initial project setup
- [x] Core data model implementation (SCHEMAS.md)
- [x] Database migrations (001_initial_schema, 002_seed_game_systems)
- [x] Game system schemas (CoC 7e, GURPS 4e, FitD)
- [x] Docker Compose configuration (PostgreSQL + MCP Server)
- [x] Migration runner script
- [x] Backup/restore scripts
- [x] Claude sub-agent setup (CLAUDE.md, .claude/agents/)
- [x] GitHub Actions CI workflows
- [x] React/Vite client scaffolding
- [x] Client pages (Dashboard, Campaigns, Entities, Timeline, Import)
- [x] Evernote importer (internal/importers/evernote)
- [x] Google Docs importer (internal/importers/googledocs)
- [x] Makefile with test-all, lint, coverage commands
