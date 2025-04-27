# 🌐 GAT v0.2 Project Report
*A CLI tool for cross-platform Git identity routing*

---

## 🧭 Overview
GAT is a smart CLI for managing multiple Git identities across different platforms. It allows seamless switching between GitHub, GitLab, Bitbucket, Hugging Face, Azure DevOps, and self-hosted providers. Profiles include credentials, SSH keys, and host configurations, allowing platform-specific routing.

This document outlines the current project state and reflects recent development milestones, QA efforts, and architectural updates.

---

## 🔄 What’s New in v0.2

- Introduced multi-platform support for GitLab, Bitbucket, Azure DevOps, Hugging Face, and custom/self-hosted platforms
- Added `gat platform register` to define platforms via CLI or YAML
- Automatic SSH host aliasing with per-profile SSH identity routing
- Dry-run support for `gat switch` and other commands
- Profile schema extended with `platform` and `host`
- Platform registry engine with fallback inference logic
- Improved `gat doctor` diagnostics for SSH and HTTPS workflows
- Refactored CLI help output and command descriptions
- Integration test framework (PowerShell-based) with modular test files
- Default platform templates included via `platforms.yaml.example`

---

## 🗂️ Project Structure

```
gat-base/
├── cmd/
│   └── gat/
│       ├── add.go
│       ├── doctor.go
│       ├── list.go
│       ├── main.go
│       ├── platform_register.go
│       ├── platforms.go
│       ├── remove.go
│       ├── root.go
│       ├── status.go
│       └── switch.go
├── pkg/
│   ├── config/config.go
│   ├── git/git.go
│   ├── platform/platform.go
│   ├── ssh/ssh.go
│   └── utils/utils.go
├── tests/
│   ├── run_tests.ps1
│   ├── test_01_basic.ps1
│   ├── test_02_profiles.ps1
│   ├── test_03_platforms.ps1
│   ├── test_04_doctor.ps1
│   └── test_utils.ps1
├── README.md
├── LICENSE
├── go.mod / go.sum
├── hologram.py
├── worm.py
└── platforms.yaml.example
```

---

## 🔧 How to Use This Report
This document serves as an internal snapshot and public-facing milestone summary for v0.2 of the `gat` CLI tool. It can be used for:
- Contributor onboarding
- Technical status reporting
- Release documentation
- Audit trail of architectural changes

---

## ✨ Contributors

- **Lynn Cole** — Project lead, creative architect, systems engineer  
- **Casius** — CLI implementer, unit test machine, YAML validator  
- **Mini** — QA strategist, document scribe, system explainer

---

## 🧩 Next Steps
- Final pass over README structure and installation sections
- Prepare release notes and version tag
- Evaluate whether to begin UX/UI prototype for optional front-end
- Publish `platforms.yaml` schema and examples to GitHub

---

**End of Report**

