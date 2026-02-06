# ShareHODL Agent Team

## Team Overview

This project uses a coordinated AI agent team that works together to handle all aspects of development - from planning to implementation, testing, security, and documentation.

## Agent Roster

| Agent | Role | Model | Primary Tools | When to Use |
|-------|------|-------|---------------|-------------|
| **architect** | System design, planning | opus | Read, Glob, Grep, Task | New features, architecture decisions, cross-module changes |
| **security-auditor** | Security review | opus | Read, Glob, Grep, Bash | Code audits, vulnerability detection, security concerns |
| **test-engineer** | Testing & QA | sonnet | Read, Glob, Grep, Bash, Edit, Write | Writing tests, coverage analysis, bug reproduction |
| **performance-optimizer** | Performance tuning | sonnet | Read, Glob, Grep, Bash | Bottleneck identification, optimization, profiling |
| **documentation-specialist** | Technical writing | haiku | Read, Glob, Grep, Edit, Write | Docs updates, API documentation, guides |
| **implementation-engineer** | Code implementation | sonnet | Read, Glob, Grep, Bash, Edit, Write, Task | Writing code, bug fixes, feature implementation |

## Workflow Patterns

### New Feature Development
```
1. architect → Design & plan the feature
2. security-auditor → Review design for security
3. implementation-engineer → Write the code
4. test-engineer → Write and run tests
5. security-auditor → Security review of implementation
6. performance-optimizer → Performance review (if applicable)
7. documentation-specialist → Update docs
```

### Bug Fix
```
1. test-engineer → Reproduce and write failing test
2. implementation-engineer → Fix the bug
3. test-engineer → Verify fix and add regression test
4. security-auditor → Review if security-related
```

### Security Audit
```
1. security-auditor → Full security review
2. implementation-engineer → Fix issues
3. security-auditor → Verify fixes
4. test-engineer → Add security tests
```

### Performance Optimization
```
1. performance-optimizer → Profile and identify issues
2. implementation-engineer → Implement optimizations
3. test-engineer → Performance tests
4. performance-optimizer → Verify improvements
```

## Agent Invocation

### Automatic
Claude will automatically invoke agents based on task description:
- "Plan a new feature..." → architect
- "Review security..." → security-auditor
- "Write tests..." → test-engineer
- "Optimize..." → performance-optimizer
- "Update docs..." → documentation-specialist
- "Implement..." → implementation-engineer

### Explicit
Request specific agents:
- "Use the architect agent to design..."
- "Have security-auditor review..."
- "Ask test-engineer to write tests for..."

### Parallel
For independent tasks:
- "Run security-auditor and performance-optimizer in parallel"

## Coordination Protocol

### Handoffs
When an agent completes their work, they should:
1. Summarize what was done
2. List files modified
3. Identify next agent in workflow
4. Update `.claude/memory/progress.md`

### Shared Context
All agents have access to:
- `CLAUDE.md` - Project overview
- `.claude/rules/` - Coding standards
- `.claude/memory/` - Shared state

### Communication
Agents communicate through:
1. Memory files (progress.md, decisions.md, issues.md)
2. Todo lists (TodoWrite tool)
3. Inline code comments
4. Handoff summaries

## Quality Gates

### Before Merge
- [ ] architect approved design (for features)
- [ ] implementation-engineer completed code
- [ ] test-engineer verified tests pass
- [ ] security-auditor approved (for sensitive code)
- [ ] documentation-specialist updated docs

### For Security-Critical Code
- [ ] security-auditor full review
- [ ] test-engineer security tests
- [ ] Two-pass review (different sessions)

## Best Practices

### For Users
1. Be specific about what you want
2. Provide context and constraints
3. Let agents finish before interrupting
4. Review agent work before accepting

### For Agents
1. Always read relevant code before modifying
2. Follow existing patterns in codebase
3. Update memory files after significant work
4. Request review from appropriate agents
5. Never skip tests

## Troubleshooting

### Agent Not Working?
- Check agent definition in `.claude/agents/`
- Verify tools are permitted in settings
- Ensure model is available

### Coordination Issues?
- Check `.claude/memory/progress.md` for state
- Review recent changes for context
- Reset memory files if corrupted

### Performance Issues?
- Use haiku model for simple tasks
- Run agents in parallel when possible
- Keep prompts focused and specific
