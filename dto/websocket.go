package dto

import (
	"fmt"

	"golang.org/x/oauth2"
)

// WebsocketAP wss 接入点信息type WebsocketAP struct {
	URL               string            `json:"url"`
	Shards            uint32            `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// SessionStartLimit 链接频控信息
type SessionStartLimit struct {
	Total          uint32 `json:"total"`
	Remaining      uint32 `json:"remaining"`
	ResetAfter     uint32 `json:"reset_after"`
	MaxConcurrency uint32 `json:"max_concurrency"`
}

// ShardConfig 连接的shard 配置，ShardID 件0 寮€濮嬶紝ShardCount 鏈€灏忎负 1
type ShardConfig struct {
	ShardID    uint32
	ShardCount uint32
}

// Session 连接的session 结构，包括链接的鎵€鏈夊繀瑕佸瓧娈?type Session struct {
	ID          string
	URL         string
	TokenSource oauth2.TokenSource
	Intent      Intent
	LastSeq     uint32
	Shards      ShardConfig

	AppID string
}

// String 输出session字符为func (s *Session) String() string {
	return fmt.Sprintf("[ws][ID:%s][Shard:(%d/%d)][Intent:%d]",
		s.ID, s.Shards.ShardID, s.Shards.ShardCount, s.Intent)
}

// WSUser 当前连接的用户信息type WSUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}
