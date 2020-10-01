# Mattermost Plugin CircleCI [![CircleCI branch](https://img.shields.io/circleci/project/github/nathanaelhoun/mattermost-plugin-circleci/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-circleci)

A Work-In-Progress [CircleCI](https://circleci.com) plugin to interact with jobs and builds, with slash commands in Mattermost.

To learn more about plugins, see [the Mattermost plugin documentation](https://developers.mattermost.com/extend/plugins/).

This plugin uses the CircleCI Orb for Mattermost Plugin by **[@nathanaelhoun](https://github.com/nathanaelhoun)**: [check it out here](https://github.com/nathanaelhoun/circleci-orb-mattermost-plugin-notify).

**This plugin is under development and is not ready for production**

## Features

#### Connect to your CircleCI account

-   `/circleci account view` - Get informations about yourself
-   `/circleci account connect <API token>` - Connect your Mattermost account to CircleCI
-   `/circleci account disconnect` - Disconnect your Mattermost account from CircleCI

#### Manage CircleCI projects

-   `/circleci project list-followed` - List followed projects
-   `/circleci project recent-build <owner-name> <project-name> <branch>` - List the 10 last builds for a project

#### Subscribe to notifications projects

-   `/circleci subscription list` — List the CircleCI subscriptions for the current channel
-   `/circleci subscription subscribe <owner> <repository> [--flags]` — Subscribe the current channel to CircleCI notifications for a repository
-   `/circleci subscription unsubscribe <owner> <repository> [--flags]` — Unsubscribe the current channel to CircleCI notifications for a repository
-   `/circleci subscription list-channels <owner> <repository>` — List all channels subscribed to this repository in the current team

#### Config

-   `/circleci config <vcs/org-name/project-name>` - Allows you to set a default project to run your commands against

## TODO (tracking list)

-   [x] Get help

-   [x] Connect to CircleCI, see your profile, disconnect

-   [x] Setup webhook notifications about successful and failed CircleCI builds

-   [ ] Interact with CircleCI jobs

    -   [ ] Trigger jobs with and without parameters
    -   [ ] Abort a job
    -   [ ] Configure/create/delete jobs
    -   [ ] Get logs from a job in a file attachment, not as a message (this is because the logs can be huge, so it’s easier to preview a file attachment)
    -   [ ] Get artifacts
    -   [ ] Get test results

-   [ ] Meet [requirements](https://developers.mattermost.com/extend/plugins/community-plugin-marketplace/#requirements-for-adding-a-community-plugin-to-the-marketplace) to publish to Marketplace

## Installation

_Coming_

## Contributing

### I saw a bug, I have a feature request or a suggestion

Please fill a [Github Issue](https://github.com/nathanaelhoun/mattermost-plugin-circleci/issues/new/choose), it will be very useful!

### I want to code

Pull Requests are welcome! You can contact me on the [Mattermost Community ~plugin-circleci channel](https://community.mattermost.com/core/channels/plugin-circleci) where I am `@nathanaelhoun`.

## Development

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options.

### Deploying with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. After configuring it, just run:

```
make deploy
```

### Deploying with credentials

Alternatively, you can authenticate with the server's API with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):

```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

## License

Apache License.

## Thanks to

-   **[@jszwedko](https://github.com/jszwedko)** for his [CircleCI v1 Go API](https://github.com/jszwedko/go-circleci)
-   **[@darkLord19](https://github.com/darkLord19)** for this [CircleCI v2 Go API](https://github.com/darkLord19/circleci-v2/circleci)
-   [Mattermost](https://mattermost.org) for providing a good software and having a great community
