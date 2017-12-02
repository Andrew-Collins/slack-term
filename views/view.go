package views

import (
	"github.com/erroneousboat/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/service"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
	Debug    *components.Debug
}

func CreateView(svc *service.SlackService) *View {
	// Create Input component
	input := components.CreateInputComponent()

	// Channels: create the component
	channels := components.CreateChannelsComponent(input.Par.Height)

	// Channels: fill the component
	slackChans := svc.GetChannels()
	channels.SetChannels(slackChans)
	channels.SetPresenceChannels(slackChans)

	// Chat: create the component
	chat := components.CreateChatComponent(input.Par.Height)

	// Chat: fill the component
	slackMsgs := svc.GetMessages(
		svc.GetSlackChannel(channels.SelectedChannel),
		chat.GetMaxItems(),
	)
	chat.SetMessages(slackMsgs)
	chat.SetBorderLabel(svc.Channels[channels.SelectedChannel])

	// Debug: create the component
	debug := components.CreateDebugComponent(input.Par.Height)

	// Mode: create the component
	mode := components.CreateModeComponent()

	view := &View{
		Input:    input,
		Channels: channels,
		Chat:     chat,
		Mode:     mode,
		Debug:    debug,
	}

	return view
}

func (v *View) Refresh() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Mode,
	)
}
