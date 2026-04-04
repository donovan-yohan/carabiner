# Philosophy

Foundational thinking behind carabiner's design. Reference this when brainstorming features, evaluating competing approaches, or explaining carabiner's position in the agentic coding space.

Source: Slate binary analysis session + adversarial review research + belayer v5 direction (2026-03-29/30), join-layer pivot session (2026-04-04)

---

## The Three Roles

Agent, Harness, and Orchestrator are separate concerns. Carabiner is the Harness.

**Agent** = raw model capability. Claude, Codex, Slate, Gemini, whatever comes next. The agent reads code, writes code, runs commands, solves problems. Agent capability is improving rapidly. Neither carabiner nor belayer own this.

**Harness** = the repo's forensic memory. Who wrote what, why, and whether the spec was clear. This is carabiner. Any agent that can run a shell command can query it.

**Orchestrator** = the layer responsible for getting work done end-to-end. Intake (idea to spec), Implementation (agent does the work), Output (review, gates, PR). This is belayer. YAML pipelines, node execution, gate contracts.

The separation test: you can use carabiner without belayer (query attribution history directly). You can use belayer without carabiner (pipelines work, no forensic memory). They compose but don't depend on each other.

---

## The Join Is the Product

The data for AI code attribution already exists in three disconnected systems:

- **git-ai** tracks which agent session wrote which lines of which commit (Git Notes at `refs/notes/ai`). This is the "what."
- **agentlytics** records full session transcripts, tool calls, file accesses, and timestamps. This is the "how."
- **Work-item systems** (Linear, Jira, GitHub) link PRs to issues, specs, and plans. This is the "why."

Nobody joins these. git-ai knows authorship but not intent. agentlytics knows what the agent did but not which work item it was for. Linear knows the issue but not the agent session.

Carabiner is the join layer. It reads all three data sources and connects them into a single forensic query: "this line was written by this agent session, working on this issue, using this spec, with this confidence."

### Why Not Build Our Own Attribution?

Because git-ai already does it well. 1,481 stars, daily commits, Apache-2.0, handles rebases and squashes, expanding agent support. Building a competing attribution system would duplicate mature work and fragment the ecosystem.

### Why Not Build Our Own Collection?

Because agentlytics already does it well. Reads session stores from every major agent, normalizes into SQLite, 30-second scan. The collection problem is solved.

### What We Build Instead

The cross-system query that nobody else provides. The deterministic join key (`git-ai.agent_id.id` = `agentlytics.conversation_id`) turns "AI archaeology" from fuzzy timestamp correlation into a foreign key join. Everything else is enrichment on top of that spine.

---

## Confidence Is Core Product

A bug is rarely traceable to a single clean cause. One session wrote the code, another edited it, a human reviewed it, the spec drifted, and the work-item link might be weak or missing.

If carabiner tells a clean story when the evidence is messy, users stop trusting it. Every hop in the attribution chain carries a confidence label:

- **High**: deterministic join (git-ai hash match, agentlytics conversation_id match)
- **Medium**: heuristic correlation (timestamp + file overlap Jaccard > 0.3)
- **Low**: weak signal (timestamp window only, branch name pattern only)
- **Missing**: no data for this hop

The dossier shows the weakest link. The confidence model isn't UI garnish, it's the core product. Honest ambiguity builds trust. False certainty destroys it.

---

## Graceful Degradation

Carabiner works with whatever data sources are available. git-ai and agentlytics are recommended, not required.

- git-ai installed: commit-to-session join is deterministic (high confidence)
- No git-ai: fall back to timestamp + file Jaccard correlation (medium confidence)
- agentlytics available: full session transcripts and tool call history
- No agentlytics: git-ai notes only, no session context
- Neither: carabiner has nothing to join

This means carabiner is useful the moment you install it in a repo that has git-ai notes, even before agentlytics is set up. And it gets richer as you add more data sources.

---

## Compose, Don't Compete

Carabiner's position in the ecosystem:

| Concern | Tool | Carabiner's relationship |
|---------|------|--------------------------|
| Line-level attribution | git-ai | Reads git-ai's Git Notes. Does not write them. |
| Session data collection | agentlytics | Reads agentlytics' SQLite cache. Does not collect. |
| Token usage / cost | tokscale | Orthogonal. Could read tokscale data in future. |
| Work-item tracking | Linear, Jira, GitHub | Reads PR/issue links. Does not manage work items. |
| Pipeline orchestration | belayer | Independent. Belayer can call carabiner, but doesn't need to. |

This is a deliberate choice. Every tool carabiner composes with is independently maintained, has its own community, and solves its problem well. Carabiner adds value by connecting them, not by replacing them.

---

## Retroactive Over Explicit

