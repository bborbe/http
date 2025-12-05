# Product Requirements Documents (PRDs)

This directory contains design documents for features in the http library.

## Status Indicators

- **Draft** 📝 - Planned feature, not yet implemented
- **In Progress** 🚧 - Currently being developed
- **Implemented** ✅ - Complete, PRD serves as historical reference
- **Rejected** ❌ - Decided not to implement

## Purpose

PRDs document the **why**, **what**, and **how** of features:

- **Why** - Problem statement and motivation
- **What** - Goals, non-goals, and success criteria
- **How** - Technical design, implementation plan, and trade-offs

## When to Create a PRD

Create a PRD for features that:
- Add new public APIs or interfaces
- Change existing behavior or architecture
- Require cross-cutting changes (multiple files/components)
- Have multiple implementation approaches to consider
- Need design review or discussion

Don't create PRDs for:
- Bug fixes (use GitHub issues)
- Minor documentation updates
- Internal refactoring without API changes
- Obvious one-line improvements

## PRD Lifecycle

### 1. Draft Phase
- Status: **Draft** 📝
- Author writes PRD with problem statement, goals, technical design
- Team reviews and provides feedback
- PRD updated based on feedback

### 2. Implementation Phase
- Status: **In Progress** 🚧
- Implementation starts following PRD design
- PRD updated if design changes during implementation
- Code references PRD in commit messages/PR descriptions

### 3. Completion Phase
- Status: **Implemented** ✅
- Feature merged and released
- PRD status updated to "Implemented"
- PRD remains as historical reference for design decisions

### 4. Rejection (Optional)
- Status: **Rejected** ❌
- Feature decided against after design phase
- PRD documents why it was rejected
- Prevents re-discussion of rejected ideas

## How to Use PRDs

### As a Contributor
1. Check `docs/prd/` before implementing features
2. Read relevant PRDs to understand design decisions
3. Follow patterns and conventions from implemented PRDs
4. Reference PRD in pull request descriptions

### As a Maintainer
1. Create PRD for significant new features
2. Update status as implementation progresses
3. Archive or clearly mark completed PRDs
4. Use PRDs to review consistency of implementations

### As a User
1. Read Draft PRDs to see planned features and roadmap
2. Review Implemented PRDs to understand design rationale
3. Provide feedback on Draft PRDs via GitHub issues

## PRD Template

See any existing PRD for reference structure. Typical sections:

- **Summary** - One-sentence description
- **Problem Statement** - What problem does this solve?
- **Goals** - What should this achieve?
- **Non-Goals** - What is explicitly out of scope?
- **User Stories** - Who benefits and how?
- **Technical Specification** - How is it implemented?
- **Implementation Plan** - Phased delivery approach
- **Testing Strategy** - How is it validated?
- **Open Questions** - Unresolved decisions

## Current PRDs

| PRD | Status | Created | Implemented |
|-----|--------|---------|-------------|
| [2025-12-05-json-error-handler.md](2025-12-05-json-error-handler.md) | Implemented ✅ | 2025-12-05 | 2025-12-05 |
| [2025-12-05-request-id-middleware.md](2025-12-05-request-id-middleware.md) | Draft 📝 | 2025-12-05 | - |

## Questions?

If you have questions about PRDs or need help creating one, open a GitHub issue or discussion.
