# Origin Story

How carabiner came to exist. Reference for context on the design decisions.

## The Session (2026-03-29/30)

Carabiner emerged from a deep-dive analysis session that started with reverse-engineering Random Labs' Slate agent binary and ended with splitting belayer's harness into its own tool.

### Phase 1: Slate Analysis

We decompiled the Slate binary (`@randomlabs/slate` npm package, 78MB Mach-O arm64) to understand their "thread weaving and episodes" architecture. Key findings:

- Slate has a `passContext` boolean toggle that switches between shared context (thread mode) and isolated context (subagent mode). Despite the blog's rhetoric that threads supersede subagents, the implementation hedges.
- The "DSL" Slate claims to use is actually just an `orchestrate` tool call that spawns child sessions. No custom language.
- The UI still calls everything "subagents." The terminology pivot is marketing.
- Model slots: separate models for main (sonnet-4.6), subagent (sonnet-4.6/codex), search (haiku-4.5/glm-5), reasoning (gpt-5.3-codex).
- No persistent cross-session memory. Their bet: the code IS the memory.

### Phase 2: Adversarial Review Debate

The Slate blog frames context isolation as a cost to minimize. We challenged this with sycophancy research showing isolated reviewers catch more issues because they lack implementation bias.

Steelmanned both sides. Threads win for implementation (shared context helps). Subagents win for review (isolation is the mechanism of quality). The ideal system needs both primitives and deterministic rules about when to use which.

This led to "H-as-feature": handoff context loss between pipeline nodes is architecturally desirable for adversarial review.

### Phase 3: Three Roles Separation

Through the Slate analysis and adversarial debate, three distinct roles emerged:

- **Agent** = raw model capability (Claude, Codex, Slate)
- **Harness** = memory and code context (AGENTS.md, quality patterns, domain knowledge)
- **Orchestrator** = workflow management (Intake → Implementation → Output)

Belayer was identified as the orchestrator. The harness concern was initially part of belayer ("belayer harness quality check") but the pressure testing revealed this violated the separation of concerns. The harness should be its own tool.

### Phase 4: The Learning Loop

The highest-value unsolved problem: when adversarial review gates fail, capture WHY and prevent recurrence. This creates a flywheel no single-agent tool has.

We analyzed 11 PRs across belayer and claude-remote-cli repos. Silent-failure-hunter was the most prolific and actionable reviewer (38 comments, ~90% actionable). Four recurring patterns accounted for ~65% of all findings — cross-repo, cross-language patterns that would benefit from prompt injection.

### Phase 5: Carabiner Split

The final architectural insight: the harness must be a separate CLI, not a belayer subcommand or a Claude Code plugin. Reasons:

1. **Agent-agnostic**: any agent that can `sh -c` can use it
2. **Forces clean separation**: separate repos prevent concern leakage
3. **CLI is the universal interface**: plugins only work in their host agent

Two jobs emerged for carabiner:
- **Quality** (reactive): patterns learned from review failures
- **Knowledge** (proactive): domain semantics agents can't derive from code

The existing harness plugin (12+ commands prescribing rigid workflows) gets deprecated. Agents should own their workflow. Carabiner provides knowledge, not process.

## Key References

- Slate blog: "Slate: moving beyond ReAct and RLM" (randomlabs.ai/blog/slate, 2026-03-09)
- Context Rot paper (Chroma, 2025): LLM attention degrades non-uniformly as context grows
- Anthropic: "Harness design for long-running application development" (GAN-inspired generator/evaluator separation)
- Meta: HyperAgents paper (self-referential agents, persistent memory, open-ended variant archives)
- Semgrep Community Edition: YAML-native rules with path scoping (Codex identified as 50% solution for quality patterns)
- Belayer design docs: 2026-03-28 self-improving harness runtime, 2026-03-29 quality learning loop

## The Full Session Transcript

Available at: belayer-guide repo, `docs/references/2026-03-29-slate-analysis-and-belayer-v5-direction.md`
