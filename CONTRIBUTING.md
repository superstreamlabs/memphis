# Contributing

We would ❤️ it if you contributed to the project and helped make Memphis{dev} even better.<br>
We will make sure that contributing to Memphis{dev} is easy, enjoyable, and educational for anyone and everyone.<br>
All contributions are welcome, including features, issues, documentation, guides, and more.

Our beloved contributors also recieve occasional Memphis Swag Pack and publications where every Memphis is!

![contribution tweet](https://user-images.githubusercontent.com/70286779/173819239-3611bbb2-3f7d-41a2-ad37-2e966b91403c.jpg)

<hr>
<div align="center">

  <h1>Process</h1>
  
![contribution guidelinues](https://user-images.githubusercontent.com/70286779/173803841-1c77d8d9-d378-4632-872d-c782ce61f2a3.png)

</div>

## Step 1: Choose a task

[Open tasks table](https://github.com/orgs/memphisdev/projects/1)

## Step 2: Make a fork

Fork the Memphis-broker repository to your GitHub organization. This means that you'll have a copy of the repository under _your-GitHub-username/repository-name_.

## Step 3: Clone the repository to your local machine

```
git clone https://github.com/{your-GitHub-username}/memphis-<component>.git
```

## Step 4: Start a dev environment

### For the broker
https://github.com/memphisdev/memphis-broker

Run the following docker-compose, it will start a local Memphis environment without the Broker
```
curl -s https://memphisdev.github.io/memphis-docker/docker-compose-dev-broker.yaml -o docker-compose.yaml && \
docker compose -f docker-compose.yaml -p memphis up
```

Run locally via VS code debugger

Install GO, and then -
```
go get -d -v .
go install -v .
```
and then click F5 in order to start the debugger

### For the UI
https://github.com/memphisdev/memphis-ui

Run the following docker-compose, it will start a local Memphis environment without the UI
```
curl -s https://memphisdev.github.io/memphis-docker/docker-compose-dev-ui.yaml -o docker-compose.yaml && \
docker compose -f docker-compose.yaml -p memphis up
```

Inside the cloned directory
```
npm install
npm start
```

### Tech stack

**Server-side**
- [Go](https://go.dev/) (Broker / SDK)
- [Python](https://www.python.org/) (SDK)
- [Node.JS](https://nodejs.org/) (CLI / SDK)

**Client-side**
- [React](https://reactjs.org/docs/getting-started.html) (UI)

**Scripting**
- [Helm](https://helm.sh/) (Deployment)
- [Docker](https://docker.com) (Deployment)
- [Bash](https://www.gnu.org/software/bash/) (Deployment)

## Step 5: Create a branch

Create a new branch (from staging branch) for your fix.

```jsx
git checkout -b branch-name-here staging
```

## Step 6: Make your changes

Update the code with your bug fix or new feature.

## Step 7: Add the changes that are ready to be committed

Stage the changes that are ready to be committed:

```jsx
git add .
```

## Step 8: Commit the changes (Git)

Commit the changes with a short message. (See below for more details on how we structure our commit messages)

```jsx
git commit -m "<type>(<package>): <subject>"
```

## Step 9: Push the changes to the remote repository

Push the changes to the remote repository using:

```jsx
git push origin branch-name-here
```

## Step 10: Create Pull Request

In GitHub, do the following to submit a pull request to the upstream repository (staging branch):

1.  Give the pull request a title and a short description of the changes made. Include also the issue or bug number associated with your change. Explain the changes that you made, any issues you think exist with the pull request you made, and any questions you have for the maintainer.

Remember, it's okay if your pull request is not perfect (no pull request ever is). The reviewer will be able to help you fix any problems and improve it!

2.  Wait for the pull request to be reviewed by a maintainer.

3.  Make changes to the pull request if the reviewing maintainer recommends them.

Celebrate your success after your pull request is merged :-)

## Git Commit Messages

We structure our commit messages like this:

```
<type>(<package>): <subject>
```

Example

```
fix(server): missing entity on init
```

## Got a question?

You can ask questions, consult with more experienced Memphis{dev} users, and discuss Memphis-related topics in the our [Discord channel](https://discord.gg/WZpysvAeTf).

## Found a bug?

If you find a bug in the source code, you can help us by [submitting an issue](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20bug) to our GitHub Repository.<br>
Even better, you can submit a Pull Request with a fix.

## Missing a feature?

You can request a new feature by [submitting an issue](https://github.com/memphisdev/memphis-broker/issues/new?assignees=&labels=type%3A%20feature%20request) to our GitHub Repository.

If you'd like to implement a new feature, it's always good to be in touch with us before you invest time and effort, since not all features can be supported.

- For a Major Feature, first open an issue and outline your proposal. This will let us coordinate efforts, prevent duplication of work, and help you craft the change so that it's successfully integrated into the project.
- Small Features can be crafted and directly [submitted as a Pull Request](#submit-pr).

### Types:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Changes to the documentation
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc.)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Packages:

- **server**
- **client**
- **data-service-gen**

## Code of conduct

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.

[Code of Conduct](https://github.com/memphisdev/memphis-broker/blob/master/code_of_conduct.md)

Our Code of Conduct means that you are responsible for treating everyone on the project with respect and courtesy.
