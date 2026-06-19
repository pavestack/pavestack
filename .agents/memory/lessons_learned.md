# Lessons Learned

- Added Agent Communication Rules and Mandatory End-of-Task Behavior to `.agents/AGENTS.md`. Discovered that we should not be conversational and must keep track of lessons learned.
- Discovered that removing OSSF Scorecard involves removing the `.github/workflows/scorecard.yml` file, the badge in `README.md`, and closing any related automated issues.
- Confirmed that Dependabot PRs updating GitHub Actions workflows must be merged by commenting `@dependabot merge` instead of using `gh pr merge`, due to OAuth scope limitations on the `gh` CLI.