Most attribution tools require agents to explicitly participate in the attribution protocol. git-ai requires `git-ai checkpoint`. The prior carabiner design required `carabiner context set`. GitHub Copilot's tracing only works for Copilot.

Carabiner's insight: session-level retroactive attribution from already-existing data is sufficient for the forensic use case ("what was the agent doing when this bug got written?"), and it's the only approach that works without agent cooperation.

When git-ai is installed, attribution is deterministic and requires no extra agent cooperation (git-ai's hooks handle it automatically). When git-ai is absent, carabiner falls back to timestamp + file overlap correlation. Less precise, but never fails silently and never requires the developer to remember to set context.

You trade line-level precision for zero-config reliability. For the forensic debugging use case, session-level attribution ("this session produced code in this commit") is the right granularity.

---

## Two Installation Scopes

**System-level** (`~/.carabiner/` or equivalent): reads agentlytics' session cache, which spans all repos on the machine. This gives carabiner access to session data regardless of which repo the agent was working in.

**Repo-level** (`.carabiner/`): reads git-ai notes and git history for this specific repo. This is where work-item linkage happens, because branch names, commit messages, and PR metadata are repo-scoped.

Both scopes feed the same query interface. `carabiner why` works regardless of which data sources are available, reporting what it found and what's missing.

---

## Why Not a Plugin

The previous harness was a Claude Code plugin. This locked it to one agent ecosystem. An agent plugin only works in the agent it's installed in.

Carabiner is a CLI. `carabiner why src/auth/handler.go:47` works from Claude Code, from Codex, from a shell script, from a CI pipeline, from any agent that can run `sh -c`. The universal interface is the shell command.

---

## H-as-Feature: Why Context Loss Is Sometimes Desirable

In the D > 4H framing:
- D = information loss from a single long context (dumb zone degradation)
- H = information loss per handoff between pipeline nodes

For adversarial review, H is a feature, not a cost.

LLM sycophancy research shows models perform better at review when they DON'T share context with the implementer. Isolated reviewers catch more issues because they lack implementation bias. They evaluate the artifact on its own merits rather than validating the author's reasoning.

This principle applies to carabiner's forensic queries too. The dossier presents evidence, not narrative. It shows which session wrote which lines with what confidence, but doesn't editorialize about whether the agent made a good decision. The human reading the dossier brings their own judgment.

---

## Trust Agent Competency

Carabiner doesn't prescribe how agents work. It doesn't tell the agent to "first plan, then implement, then review." The agent decides its own workflow.

Carabiner provides one thing: forensic context. "Here's what happened when this code was written." What the agent or developer does with that information is their business.

The previous harness plugin had 12+ commands prescribing a rigid workflow (brainstorm, plan, orchestrate, validate, review, reflect, complete, prune). This fought the natural evolution of models getting better at self-planning and self-reviewing. Carabiner is a forensic layer, not a workflow engine.

---

## The False Positive Contamination Problem

Gate failures are not ground truth. When a gate says "this auth middleware is missing a registry update," that might be a real bug, a false positive, or a stylistic preference.

If quality standards auto-evolve from gate failures, and 30% are false positives, the system encodes false positives as institutional knowledge. After 6 months, the quality patterns file contains "always do X" where X is something a model hallucinated.

The real ground truth is bugs traced back to specific runs. Code that made it through the gate, got merged, and then caused a problem. Carabiner's forensic chain (`carabiner why`) is what makes this tracing possible. When a bug is found, you can trace it back to the agent session, the spec, and the work item, and determine whether the failure was in the code, the spec, or the review process.

This external signal (traced bugs connected to sessions via carabiner) is the foundation for any future quality evolution system. The forensic layer must exist and be trusted before the quality layer can safely learn from it.

---

## Relationship to Belayer

Belayer is the orchestrator. It runs YAML pipelines, executes nodes, routes based on gate outcomes. When a gate fails, belayer routes to `on_fail`. Belayer doesn't know WHY it failed and doesn't have forensic context.

With carabiner available, a belayer framework can enrich gate failure reports: "this gate failed on code written by Claude session X, working on ENG-42, where the spec was ambiguous about token storage." This is richer than "gate failed, here's the diff."

A different framework could skip carabiner entirely. A user could use carabiner without any framework. The separation is real.

---

## Concurrent Safety

Carabiner's query layer is read-only against external data sources (git-ai notes, agentlytics cache, git history). Multiple carabiner queries can run simultaneously without conflicts. The index (Approach B, when built) uses SQLite WAL mode for concurrent read access.

Git-ai handles concurrent write safety for attribution notes. Agentlytics handles concurrent write safety for session data. Carabiner inherits their guarantees by reading, not writing.
