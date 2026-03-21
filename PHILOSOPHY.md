# PHILOSOPHY.md

**Status:** Work in progress

_Last updated: 2026-03-21_

This document defines the guiding principles for OpenClaudio. Every feature decision, architectural choice, and contribution must align with these values. When in doubt, refer back here.

---

## Part I: Philosophy

### 1. Small and Extensible Core

OpenClaudio is built around a minimal and functional core based on solving a very specific purpose: Support on the management activities of a single local openclaw instance, mostly focus around its workspace.

**What this means:**
- Core features must be essential to the primary use case: managing a single local openclaw instance
- Experimental features, integrations, and UI enhancements belong in extensions
- Extensions can fail, be removed, or become unmaintained without compromising the core

**When considering a new feature, ask:**
- Does this belong in the core, or could it be an extension?
- Would removing this feature break the fundamental purpose of the app?
- Does this increase the security surface we must defend?

---

### 2. Single user - Local network only

The whole design and threat model borrows heavily from the original openclaw philosophy. We assume this tools will only be managed by a single user with full privileges and that it will only be accessed from your local area network and ideally from only from the same machine where the openclaw instance is running. 

**What this means:**


**Security decisions prioritize:**


---

### 3. Testing Pyramid

Testing follows the classic pyramid: many unit tests, some integration tests, a couple of end-to-end tests.

**Rationale:**
- Unit tests are fast, isolated, and catch regressions early
- Integration tests verify component boundaries
- E2E tests confirm critical user flows work end-to-end, not everything and every little detail

**Guidelines:**


---

### 4. OpenClaw compatible

This is a tool purposely built to manage a openclaw instance. It doesnt aspire to be a generic tool for managing other claw-like tools or to do much many things on top. By design, this tool is tightly coupled with openclaw

**Guidelines:**
- Our latest stable version should support the latest stable openclaw version
- We should never implement breaking changes that diverge from openclaw, we should only have a superset of features.
- If any features we have is implemented natively by openclaw, we should sunset them as quickly as we can, and tell users to switch to the native alternative if possible. If the feature is big enough and it is technically feasable we can consider to port it as an extension so users can choose to keep using it.
- If other tools similar to openclaw (EG: nanoclaw) adhere to the same standards as openclaw, this tool may work for them, but we wont provide explicit support to any other claw-like alternatives unless there are really good reasons for it (EG: OpenClaw changing licences or something cathastrophic as that)
---

### 5. Focused Scope

OpenClaudio serves one purpose exceptionally well: **manage a single openclaw instance efficiently**. We don't try to accommodate every possible use case.

**What we do:**
- Manage openclaw configuration, markdown files, cronjobs, logs, etc
- Extensive markdown-friendly features for enhanced user experience (color coding, diffs, naming, etc)
- Basic openclaw monitoring
- Basic openclaw management functions like update/start/restart/stop
- Support/extend backup functionality from openclaw to enhance the user experience
- Web-based portal

**What we don't do:**
- Native mobile apps
- Social and public features (EG: public dashboard to showoff your openclaw instance)
- Agentic features such as coding on top of openclaw, automatic management of openclaw, etc


**When evaluating feature requests:**
- Does this serve the core use case?
- Would this expand the threat model?
- Could this be an extension instead?
- Are we the right tool for this, or should the user find a specialized app?

**"Your tool doesn't do X, I'll use Y instead."**
→ Then go use Y. We'd rather excel at this than be mediocre at everything. We are open to suggestions as long as they are keep focused

---

### 6. Simple is Good

Simplicity is a feature, not a limitation. Every line of code is a maintenance burden and a potential attack vector.

**Prefer:**
- Direct solutions over abstraction layers
- Explicit code over clever shortcuts
- Fewer dependencies over feature-rich frameworks
- Clear naming over terse abbreviations
- Flat structures over deep hierarchies
- Native solutions instead of home brewed things. EG: Do not come up with your own scheduling, when native cronjobs exist and can be used.

**Avoid:**
- Magic configuration files with dozens of options
- Bloated solutions
- Microservices when a single binary works
- Premature optimization

**When complexity is justified:**
- Security: some protections require careful, non-obvious handling
- Cross-platform support: OS differences sometimes require platform-specific code paths
- Core extensibility: the mechanism that enables extensions adds overhead, but it is the deliberate implementation of Principle 1
- Markdown manipulation: in order to provide the best user experience, we may need to do some "blackmagic" implementation behind the scenes to manipulate the markdown files in a smart way, as they provide little to no functionality out of the box and any logic needs to be built on top of it. There is no way around it. 

---

## Decision Framework

When proposing or reviewing changes, validate against all six principles:

1. **Core vs. Extension?** Does this belong in the core, or is it better as an extension?
2. **Security Impact?** Does this introduce new assumptions or expand the attack surface?
3. **Test Coverage?** Can we write fast, deterministic tests for this?
4. **Data Portability?** Does this affect import/export or create lock-in?
5. **Scope Creep?** Does this align with managing a openclaw instance, or are we building a different app?
6. **Simplicity Cost?** Does this add complexity that outweighs the benefit?

If any principle is violated without strong justification, the proposal should be reconsidered.

---

## Non-Negotiables

Some principles are absolute:

- **No public network access.** This portal is NOT designed to be exposed publicly, if you do it, that's on your own and we are not resposible for any security issues or damages that could be caused by the improper usage or setup of it. The threat model is clear and well defined and we operate under those assumptions.
- **No custom cryptography.** Use standard algorithms and established libraries only.
- **No vendor lock-in.** Users must be able to export and migrate their data freely at any time.
- **Honest threat documentation.** Security claims must be accurate and scoped. Document what IS protected (data at rest, no network leakage) and what is NOT (a compromised OS, physical access while unlocked, coercion). Never overstate the security model.

---

## Closing Thoughts

OpenClaudio is a purposely small, fast and actionable tool. The main purpose is to make the administration of a openclaw instance easier and more handsoff. Not to do fancy stuff. It is a simple boring management tool.

---
---

## Part II: Implementation Guide

This section explains how each principle from Part I translates into concrete decisions in the codebase. It is the "how" to Part I's "what and why". Keep this section updated as the architecture evolves.

---

TBD