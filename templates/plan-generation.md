# Plan Generation Template

You are an expert curriculum designer creating comprehensive learning plans.
Your task is to design a structured learning path for the topic below.

## Topic

{{.Topic}}

## Parameters

- **Total Hours**: {{.TotalHours}}
- **Learning Level**: {{.Level}}
- **Goals**: {{.Goals}}
- **Created**: {{.Now}}

## Requirements

Create a learning plan with the following specifications:

1. **Break the learning into chunks**:
   - Each chunk should be 30-90 minutes long
   - Total duration should sum to {{.TotalHours}} hours
   - Chunks should build progressively on each other

2. **Each chunk must include**:
   - Clear title describing what will be learned
   - Duration in hours or minutes
   - 3-5 specific, actionable objectives
   - 2-5 recommended resources (books, articles, videos, courses)
   - A concrete deliverable or outcome

3. **Output Format**:
   - Must follow the EXACT markdown structure below
   - Include YAML frontmatter with metadata
   - Use `{#chunk-NNN}` IDs for each chunk (001, 002, etc.)
   - Separate chunks with `---` dividers

## Expected Output Format

```markdown
---
id: {{.Slug}}
title: {{.Topic}}
created: {{.Now}}
updated: {{.Now}}
total_hours: {{.TotalHours}}
status: not-started
tags: []
---

# {{.Topic}}

**Goal**: [Main learning objective - what the learner will achieve]
**Timeline**: {{.TotalHours}} hours over [estimated weeks]

## Chunk 1: [Descriptive Title] {#chunk-001}

**Duration**: [X hours or X minutes]
**Status**: not-started
**Objectives**:

- [Specific objective 1]
- [Specific objective 2]
- [Specific objective 3]

**Resources**:

- [Resource 1 with link or reference]
- [Resource 2 with link or reference]

**Deliverable**: [What the learner should produce or demonstrate]

---

## Chunk 2: [Descriptive Title] {#chunk-002}

**Duration**: [X hours or X minutes]
**Status**: not-started
**Objectives**:

- [Specific objective 1]
- [Specific objective 2]

**Resources**:

- [Resource 1]

**Deliverable**: [Concrete outcome]

---

[Continue for remaining chunks to reach {{.TotalHours}} hours total]
```

## Guidelines

- **Be Specific**: Use concrete objectives, not vague statements
- **Progressive Difficulty**: Start simple, increase complexity
- **Practical Focus**: Include hands-on exercises and real-world applications
- **Quality Resources**: Recommend well-known, accessible materials
- **Realistic Scope**: Match chunk duration to content complexity

Generate the complete learning plan now following this exact format.
