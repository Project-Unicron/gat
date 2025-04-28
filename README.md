# 😸 GitHub Account Tool (gat)

A simple CLI tool for managing multiple Git identities across different platforms with ease.

## 🎯 Purpose & Philosophy

**gat** lets you seamlessly switch between multiple Git identities—each with its own username, email, token, and optional SSH identity. Perfect for developers who work across multiple accounts on GitHub, GitLab, Bitbucket, Azure DevOps, Hugging Face, and other Git platforms.

### Core Principles

1. **Documentation is Temporal Infrastructure**  
   Every file and function is a message in a bottle. It preserves the "what," the "why," and the "how" of decisions.

2. **Code is Poetry and Self-Expression**  
   Implementation reflects elegance, clarity, and rhythm. Style matters.

3. **Maintainability Above All**  
   DRY, modular, and idiomatic Go. Favor composition. Respect the reader.

4. **Emojis are UI**  
   They're not garnish. They're semantic icons—contextual glyphs for humans.

## 🔥 Features

- 📋 List all your Git profiles
- 🔄 Switch between profiles with a single command
- 🔐 Support for both HTTPS and SSH authentication
- 🔑 Profile-specific SSH identities with automatic configuration
- 🌐 Support for multiple Git platforms (GitHub, GitLab, Bitbucket, Azure DevOps, Hugging Face, and more)
- 🧩 Custom platform definitions for self-hosted instances
- 🩺 Diagnose your Git identity setup
- 🚀 Easily add and remove profiles
- 🧪 Dry-run support for testing changes

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-username/gat.git
cd gat

# Build and install
go mod tidy
go build -o gat cmd/gat/main.go
go install ./cmd/gat  # Important: both build and install steps are needed

# Or just move to a directory in your PATH
mv gat /usr/local/bin/  # Linux/macOS
# or
move gat.exe C:\Windows\System32\  # Windows (run as administrator)
```

### Windows

```powershell
# Using the prebuilt binary
# Download the latest release from the Releases page
# Add to your PATH or place in a directory that's already in your PATH
```

### macOS

```bash
# Using the prebuilt binary
# Download the latest release from the Releases page
chmod +x gat
mv gat /usr/local/bin/
```

### Linux

```bash
# Using the prebuilt binary
# Download the latest release from the Releases page
chmod +x gat
sudo mv gat /usr/local/bin/
```

## 🧰 Usage

### Adding a new profile

```bash
# Add a GitHub profile
gat add work --username "workuser" --email "work@example.com" --token "ghp_token123" --ssh-identity "~/.ssh/id_rsa_work"

# Add a GitLab profile
gat add gitlab-work --username "glworkuser" --email "work@example.com" --platform "gitlab" --token "glpat_token123" --ssh-identity "~/.ssh/id_rsa_gitlab" 

# Add a Bitbucket profile 
gat add bb-work --username "bbworkuser" --email "work@example.com" --platform "bitbucket" --token "bbtoken123" --ssh-identity "~/.ssh/id_rsa_bitbucket"

# Add an Azure DevOps profile
gat add azdo-work --username "azdouser" --email "work@example.com" --platform "azuredevops" --token "azdo_token123" --ssh-identity "~/.ssh/id_rsa_azdo"

# Add a Hugging Face profile
gat add hf-work --username "hfuser" --email "work@example.com" --platform "huggingface" --token "hf_token123" --ssh-identity "~/.ssh/id_rsa_hf"

# Add a profile for a self-hosted GitLab instance
gat add company-gitlab --username "company" --email "me@company.com" --platform "gitlab" --host "git.company.com" --token "glpat_token123" --ssh-identity "~/.ssh/id_rsa_company"
```

### Switching to a profile

```bash
# Switch and keep current protocol
gat switch personal

# Force SSH protocol
gat switch work --ssh

# Force HTTPS protocol
gat switch personal --https

# Dry run (simulate without making changes)
gat switch work --dry-run
```

### Listing all profiles

```bash
gat list
```

### Checking current status

```bash
gat status
```

### Removing a profile

```bash
gat remove outdated
```

### Diagnosing issues

```bash
gat doctor
```

### Starting the API server

```bash
# Start the server on default host (localhost) and port (9999)
gat serve

# Start the server on a specific host and port
gat serve --host 0.0.0.0 --port 8080
```

The API server exposes GAT functionality via REST and GraphQL endpoints:
- **REST:** `http://<host>:<port>/profiles`, `/platforms`, `/doctor`
- **GraphQL:** `http://<host>:<port>/graphql`
- **GraphQL Playground:** `http://<host>:<port>/playground`

### Managing platforms

```bash
# List all supported platforms
gat platforms list

# Register a custom platform using flags
gat platforms register --id gitea --name "Gitea" --host "git.example.com" \
  --ssh-prefix "git@git.example.com:" --https-prefix "https://git.example.com/"

# Register a custom platform using a YAML file
gat platforms register --yaml ~/my-platform.yaml
```

