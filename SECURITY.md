# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do not** open a public issue
2. Open a [GitHub Security Advisory](https://github.com/calmkart/ai-coding-workflow-bench/security/advisories/new)
3. Include: description, reproduction steps, and potential impact

We will acknowledge receipt within 48 hours and provide a timeline for a fix.

## Scope

workflow-bench executes user-defined commands and AI-generated code in git worktrees. By design, it runs arbitrary code. Users should:

- Run benchmarks in isolated environments (containers, VMs)
- Never run untrusted task definitions
- Review `entry_command` in custom adapter configurations before execution
