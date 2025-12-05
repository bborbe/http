# Product Requirements Documents (PRDs)

This directory contains design documents for features in the http library.

## Status Indicators

- **Draft** 📝 - Being written internally
- **Proposed** - Ready for review (PR open)
- **In Review** - Actively collecting feedback
- **Approved** - Decision made, not yet implemented
- **In Progress** 🚧 - Implementation started
- **Implemented** ✅ - Code merged, PRD now immutable
- **Superseded** - Replaced by newer PRD
- **Rejected** ❌ - Decided not to implement (keep for history)
- **Withdrawn** - Pulled back by author
- **Deprecated** - Feature removed or obsolete

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
- PRD becomes **immutable** - no further edits

### 4. Rejection (Optional)
- Status: **Rejected** ❌
- Feature decided against after design phase
- PRD documents why it was rejected
- Prevents re-discussion of rejected ideas

## PRD Immutability

**Once Status: Implemented, STOP editing the PRD.**

PRDs become **historical snapshots** of original design decisions. This preserves the "why" behind architectural choices.

### Why Immutability Matters

- ✅ Original reasoning preserved forever
- ✅ Can trace feature evolution through multiple PRDs
- ✅ Future maintainers understand past constraints
- ✅ Prevents re-litigating old decisions

### Where Future Changes Go

**If the feature changes after implementation:**

1. **Update Living Documentation** - README.md, API docs, Wiki
   - Living docs = always current, reflects actual behavior
   - PRD = historical snapshot of original design

2. **Write New PRD** (for major redesigns)
   - Create new PRD with new date: `YYYY-MM-DD-feature-v2.md`
   - Reference original PRD in frontmatter: `superseded_by: "2025-12-05-feature-v2"`
   - Both PRDs preserved = complete evolution history

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

## PRD Structure

### Frontmatter (YAML)

All PRDs include structured metadata at the top:

```yaml
---
id: YYYY-MM-DD-feature-name
title: Feature Name
status: Draft
created: 2025-12-05
authors: ["@username"]
tags: [prd, component-name]
related_issues: []
implemented_pr: ""
superseded_by: ""
---
```

**Benefits:**
- Enables automated indexing and filtering
- Links to implementation (implemented_pr)
- Tracks PRD evolution (superseded_by)
- Searchable by tags and status

### Warning Banner

Warning banners are **only used for PRDs that are no longer relevant**:

**Superseded** (replaced by newer PRD):
```markdown
> ⚠️ SUPERSEDED
> This PRD has been replaced. See [2026-01-15-feature-v2.md](link)
```

**Rejected** (decided not to implement):
```markdown
> ⚠️ REJECTED
> This feature was not implemented. See rejection rationale below.
```

**Deprecated** (feature was removed):
```markdown
> ⚠️ DEPRECATED
> This feature has been removed from the library.
```

**Withdrawn** (pulled back by author):
```markdown
> ⚠️ WITHDRAWN
> This proposal was withdrawn and is no longer being pursued.
```

**No banner needed** for:
- Draft, Proposed, In Review, Approved, In Progress (work in progress)
- **Implemented** (feature exists, PRD documents design)

### Standard Sections

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
