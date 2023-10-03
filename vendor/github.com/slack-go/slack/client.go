package slack

import (
	"context"
	"net/url"
)

type ThreadsInfo struct {
	HasUnreads            int   `json:"has_unreads"`
	MentionCount          int   `json:"mention_count"`
	MentionCountByChannel []int `json:"mention_count_by_channel"`
	UnreadCountByChannel  []int `json:"unread_count_by_channel"`
}

type CountInfo struct {
	ID           string `json:"id"`
	LastRead     string `json:"last_read"`
	Latest       string `json:"latest"`
	Updated      string `json:"update"`
	MentionCount int    `json:"mention_count"`
	HasUnreads   bool   `json:"has_unreads"`
}

type ChannelBadges struct {
	Channels       int `json:"channels"`
	DMS            int `json:"dms"`
	AppDMS         int `json:"app_dms"`
	ThreadMentions int `json:"thread_mentions"`
	ThreadUnreads  int `json:"thread_unreads"`
}

// type FileChannels struct {
// 	Quip Quip `json:"quip"`
// }

type Quip struct {
	MentionCountByChannel []int `json:"mention_count_by_channel"`
}

type CountSaved struct {
	UncompletedCount        int `json:"uncompleted_count"`
	UncompletedOverdueCount int `json:"uncompleted_overdue_count"`
}

type GetChannelsInfoParams struct {
	ThreadCountsByChannel bool
	OrgWideAware          bool
	IncludeFileChannels   bool
}

func (api *Client) GetClientCounts(params *GetChannelsInfoParams) ([]CountInfo, []CountInfo, []CountInfo, error) {
	return api.GetClientCountsContext(context.Background(), params)
}

func (api *Client) GetClientCountsContext(ctx context.Context, params *GetChannelsInfoParams) ([]CountInfo, []CountInfo, []CountInfo, error) {
	values := url.Values{
		"token": {api.token},
	}
	if params.ThreadCountsByChannel {
		values.Add("thread_counts_by_channel", "true")
	}
	if params.OrgWideAware {
		values.Add("org_wide_aware", "true")
	}
	if params.IncludeFileChannels {
		values.Add("include_file_channels", "true")
	}
	response := struct {
		Channels      []CountInfo   `json:"channels,omitempty"`
		MpIms         []CountInfo   `json:"mpims,omitempty"`
		Ims           []CountInfo   `json:"ims,omitempty"`
		ChannelBadges ChannelBadges `json:"channel_badges"`
		// FileChannels  FileChannels  `json:"file_channels"`
		Saved CountSaved `json:"saved"`
		SlackResponse
	}{}

	err := api.postMethod(ctx, "client.counts", values, &response)
	if err != nil {
		return response.Channels, response.MpIms, response.Ims, err
	}

	if err := response.Err(); err != nil {
		return response.Channels, response.MpIms, response.Ims, err
	}

	return response.Channels, response.MpIms, response.Ims, nil
}
