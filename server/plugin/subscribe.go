package plugin

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/nathanaelhoun/mattermost-plugin-circleci/server/store"
)

const (
	subscribeTrigger = "subscription"
	subscribeHint    = "<" + subscribeListTrigger + "|" +
		subscribeChannelTrigger + "|" +
		subscribeUnsubscribeChannelTrigger + "|" +
		subscribeListAllChannelsTrigger + ">"
	subscribeHelpText = "Manage your subscriptions"

	subscribeListTrigger  = "list"
	subscribeListHint     = ""
	subscribeListHelpText = "List the CircleCI subscriptions for the current channel"

	subscribeChannelTrigger  = "add"
	subscribeChannelHint     = "[--flags]"
	subscribeChannelHelpText = "Subscribe the current channel to CircleCI notifications for a project"

	subscribeUnsubscribeChannelTrigger  = "remove"
	subscribeUnsubscribeChannelHint     = "[--flags]"
	subscribeUnsubscribeChannelHelpText = "Unsubscribe the current channel to CircleCI notifications for a project"

	subscribeListAllChannelsTrigger  = "list-channels"
	subscribeListAllChannelsHint     = ""
	subscribeListAllChannelsHelpText = "List all channels in the current team subscribed to a project"
)

func getSubscribeAutoCompleteData() *model.AutocompleteData {
	subscribe := model.NewAutocompleteData(subscribeTrigger, subscribeHint, subscribeHelpText)

	subscribeList := model.NewAutocompleteData(subscribeListTrigger, subscribeListHint, subscribeListHelpText)

	subscribeChannel := model.NewAutocompleteData(subscribeChannelTrigger, subscribeChannelHint, subscribeChannelHelpText)
	subscribeChannel.AddNamedTextArgument(store.FlagOnlyFailedBuilds, "Only receive notifications for failed builds", "[write anything here]", "", false) // TODO use boolean flag when then are available. See https://github.com/mattermost/mattermost-server/pull/14810
	subscribeChannel.AddNamedTextArgument(namedArgProjectName, namedArgProjectHelpText, namedArgProjectHint, namedArgProjectPattern, false)

	unsubscribeChannel := model.NewAutocompleteData(subscribeUnsubscribeChannelTrigger, subscribeUnsubscribeChannelHint, subscribeUnsubscribeChannelHelpText)
	unsubscribeChannel.AddNamedTextArgument(namedArgProjectName, namedArgProjectHelpText, namedArgProjectHint, namedArgProjectPattern, false)

	listAllSubscribedChannels := model.NewAutocompleteData(subscribeListAllChannelsTrigger, subscribeListAllChannelsHint, subscribeListAllChannelsHelpText)
	listAllSubscribedChannels.AddNamedTextArgument(namedArgProjectName, namedArgProjectHelpText, namedArgProjectHint, namedArgProjectPattern, false)

	subscribe.AddCommand(subscribeList)
	subscribe.AddCommand(subscribeChannel)
	subscribe.AddCommand(unsubscribeChannel)
	subscribe.AddCommand(listAllSubscribedChannels)

	return subscribe
}

func (p *Plugin) executeSubscribe(context *model.CommandArgs, circleciToken string, config *store.Config, split []string) (*model.CommandResponse, *model.AppError) {
	subcommand := commandHelpTrigger
	if len(split) > 0 {
		subcommand = split[0]
	}

	switch subcommand {
	case commandHelpTrigger:
		return p.sendHelpResponse(context, subscribeTrigger)

	case subscribeListTrigger:
		return executeSubscribeList(p, context)

	case subscribeChannelTrigger:
		return executeSubscribeChannel(p, context, config, split[1:])

	case subscribeUnsubscribeChannelTrigger:
		return executeUnsubscribeChannel(p, context, config)

	case subscribeListAllChannelsTrigger:
		return executeSubscribeListAllChannels(p, context, config)

	default:
		return p.sendIncorrectSubcommandResponse(context, subscribeTrigger)
	}
}

func executeSubscribeList(p *Plugin, context *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	allSubs, err := p.Store.GetSubscriptions()
	if err != nil {
		p.API.LogError("Unable to get subscriptions", "err", err)
		return p.sendEphemeralResponse(context, "Internal error when getting subscriptions"), nil
	}

	subs := allSubs.GetSubscriptionsByChannel(context.ChannelId)
	if subs == nil {
		return p.sendEphemeralResponse(
			context,
			fmt.Sprintf(
				"This channel is not subscribed to any repository. Try `/%s %s %s`",
				commandTrigger,
				subscribeTrigger,
				subscribeChannelTrigger,
			),
		), nil
	}

	attachment := model.SlackAttachment{
		Title:    "Repositories this channel is subscribed to :",
		Fallback: "List of repositories this channel is subscribed to",
	}

	for _, sub := range subs {
		p.API.LogDebug("Parsing CircleCI subscription", "sub", sub)

		username := "Unknown user"
		if user, appErr := p.API.GetUser(sub.CreatorID); appErr != nil {
			p.API.LogError("Unable to get username", "userID", sub.CreatorID)
		} else {
			username = user.Username
		}

		attachment.Fields = append(attachment.Fields, sub.ToSlackAttachmentField(username))
	}

	p.sendEphemeralPost(context, "", []*model.SlackAttachment{&attachment})
	return &model.CommandResponse{}, nil
}

