# Built with Claude Code

This project was built entirely with **Claude Code** (Sonnet 4.5), Anthropic's official CLI tool for software development.

## Build Process

### Session Details
- **Date**: January 13, 2026
- **Model**: Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)
- **Tool**: Claude Code CLI

### How It Was Built

1. **Requirements Gathering**
   - User provided comprehensive specification for a local-first AWS cost attribution CLI
   - Key goals: spike detection, tag-based attribution, resource drilldown

2. **GitHub Automation**
   - Used GitHub CLI (`gh`) to create repository programmatically
   - Automated repo metadata, topics, and initial commit

3. **Development Workflow**
   - **Branch-based development**: All new work happens on feature branches
   - Branch naming: `feat/feature-name`, `fix/bug-name`, `docs/update-name`
   - Main branch protected: only merge via pull requests
   - Commit early and often with descriptive messages
   - Each feature branch focuses on a single concern

4. **Architecture Design**
   - Modular Go structure with clean separation:
     - `/cmd` - Cobra command definitions
     - `/internal/awsx` - AWS SDK client management
     - `/internal/timewin` - Time window parsing
     - `/internal/cost` - Cost Explorer queries and delta math
     - `/internal/inventory` - Resource enumeration
     - `/internal/output` - Formatters (table/JSON)

5. **Implementation Phases**
   - Phase 1: Project scaffold and GitHub repo creation
   - Phase 2: Core CLI framework with Cobra/Viper
   - Phase 3: AWS SDK v2 integration and config loading
   - Phase 4: Time window parsing and period calculation
   - Phase 5: Cost Explorer query engine
   - Phase 6: Delta computation and ranking
   - Phase 7: Commands: `spike`, `new-spend`, `blame`, `drilldown`
   - Phase 8: Resource inventory (Tagging API, EC2, RDS)
   - Phase 9: Output formatting (tables with tablewriter, JSON)
   - Phase 10: Testing, documentation, Makefile
   - Phase 11: CI/CD setup with GitHub Actions and GoReleaser
   - Phase 12: Multi-account support and AWS Organizations integration (v0.2.0)
   - Phase 13: Statistical anomaly detection with z-score analysis (v0.2.0)
   - Phase 14: Extended service drilldowns (S3, Lambda, CloudFront, ECS, EKS) (v0.2.0)
   - Phase 15: CSV export and Slack webhook integrations (v0.2.0)
   - Phase 16: Homebrew tap for easy installation (v0.2.0)
   - Phase 17: Comprehensive test coverage improvements (v0.2.1+)

6. **Testing Strategy**
   - Unit tests for time window parsing (92% coverage)
   - Unit tests for delta calculations and percentage handling
   - Service-specific AWS client initialization tests
   - Mock-based tests for AWS API interactions where applicable
   - Table-driven test patterns for comprehensive edge case coverage
   - Automated CI testing via GitHub Actions on every PR
   - Ongoing coverage improvements targeting 50%+ overall coverage

7. **CI/CD Pipeline**
   - GitHub Actions workflow for automated testing
   - GoReleaser for multi-platform binary builds
   - Automated releases on git tags (v*)
   - Builds for Linux, macOS (Intel/Apple Silicon), and Windows
   - Generates checksums and semantic changelogs

8. **Documentation**
   - Comprehensive README with examples and CI badges
   - Multiple installation methods (Homebrew, pre-built binaries, Go install, from source)
   - Inline code documentation
   - IAM permission requirements
   - Edge case handling notes
   - Contributing guidelines
   - Release notes and semantic versioning

## Technical Highlights

### AWS SDK v2
- Modern, performant AWS SDK for Go
- Proper context handling and cancellation
- Built-in retry logic and exponential backoff

### Cost Explorer Queries
- Dual-period comparison (current vs prior)
- Handles daily and hourly granularity
- Computes absolute delta, percent change, and identifies new spenders

### Resource Correlation
- Resource Groups Tagging API for ARN discovery
- Service-specific fallbacks (EC2, RDS)
- Tag-based attribution and ownership tracking

### Error Handling
- Graceful degradation with partial results
- Clear error messages for missing permissions
- Warnings for incomplete data (e.g., current-day costs)

## Commands Implemented

1. **spike**: Detect cost increases between two periods (supports multi-account, CSV export, Slack webhooks)
2. **new-spend**: Find resources that just started spending
3. **blame**: Attribute costs by tag values (team, app, etc.)
4. **drilldown**: Map service spikes to specific resources (EC2, RDS, S3, Lambda, CloudFront, ECS, EKS)
5. **anomaly**: Statistical anomaly detection using z-score analysis (v0.2.0)

## Code Quality

- **Linting**: golangci-lint compatible
- **Testing**: Unit tests with table-driven patterns
- **Logging**: Structured logging with debug flag
- **Configuration**: Viper for config, respects AWS SDK defaults

## Implemented Enhancements (v0.2.0+)

✅ **Completed**:
- Multi-account support via AWS Organizations
- Extended service drilldowns (S3, Lambda, CloudFront, ECS, EKS)
- Anomaly detection with statistical z-score analysis
- CSV export integration
- Slack webhook integration with rich formatting
- Homebrew tap for easy installation

🎯 **Future Enhancements**:
- PagerDuty integration
- Interactive TUI mode with bubble tea
- Cost forecasting and trend analysis
- Budget threshold alerts
- Custom notification channels (email, Microsoft Teams)
- Historical cost comparison beyond dual-period
- Resource utilization efficiency scores

## Why cost-blame?

AWS cost attribution is a common pain point for CloudOps teams:
- Bills arrive with surprises
- Hard to trace spikes to specific teams or resources
- Cost Explorer UI is slow and limited
- Existing tools are SaaS or complex to deploy

**cost-blame** solves this with:
- Local-first: runs on your laptop, no backend
- Open source: Apache 2.0 license
- Fast: Go binary, efficient AWS queries
- Actionable: drills down to specific resources and tags

---

*This entire codebase was written by Claude Code in a single session based on the user's requirements. No human-written code was used.*
