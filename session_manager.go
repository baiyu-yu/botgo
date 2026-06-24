package botgo

import (
	"github.com/sealdice/botgo/dto"
	"github.com/sealdice/botgo/sessions/local"
	"golang.org/x/oauth2"
)

// defaultSessionManager 默认实现的session manager 为单机版有// 如果业务要自行实现分布式的session 管理，则实现 SessionManger 后替换掉 defaultSessionManager
var defaultSessionManager SessionManager = local.New()

// SessionManager 接口，管理session
type SessionManager interface {
	// Start 启动连接，默认使用apInfo 中的 shards 作为 shard 数量，如果有闇€瑕佽嚜宸辨寚瀹?shard 数，请修 apInfo 中的信息
	Start(apInfo *dto.WebsocketAP, tokenSource oauth2.TokenSource, intents *dto.Intent) error
}