## 🔐 SSH Configuration

`gat` automatically manages SSH configurations for your profiles:

1. Creates `~/.ssh/gat_config` with profile-specific Git hosts
2. Adds an include directive to your main `~/.ssh/config` file
3. Configures per-profile SSH identities

This allows you to use different SSH keys for different accounts without conflicts:

```bash
# Clone a GitHub repository using profile-specific SSH
git clone git@github-work:username/repo.git

# Clone a GitLab repository using profile-specific SSH
git clone git@gitlab-work:username/repo.git

# Clone a Bitbucket repository using profile-specific SSH
git clone git@bitbucket-work:username/repo.git

# Clone an Azure DevOps repository using profile-specific SSH
git clone git@azuredevops-work:username/repo.git

# Clone a Hugging Face repository using profile-specific SSH
git clone git@huggingface-work:username/repo.git

# Clone from a self-hosted GitLab instance
git clone git@gitlab-company:username/repo.git

# Or let gat update your remote URL automatically
cd my-repo
gat switch work --ssh
```

## ⚙️ Configuration

### Profile Configuration

Profiles are stored in `~/.gat/creds.json` with the following structure:

```json
{
  "current": "work",
  "profiles": {
    "work": {
      "username": "nodeops",
      "email": "lynn@workplace.ai",
      "token": "ghp_xyz...",
      "ssh_identity": "~/.ssh/id_rsa_work",
      "platform": "github",
      "host": ""
    },
    "personal": {
      "username": "lynnc",
      "email": "lynn@somewhere.com",
      "token": "ghp_abc...",
      "ssh_identity": "~/.ssh/id_rsa_personal",
      "platform": "github",
      "host": ""
    },
    "azdo-work": {
      "username": "azdouser",
      "email": "work@example.com",
      "token": "azdo_token123",
      "ssh_identity": "~/.ssh/id_rsa_azdo",
      "platform": "azuredevops",
      "host": ""
    },
    "hf-work": {
      "username": "hfuser",
      "email": "work@example.com",
      "token": "hf_token123",
      "ssh_identity": "~/.ssh/id_rsa_hf",
      "platform": "huggingface",
      "host": ""
    },
    "company-gitlab": {
      "username": "gitlab-user",
      "email": "me@company.com",
      "token": "glpat_abc...",
      "ssh_identity": "~/.ssh/id_rsa_company",
      "platform": "gitlab",
      "host": "git.company.com"
    }
  }
}
```

**Note:** If the `creds.json` file contains profiles with missing or invalid fields (e.g., incorrect email format, invalid auth method), `gat` will attempt to load all *valid* profiles and report warnings for the invalid ones. This allows you to continue using your valid profiles even if some configurations are broken.

### Platform Configuration

You can define custom platforms in `~/.gat/platforms.yaml`:

```yaml
# Custom platform example
gitea:
  name: "Gitea"
  defaultHost: "git.example.com"
  sshPrefix: "git@git.example.com:"
  httpsPrefix: "https://git.example.com/"
  sshUser: "git"
  tokenAuthScope: "git.example.com"

# Self-hosted GitHub Enterprise example
github-enterprise:
  name: "GitHub Enterprise"
  defaultHost: "github.mycompany.com"
  sshPrefix: "git@github.mycompany.com:"
  httpsPrefix: "https://github.mycompany.com/"
  sshUser: "git"
  tokenAuthScope: "github.mycompany.com"
```

## 🔧 Troubleshooting

### Common Issues

**Error: profile does not exist**
- Make sure you spelled the profile name correctly
- Check existing profiles with `gat list`

**Error: invalid GitHub username format**
- GitHub usernames must start and end with alphanumeric characters
- Hyphens are allowed in the middle, but not consecutively
- Username length must be between 1-39 characters
- Example valid formats: `user123`, `user-name`, `u`

**SSH authentication fails after switching**
- Verify your SSH key exists at the specified path
- Ensure SSH agent is running with `ssh-add -l`
- Run `gat doctor` to diagnose potential issues

**Git credentials not updating**
- Make sure your token has the necessary permissions
- Check if credential helper is set correctly with `git config --global credential.helper`

**Permission denied errors on Linux/macOS**
- Ensure config files have proper permissions with `chmod 600 ~/.gat/creds.json`

**SSH config includes problems**
- If SSH configuration is not working, try running `gat doctor` to diagnose
- Manually verify that `~/.ssh/config` includes the line `Include ~/.ssh/gat_config`

## 🤝 Contributing

Yes, you can emoji your PR! In fact, it's encouraged.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m '✨ Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Make sure to maintain the emoji-driven UI paradigm and follow the existing code style.

## 📄 License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0) - see the [LICENSE](LICENSE) file for details.

This ensures `gat` remains free and open source forever. Any derivative work must also remain open and under the same license. 