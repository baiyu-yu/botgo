// Package local 基于 golang chan 实现的单有manager銆?package local

import (
	"fmt"
	"time"

	"github.com/sealdice/botgo/dto"
	"github.com/sealdice/botgo/log"
	"github.com/sealdice/botgo/sessions/manager"
	"github.com/sealdice/botgo/websocket"
	"golang.org/x/oauth2"
)

// New 创建本地session管理鍣?func New() *ChanManager {
	return &ChanManager{}
}

// ChanManager 默认的本鍦?session manager 实现
type ChanManager struct {
	sessionChan chan dto.Session
}

// Start 启动本地 session manager
func (l *ChanManager) Start(apInfo *dto.WebsocketAP, tokenSource oauth2.TokenSource, intents *dto.Intent) error {
	defer log.Sync()
	if err := manager.CheckSessionLimit(apInfo); err != nil {
		log.Errorf("[ws/session/local] session limited apInfo: %+v", apInfo)
		return err
	}
	startInterval := manager.CalcInterval(apInfo.SessionStartLimit.MaxConcurrency)
	log.Infof("[ws/session/local] will start %d sessions and per session start interval is %s",
		apInfo.Shards, startInterval)

	// 按照shards数量初始化，用于启动连接的管鐞?	l.sessionChan = make(chan dto.Session, apInfo.Shards)
	for i := uint32(0); i < apInfo.Shards; i++ {
		session := dto.Session{
			URL:         apInfo.URL,
			TokenSource: tokenSource,
			Intent:      *intents,
			LastSeq:     0,
			Shards: dto.ShardConfig{
				ShardID:    i,
				ShardCount: apInfo.Shards,
			},
		}
		l.sessionChan <- session
	}

	for session := range l.sessionChan {
		// MaxConcurrency 代表的是姣?5s 可以连多少个请求
		time.Sleep(startInterval)
		go l.newConnect(session)
	}
	return nil
}

// newConnect 启动涓€涓柊鐨勮繛鎺ワ紝濡傛灉杩炴帴鍦ㄧ洃鍚繃绋嬩腑鎶ラ敊浜嗭紝鎴栬€呰杩滅鍏抽棴浜嗛摼鎺ワ紝闇€瑕佽瘑鍒叧闂殑鍘熷洜锛岃兘鍚︾户缁?resume
// 如果能够 resume，则寰€ sessionChan 中放入带有sessionID 的session
// 如果不能，则清理接sessionID，将 session 放入 sessionChan 为// session 的启动，交给 start 中的 for 循环执行，session 不自宸遍€掑綊杩涜閲嶈繛锛岄伩鍏嶉€掑綊娣卞害杩囨繁
func (l *ChanManager) newConnect(session dto.Session) {
	defer func() {
		// panic 留下日志，放鍥?session
		if err := recover(); err != nil {
			websocket.PanicHandler(err, &session)
			l.sessionChan <- session
		}
	}()
	wsClient := websocket.ClientImpl.New(session)
	if err := wsClient.Connect(); err != nil {
		log.Error(err)
		l.sessionChan <- session // 连接失败，丢回去队列排队重连
		return
	}
	var err error
	// 如果 session id 不为空，则执行的是resume 操作，如果为空，则执行的是identify 操作
	if session.ID != "" {
		err = wsClient.Resume()
	} else {
		// 初次鉴权
		err = wsClient.Identify()
	}
	if err != nil {
		log.Errorf("[ws/session] Identify/Resume err %+v", err)
		return
	}
	if err = wsClient.Listening(); err != nil {
		log.Errorf("[ws/session] Listening err %+v", err)
		currentSession := wsClient.Session()
		// 对于不能够进行重连的session，需要清绌?session id 为seq
		if manager.CanNotResume(err) {
			currentSession.ID = ""
			currentSession.LastSeq = 0
		}
		// 涓€浜涢敊璇笉鑳藉閴存潈锛屾瘮濡傛満鍣ㄤ汉琚皝绂侊紝杩欓噷灏辩洿鎺ラ€€鍑轰簡
		if manager.CanNotIdentify(err) {
			msg := fmt.Sprintf("can not identify because server return %+v, so process exit", err)
			log.Errorf(msg)
			panic(msg) // 当机器人被下架，鎴栬€呭皝绂侊紝灏嗕笉鑳藉啀杩炴帴锛屾墍件panic
		}
		// 灏?session 放到 session chan 中，用于启动新的连接，当前连鎺ラ€€鍑?		l.sessionChan <- *currentSession
		return
	}
}
