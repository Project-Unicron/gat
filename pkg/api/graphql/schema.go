package graphql

// Schema defines the GraphQL schema for the GAT API
const Schema = `
  type Query {
    # Get all profiles
    profiles: [Profile!]!
    
    # Get a specific profile by name
    profile(name: String!): Profile
    
    # Get the current active profile
    currentProfile: Profile
    
    # Get all supported platforms
    platforms: [Platform!]!
    
    # Get a specific platform by ID
    platform(id: String!): Platform
    
    # Run diagnostic checks
    doctor: DiagnosticResult!
  }

  type Mutation {
    # Switch to a different profile
    switchProfile(input: SwitchProfileInput!): SwitchProfileResult!
    
    # Add a new profile
    addProfile(input: AddProfileInput!): AddProfileResult!
    
    # Remove a profile
    removeProfile(name: String!): RemoveProfileResult!
    
    # Register a custom platform
    registerPlatform(input: RegisterPlatformInput!): RegisterPlatformResult!
  }

  # A Git profile with identity information
  type Profile {
    name: String!
    username: String!
    email: String!
    platform: String!
    platformDetails: Platform
    host: String
    hasToken: Boolean!
    sshIdentity: String
    isActive: Boolean!
  }

  # A Git hosting platform definition
  type Platform {
    id: String!
    name: String!
    defaultHost: String!
    sshPrefix: String!
    httpsPrefix: String!
    sshUser: String!
    tokenAuthScope: String!
    isCustom: Boolean!
  }

  # Input for switching profiles
  input SwitchProfileInput {
    name: String!
    protocol: Protocol
    dryRun: Boolean
  }

  # Supported protocols
  enum Protocol {
    SSH
    HTTPS
  }

  # Result of a profile switch operation
  type SwitchProfileResult {
    success: Boolean!
    message: String
    profile: Profile
    gitConfigChanges: [GitConfigChange!]
  }

  # A Git configuration change
  type GitConfigChange {
    key: String!
    oldValue: String
    newValue: String
  }

  # Input for adding a new profile
  input AddProfileInput {
    name: String!
    username: String!
    email: String!
    platform: String!
    host: String
    token: String
    sshIdentity: String
    setupSsh: Boolean
    overwrite: Boolean
  }

  # Result of an add profile operation
  type AddProfileResult {
    success: Boolean!
    message: String
    profile: Profile
  }

  # Result of a remove profile operation
  type RemoveProfileResult {
    success: Boolean!
    message: String
    profileName: String!
  }

  # Input for registering a custom platform
  input RegisterPlatformInput {
    id: String!
    name: String!
    defaultHost: String!
    sshPrefix: String!
    httpsPrefix: String!
    sshUser: String
    tokenAuthScope: String
    force: Boolean
  }

  # Result of a platform registration operation
  type RegisterPlatformResult {
    success: Boolean!
    message: String
    platform: Platform
  }

  # Result of diagnostic checks
  type DiagnosticResult {
    checks: [DiagnosticCheck!]!
    summary: String!
    overallStatus: DiagnosticStatus!
  }

  # A diagnostic check
  type DiagnosticCheck {
    name: String!
    status: DiagnosticStatus!
    message: String
    details: String
  }

  # Status of a diagnostic check
  enum DiagnosticStatus {
    PASS
    WARN
    FAIL
    INFO
  }
`
