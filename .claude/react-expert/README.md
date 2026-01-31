# React Expert Knowledge Base

This directory contains React development guidance for Imagineer.

## Documents

- `architecture-overview.md` - Client architecture
- `component-structure.md` - Component organization
- `state-management.md` - React Query patterns
- `mui-patterns.md` - Material-UI usage

## Quick Reference

### Project Structure

```
client/src/
├── main.tsx          - Entry point
├── App.tsx           - Root component, routing
├── components/       - Reusable components
│   └── Layout.tsx    - Main layout
├── pages/            - Page components
│   ├── Dashboard.tsx
│   ├── Campaigns.tsx
│   ├── Entities.tsx
│   ├── Timeline.tsx
│   └── Import.tsx
├── types/            - TypeScript definitions
│   └── index.ts
└── test/             - Test setup
```

### Key Libraries

| Library | Purpose |
|---------|---------|
| React 18 | UI framework |
| TypeScript | Type safety |
| Material-UI | Component library |
| React Router | Navigation |
| React Query | Server state |
| Vite | Build tooling |
| Vitest | Testing |

### Common Commands

```bash
npm run dev      # Start dev server
npm run build    # Production build
npm run test     # Run tests
npm run lint     # Check linting
```

Last Updated: 2026-01-30
