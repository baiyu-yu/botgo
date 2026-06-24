package dto

// Message 消息结构体定涔?type Message struct {
	// 消息ID
	ID string `json:"id"`
	// 子频道ID
	ChannelID string `json:"channel_id"`
	// 频道ID
	GuildID string `json:"guild_id"`
	// 群ID
	GroupID string `json:"group_id"`

	// 内容
	Content string `json:"content"`
	// 鍙戦€佹椂闂?	Timestamp Timestamp `json:"timestamp"`
	// 消息编辑时间
	EditedTimestamp Timestamp `json:"edited_timestamp"`
	// 是否@all
	MentionEveryone bool `json:"mention_everyone"`
	// 消息鍙戦€佹柟
	Author *User `json:"author"`
	// 消息鍙戦€佹柟Author的member灞炴€э紝鍙槸閮ㄥ垎灞炴€?	Member *Member `json:"member"`
	// 附件
	Attachments []*MessageAttachment `json:"attachments"`
	// 结构化消息embeds
	Embeds []*Embed `json:"embeds"`
	// 消息中的提醒信息(@)列表
	Mentions []*User `json:"mentions"`
	// ark 消息
	Ark *Ark `json:"ark"`
	// 私信消息
	DirectMessage bool `json:"direct_message"`
	// 子频閬?seq，用于消息间的排序，seq 在同涓€瀛愰閬撲腑鎸変粠鍏堝埌鍚庣殑椤哄簭閫掑锛屼笉鍚岀殑瀛愰閬撲箣鍓嶆秷鎭棤娉曟帓搴?	SeqInChannel string `json:"seq_in_channel"`
	// 引用的消息	MessageReference *MessageReference `json:"message_reference,omitempty"`
	// 私信场景下，该字段用来标识从哪个频道发起的私淇?	SrcGuildID string `json:"src_guild_id"`
	// 上传富媒体文件后返回的文件信鎭€?注意以群鎴栬€匔2C消息上传后， 同类型可以重复使用，不同类型闇€瑕佷笉鑳戒娇鐢ㄣ€?	FileInfo []byte `json:"file_info,omitempty"`
	// 上传富媒体文件后的有效期, 单位:绉? 在有效期内可以重复使鐢ㄣ€?	TTL uint `json:"ttl,omitempty"`
	// 消息场景描述
	MessageScene MessageScene `json:"message_scene,omitempty"`
}

// Embed 结构
type Embed struct {
	Title       string                `json:"title,omitempty"`
	Description string                `json:"description,omitempty"`
	Prompt      string                `json:"prompt"` // 消息弹窗内容，消息列表摘瑕?	Thumbnail   MessageEmbedThumbnail `json:"thumbnail,omitempty"`
	Fields      []*EmbedField         `json:"fields,omitempty"`
}

// MessageEmbedThumbnail embed 消息的缩略图对象
type MessageEmbedThumbnail struct {
	URL string `json:"url"`
}

// EmbedField Embed字段描述
type EmbedField struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// MessageAttachment 附件定义
type MessageAttachment struct {
	URL         string `json:"url,omitempty"`
	FileName    string `json:"filename,omitempty"`
	Height      int    `json:"height,omitempty"`
	Size        int    `json:"size,omitempty"`
	Width       int    `json:"width,omitempty"`
	ContentType string `json:"content_type,omitempty"` // voice:语音, image/xxx: 图片 video/xxx: 视频
}

// MessageReactionUsers 消息表情琛ㄦ€佺敤鎴峰垪琛?type MessageReactionUsers struct {
	Users  []*User `json:"users,omitempty"`
	Cookie string  `json:"cookie,omitempty"`
	IsEnd  bool    `json:"is_end,omitempty"`
}

// MessageScene 消息场景
type MessageScene struct {
	Source       string `json:"source,omitempty"`        // 消息来源, realtime_voice: 实时通话场景, ai_search: AI搜索 其它默认为AIO消息
	CallbackData string `json:"callback_data,omitempty"` // 回调数据
}
