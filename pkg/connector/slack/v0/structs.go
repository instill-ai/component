package slack

type SlackChannelApiResp struct {
	Ok       bool           `json:"ok"`
	Channels []SlackChannel `json:"channels"`
}

type SlackChannel struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	IsChannel          bool          `json:"is_channel"`
	IsGroup            bool          `json:"is_group"`
	IsIM               bool          `json:"is_im"`
	Created            int64         `json:"created"`
	Creator            string        `json:"creator"`
	IsArchived         bool          `json:"is_archived"`
	IsGeneral          bool          `json:"is_general"`
	Unlinked           int           `json:"unlinked"`
	NameNormalized     string        `json:"name_normalized"`
	IsShared           bool          `json:"is_shared"`
	IsExtShared        bool          `json:"is_ext_shared"`
	IsOrgShared        bool          `json:"is_org_shared"`
	PendingShared      []interface{} `json:"pending_shared"`
	IsPendingExtShared bool          `json:"is_pending_ext_shared"`
	IsMember           bool          `json:"is_member"`
	IsPrivate          bool          `json:"is_private"`
	IsMPIM             bool          `json:"is_mpim"`
	Updated            int64         `json:"updated"`
	Topic              struct {
		Value   string `json:"value"`
		Creator string `json:"creator"`
		LastSet int    `json:"last_set"`
	} `json:"topic"`
	Purpose struct {
		Value   string `json:"value"`
		Creator string `json:"creator"`
		LastSet int    `json:"last_set"`
	} `json:"purpose"`
	PreviousNames []interface{} `json:"previous_names"`
	NumMembers    int           `json:"num_members"`
}

type SendingSlackMessage struct {
	ChannelName string `json:"channel_name"`
	Message     string `json:"message"`
}

type SendingData struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type SendingMessageResp struct {
	OK      bool   `json:"ok"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	Error   string `json:"error"`
	Needed  string `json:"needed"`
	Message struct {
		User       string `json:"user"`
		Type       string `json:"type"`
		Ts         string `json:"ts"`
		BotID      string `json:"bot_id"`
		AppID      string `json:"app_id"`
		Text       string `json:"text"`
		Team       string `json:"team"`
		BotProfile struct {
			ID    string `json:"id"`
			AppID string `json:"app_id"`
			Name  string `json:"name"`
			Icons struct {
				Image36 string `json:"image_36"`
				Image48 string `json:"image_48"`
				Image72 string `json:"image_72"`
			} `json:"icons"`
			Deleted bool   `json:"deleted"`
			Updated int    `json:"updated"`
			TeamID  string `json:"team_id"`
		} `json:"bot_profile"`
	} `json:"message"`
	Blocks []struct {
		Type     string `json:"type"`
		BlockID  string `json:"block_id"`
		Elements []struct {
			Type     string `json:"type"`
			Elements []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"elements"`
		} `json:"elements"`
	} `json:"blocks"`
}


type WriteTaskResp struct {
	Result string `json:"result"`
}