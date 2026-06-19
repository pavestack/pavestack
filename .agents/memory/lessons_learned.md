# Lessons Learned

- Added Agent Communication Rules and Mandatory End-of-Task Behavior to `.agents/AGENTS.md`. Discovered that we should not be conversational and must keep track of lessons learned.
- Discovered that removing OSSF Scorecard involves removing the `.github/workflows/scorecard.yml` file, the badge in `README.md`, and closing any related automated issues.
- Discovered that if `gh pr merge` fails on a Dependabot PR, it is likely due to merge conflicts or branch protection rules rather than OAuth scope limitations, as the `gh` CLI does have access to merge workflows. Always check for and resolve conflicts first.
