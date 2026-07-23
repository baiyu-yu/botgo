package dto

// WSGroupRobotEventData 机器人加入/退出群聊事件数据
type WSGroupRobotEventData struct {
	GroupOpenID    string `json:"group_openid"`     // 群OpenID
	OpMemberOpenID string `json:"op_member_openid"` // 操作者OpenID
	Timestamp      int64  `json:"timestamp"`        // 时间戳（秒）
}

// WSGroupMemberAddData 群成员增加事件数据，对应 GROUP_MEMBER_ADD 事件
type WSGroupMemberAddData struct {
	GroupOpenID    string `json:"group_openid"`     // 群OpenID
	MemberOpenID   string `json:"member_openid"`    // 新成员OpenID
	OpMemberOpenID string `json:"op_member_openid"` // 操作者OpenID
	Timestamp      int64  `json:"timestamp"`        // 时间戳（秒）
}

// WSGroupMemberRemoveData 群成员减少事件数据，对应 GROUP_MEMBER_REMOVE 事件
type WSGroupMemberRemoveData struct {
	GroupOpenID    string `json:"group_openid"`     // 群OpenID
	MemberOpenID   string `json:"member_openid"`    // 被移除成员OpenID
	OpMemberOpenID string `json:"op_member_openid"` // 操作者OpenID
	Timestamp      int64  `json:"timestamp"`        // 时间戳（秒）
}
