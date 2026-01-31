# Security Auditor Knowledge Base

This directory contains security guidelines for the Imagineer project.

## Purpose

This knowledge base provides:

- Security-sensitive code locations
- Attack surface documentation
- Security checklists
- Credential handling guidelines

## Documents

### [security-sensitive-areas.md](security-sensitive-areas.md)

High-risk code locations requiring security focus.

### [attack-surface.md](attack-surface.md)

API endpoints and input validation requirements.

### [credential-handling.md](credential-handling.md)

How to handle credentials and secrets.

### [security-checklist.md](security-checklist.md)

Component-specific security checklists.

## Critical Security Requirements

### GM Notes Protection

GM notes (GM-only content) must NEVER be exposed to players. This is
the most critical security requirement in Imagineer.

### Canon Conflict Handling

Never auto-resolve canon conflicts. Data integrity requires human decision.

## Quick Reference

### OWASP Top 10 Focus

1. Broken Access Control - GM/player separation
2. Injection - SQL parameterization
3. XSS - React escaping, no dangerouslySetInnerHTML

Last Updated: 2026-01-30
