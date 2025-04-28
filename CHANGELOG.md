# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Refactored `config.LoadConfig` to gracefully handle invalid profiles in `creds.json`. The function now loads all valid profiles and returns a map of validation errors for invalid ones, instead of failing on the first error. Commands using `LoadConfig` now report these errors as warnings.

### Fixed
- Resolved errors in the PowerShell integration test script `tests/test_02_profiles.ps1` related to profile block extraction using regex. Replaced regex block matching with procedural line-by-line parsing for robustness.
- Updated `README.md` to mention the improved handling of invalid configuration files. 