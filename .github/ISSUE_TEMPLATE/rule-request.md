---
name: Rule request
about: Requesting a new rule
title: ''
labels: rule requests
assignees: ''

---

**Name of the rule you'd like to add**
Add rule name here. Prevent/Ensure ...

**Describe the rule**
A clear description in a single sentence

**What triggers the rule**
Add a detailed description of what behaviors should trigger this rule. This will help us best understand how to implement it

 **Failure message should the rule fail**
Short message for when this rule fails

**What SCMs is this rule eligible for**
Add here one or more: Github, Gitlab, etc.

**What CI/CD platforms is this rule eligible for**
Add here one or more: Github Actions, GitlabCI, JFrog Pipelines, etc.

**Should this rule be enabled by default**
Rules with clear failing scenarios should be enabled. Conditional rules that require a specific state (for example: helm charts related rule) should be opt-in and disabled by default

**Sample repos/orgs to test the rule**
If possible, add a link to one or more repositories/organizations that qualifies for testing the rule
