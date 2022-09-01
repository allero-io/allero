<p align="center">
 <img src="./static/allero_banner.png" alt="allero=github" border="0" />
</p>

<h1 align="center">
 Protecting Your Production Pipelines!
</h1>


## What is Allero?
Allero is a CLI policy enforcement tool that prevents bad practices in any CI/CD pipeline.
CI/CD pipelines tend to be messy, and there are so many variations of pipeline manifests spread across different repositories.
This makes it difficult to ensure security, code quality, and compliance standards are in place in every pipeline.

By running Allero, you can easily reveal and prevent problematic pipelines across multiple oragnizations and repositories.


## Getting Started
Allero CLI can be run from anywhere! We recommend running Allero directly from a GitHub Action to ensure bad practices are validated on a regular basis (just like crontab).

### 🏎️ One minute installation to run allero validation on a daily basis (most recommended)
Allero repo has a GitHub Action that runs the CLI every day at 8am on your entire organization. By forking the `allero` repo you'll get the same setup.
1. [Fork](https://github.com/allero-io/allero/fork) Allero repo
2. Create a GitHub Personal Access Token and store it in your forked repo as an encrypted secret named `ALLERO_GITHUB_TOKEN`.
3. GitHub disables scheduled Actions on a forked repo by default. To enable the Allero Action, browse to your forked allero repo, navigate to GitHub Actions and click enable workflow. 

* You can of course change the schedule and the fetched repos by editing the workflow file!

### 👩‍💻 CLI Installation
Since Allero is a CLI, you can run it everywhere - including your local machine! Download our CLI now!

```bash
# Get allero cli
curl https://get.allero.io | /bin/bash
# Fetch one or more organizations / repos
allero fetch github allero-io dapr/dapr
# Run allero validation!
allero validate
```
### Homebrew
```bash
# Install allero cli
brew install allero-io/allero/allero
 # Fetch one or more organizations / repos
allero fetch github allero-io dapr/dapr
# Run allero validation!
allero validate
```
#### GitHub Token
Fetching data from a private GitHub organization requires a personal access token (PAT).
1. Create a GitHub PAT with access to the repos you want to scan. Click [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-token) to learn how to create a Github PAT.
Generate the token with the following permissions:
    - [x]  repo:
        - [x]  repo:status
        - [x]  repo_deployment
        - [x]  public_repo
        - [x]  repo:invite
        - [x]  security_events

2. The PAT should be stored as an environment variable named `ALLERO_GITHUB_TOKEN`.
- When running Allero from GitHub Actions, the PAT should be stored as an [encrypted secret](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-a-repository).

## 🚨 Supported Rules
| _Rule Name_               | _Description_                                            | _Reason_                                                                                                               |
| ------------------------- | -------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| prevent-npm-install       | Prevents the usage of `npm install` in pipelines. We recommend using `npm ci` instead           | [link](https://betterprogramming.pub/npm-ci-vs-npm-install-which-should-you-use-in-your-node-js-projects-51e07cb71e26) |
| prevent-kubectl-apply     | Prevents the usage of kubectl apply in pipelines. We recommend using helm or any other k8s deployment tool         | [link](https://medium.com/@RedBaronDr1/helm-vs-kubectl-5aaf2dba7d71)                                                   |
| ensure-npm-ignore-scripts | Ensures that pre/post-install scripts are not run by NPM | [link](https://snyk.io/blog/ten-npm-security-best-practices/)                                                     |
snyk-prevent-continue-on-error | Prevent continuing workflows when snyk detects vulnerabilities | Keep production secured
prevent-password-plain-text | Prevent use of password as plain text | Keep passwords from leaking
ensure-node-version | Make sure a specific version is set when using a node image | Avoid unexpected behavior
ensure-python-version | Make sure a specific version is set when using a python image | Avoid unexpected behavior
ensure-github-action-version | Ensure github action version is set | Avoid unexpected behavior

### Adding your own rules
Rules can be defined using the [Json Schema](https://json-schema.org/) format. Json Schema rules should be based on our data schema. An example of our data schema structure can be found [here](https://github.com/allero-io/allero/tree/main/examples/github/data-schema-example.json).
1. Create a new json file and define your rule. Example rules can be found [here](https://github.com/allero-io/allero/tree/main/examples).
Make sure to update the rule description and failureMessage.
2. Copy-paste the file to "~/.allero/rules/github/"
3. Run `allero validate`
## Contribution
We encourage you to contribute to Allero!
#### Created a new rule and want to give back to the community?
1. **Fork our repo**
2. **Add your rule to [pkg/rulesConfig/github](https://github.com/allero-io/allero/tree/main/pkg/rulesConfig/defaultRules/github) directory.**
3. **Create a PR!**

**Interested in contributing more to the CLI?**
We will provide a more detailed explanation on how to contribute soon. If you're intrested, you can [contact us](mailto:contact@allero.io) to get our help with your first PR!

## 🔏 Privacy
Your privacy and code integrity are very important to us. That's why our CLI operates locally only, and doesn't save any sensitive information related to your code anywhere. We only track metrics that reflect your usage of the CLI :)
### Contact Us
Open an issue or shoot us an [email](mailto:contact@allero.io).