func executeSubscribeChannel(p *Plugin, context *model.CommandArgs, config *store.Config, split []string) (*model.CommandResponse, *model.AppError) {
	// ? TODO check that project exists

	subs, err := p.Store.GetSubscriptions()
	if err != nil {
		p.API.LogError("Unable to get subscriptions", "err", err)
		return p.sendEphemeralResponse(context, "Internal error when getting subscriptions"), nil
	}

	newSub := &store.Subscription{
		ChannelID:  context.ChannelId,
		CreatorID:  context.UserId,
		Owner:      config.Org,
		Repository: config.Project,
		Flags:      store.SubscriptionFlags{},
	}

	for _, arg := range split[2:] {
		if strings.HasPrefix(arg, "--") {
			flag := arg[2:]
			err := newSub.Flags.AddFlag(flag)
			if err != nil {
				return p.sendEphemeralResponse(context, fmt.Sprintf(
					"Unknown subscription flag `%s`. Try `/%s %s %s`",
					arg,
					commandTrigger,
					subscribeTrigger,
					commandHelpTrigger,
				)), nil
			}
		}
	}

	p.API.LogDebug("Adding a new subscription", "subscription", newSub)
	subs.AddSubscription(newSub)

	if err := p.Store.StoreSubscriptions(subs); err != nil {
		p.API.LogError("Unable to store subscriptions", "error", err)
		return p.sendEphemeralResponse(context, "Internal error when storing new subscription."), nil
	}

	msg := fmt.Sprintf(
		"This channel has been subscribed to notifications from %s with flags: %s\n"+
			"#### How to finish setup:\n"+
			"(See the full guide [here](%s/blob/master/docs/HOW_TO.md#subscribe-to-webhooks-notifications))\n"+
			"1. Setup the [Mattermost Plugin Notify Orb](https://circleci.com/developer/orbs/orb/nathanaelhoun/mattermost-plugin-notify) for your CircleCI project\n"+
			"2. Add the `MM_WEBHOOK` environment variable to your project using `/%s %s %s %s` or the [CircleCI UI](https://circleci.com/docs/2.0/env-vars/#setting-an-environment-variable-in-a-project)\n"+
			"**Webhook URL: `%s`**",
		config.ToMarkdown(),
		newSub.Flags,
		manifest.HomepageURL,
		commandTrigger,
		projectTrigger,
		projectEnvVarTrigger,
		projectEnvVarAddTrigger,
		p.getWebhookURL(),
	)

	return p.sendEphemeralResponse(context, msg), nil
}

func executeUnsubscribeChannel(p *Plugin, args *model.CommandArgs, config *store.Config) (*model.CommandResponse, *model.AppError) {
	subs, err := p.Store.GetSubscriptions()
	if err != nil {
		p.API.LogError("Unable to get subscriptions", "err", err)
		return p.sendEphemeralResponse(args, "Internal error when getting subscriptions"), nil
	}

	if removed := subs.RemoveSubscription(args.ChannelId, config.Org, config.Project); !removed {
		return p.sendEphemeralResponse(args,
			fmt.Sprintf("This channel was not subscribed to %s", config.ToMarkdown()),
		), nil
	}

	if err := p.Store.StoreSubscriptions(subs); err != nil {
		p.API.LogError("Unable to store subscriptions", "error", err)
		return p.sendEphemeralResponse(args, "Internal error when storing new subscription."), nil
	}

	return p.sendEphemeralResponse(args,
		fmt.Sprintf("Successfully unsubscribed this channel from %s", config.ToMarkdown()),
	), nil
}

func executeSubscribeListAllChannels(p *Plugin, context *model.CommandArgs, config *store.Config) (*model.CommandResponse, *model.AppError) {
	allSubs, err := p.Store.GetSubscriptions()
	if err != nil {
		p.API.LogError("Unable to get subscriptions", "err", err)
		return p.sendEphemeralResponse(context, "Internal error when getting subscriptions"), nil
	}

	channelIDs := allSubs.GetSubscribedChannelsForRepository(config.Org, config.Project)
	if channelIDs == nil {
		return p.sendEphemeralResponse(
			context,
			fmt.Sprintf(
				"No channel is subscribed to the project %s. Try `/%s %s %s`",
				config.ToMarkdown(),
				commandTrigger,
				subscribeTrigger,
				subscribeChannelTrigger,
			),
		), nil
	}

	message := fmt.Sprintf("Channels of this team subscribed to %s\n", config.ToMarkdown())
	for _, channelID := range channelIDs {
		channel, appErr := p.API.GetChannel(channelID)
		if appErr != nil {
			p.API.LogError("Unable to get channel", "channelID", channelID)
		}

		message += fmt.Sprintf("- ~%s\n", channel.Name)
	}

	return p.sendEphemeralResponse(context, message), nil
}
