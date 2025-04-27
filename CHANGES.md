# Multi-Platform Git Identity Routing

This document summarizes the changes made to implement multi-platform Git identity routing in the `gat` CLI tool.

## Overview

Previously, `gat` was primarily focused on GitHub identities. This implementation expands support to multiple Git hosting platforms including GitHub, GitLab, Bitbucket, Hugging Face, Azure DevOps, and custom self-hosted instances.

## Changes Made

1. **Enhanced Profile Schema**
   - Added `platform` field to identify which platform the profile is for
   - Added `host` field to support custom domains and self-hosted instances

2. **Platform Registry**
   - Expanded the platform registry to include major Git hosting platforms:
     - GitHub (`github.com`)
     - GitLab (`gitlab.com`)
     - Bitbucket (`bitbucket.org`)
     - Hugging Face (`huggingface.co`)
     - Azure DevOps (`dev.azure.com`)
   - Support for custom platform definitions via `~/.gat/platforms.yaml`

3. **Protocol Routing Rules**
   - Updated SSH and HTTPS URL generation to be platform-aware
   - Enhanced remote URL conversion to handle various platform URL formats
   - Improved validation to accept URLs from all supported platforms

4. **SSH Configuration Management**
   - Updated SSH config generation to use platform-specific hosts
   - Added support for custom hosts in SSH configuration
   - Ensured unique host alias generation for each platform+profile combination

5. **User Interface Improvements**
   - Updated `list` command to show platform information
   - Enhanced `doctor` command to validate platform configurations
   - Added platform-specific examples to the CLI help

6. **Documentation**
   - Updated README with multi-platform usage examples
   - Created a sample `platforms.yaml.example` file as a template

## Examples

### Adding Profiles for Different Platforms

```bash
# GitHub profile
gat add github-work --username "ghuser" --email "work@example.com" --platform "github"

# GitLab profile
gat add gitlab-work --username "gluser" --email "work@example.com" --platform "gitlab"

# Bitbucket profile
gat add bitbucket-work --username "bbuser" --email "work@example.com" --platform "bitbucket"

# Self-hosted GitLab instance
gat add company-gitlab --username "companyuser" --email "work@example.com" --platform "gitlab" --host "gitlab.company.com"
```

### SSH Host Configuration

For each profile, an SSH host entry is created like:

```ssh
Host github-work
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_github_work
    IdentitiesOnly yes

Host gitlab-work
    HostName gitlab.com
    User git
    IdentityFile ~/.ssh/id_gitlab_work
    IdentitiesOnly yes

Host gitlab-company
    HostName gitlab.company.com
    User git
    IdentityFile ~/.ssh/id_company_gitlab
    IdentitiesOnly yes
```

### Custom Platform Definition

Users can define additional platforms in `~/.gat/platforms.yaml`:

```yaml
gitea:
  name: "Gitea"
  defaultHost: "git.example.com"
  sshPrefix: "git@git.example.com:"
  httpsPrefix: "https://git.example.com/"
  sshUser: "git"
  tokenAuthScope: "git.example.com"
```

## Validation and Safety Features

1. Platform conflict detection in the `doctor` command
2. Secure handling of custom domains
3. Validation of URL formats across all platforms
4. Proper error handling for unknown platforms

## Next Steps

Future enhancements could include:
1. More granular token scopes per platform
2. Additional platform-specific features
3. Support for more advanced authentication methods 