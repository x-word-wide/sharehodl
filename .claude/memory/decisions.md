# Architecture Decision Records

> This file tracks important architecture decisions. Add new decisions at the top.

## ADR Template

```markdown
## ADR-XXX: Title

**Date**: YYYY-MM-DD
**Status**: Proposed | Accepted | Deprecated | Superseded
**Deciders**: [agent names]

### Context
What is the issue that we're seeing that is motivating this decision?

### Decision
What is the change that we're proposing?

### Consequences
What becomes easier or more difficult because of this change?

### Alternatives Considered
What other options were evaluated?
```

---

## ADR-001: Agent Team Architecture

**Date**: 2026-01-28
**Status**: Accepted
**Deciders**: User, Claude

### Context
ShareHODL is a complex blockchain project requiring expertise across multiple domains: system architecture, security, testing, performance, documentation, and implementation. A single-agent approach would lack specialization and could miss domain-specific issues.

### Decision
Implement a multi-agent team with 6 specialized agents:
1. **architect** (opus) - Design and planning
2. **security-auditor** (opus) - Security review
3. **test-engineer** (sonnet) - Testing
4. **performance-optimizer** (sonnet) - Performance
5. **documentation-specialist** (haiku) - Documentation
6. **implementation-engineer** (sonnet) - Coding

Each agent has:
- Specific tools and permissions
- Domain-specific instructions
- Appropriate model (opus for complex reasoning, sonnet for coding, haiku for docs)

### Consequences

**Positive**:
- Specialized expertise for each domain
- Parallel execution capability
- Clear separation of concerns
- Cost optimization (haiku for simple tasks)

**Negative**:
- Coordination overhead
- Potential context fragmentation
- Need for explicit handoffs

### Alternatives Considered
1. **Single agent with modes**: Simpler but less specialized
2. **Two agents (dev/review)**: Less granular but easier coordination
3. **External tools only**: Less intelligent but more predictable

---

## ADR-002: Memory System Design

**Date**: 2026-01-28
**Status**: Accepted
**Deciders**: User, Claude

### Context
Agents need to share context and track progress across sessions. Without persistent memory, each session starts fresh and loses accumulated knowledge.

### Decision
Implement a file-based memory system in `.claude/memory/`:
- `progress.md` - Sprint tracking and active tasks
- `decisions.md` - Architecture decision records
- `learnings.md` - Project-specific learnings
- `issues.md` - Known issues and blockers

Agents update these files as they work, creating a persistent shared state.

### Consequences

**Positive**:
- Persistent context across sessions
- Shared knowledge between agents
- Audit trail of decisions
- Human-readable format

**Negative**:
- Manual updates required
- Potential for stale data
- Git merge conflicts possible

### Alternatives Considered
1. **Database storage**: More structured but harder to review
2. **Comments in code**: Scattered and hard to aggregate
3. **External wiki**: Separate from code, harder to maintain

---

## ADR-003: Rule-Based Code Standards

**Date**: 2026-01-28
**Status**: Accepted
**Deciders**: User, Claude

### Context
Different parts of the codebase have different standards (Go vs TypeScript, security-critical vs general). A single set of rules would be too broad or too restrictive.

### Decision
Implement path-specific rules in `.claude/rules/`:
- `go-cosmos.md` - For `x/**/*.go`, `app/**/*.go`
- `typescript-react.md` - For `sharehodl-frontend/**/*.ts(x)`
- `security.md` - For all code files
- `testing.md` - For test files

Rules use frontmatter `paths` to specify which files they apply to.

### Consequences

**Positive**:
- Targeted guidance per file type
- Layered rules (security applies everywhere)
- Easy to extend for new domains

**Negative**:
- Multiple files to maintain
- Potential for conflicting rules

---

*Add new ADRs above this line*
