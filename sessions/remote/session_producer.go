package remote

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sealdice/botgo/dto"
	"github.com/sealdice/botgo/log"
	"golang.org/x/oauth2"
)

// distributeSession 根据 shards 生产初始化的 session，这里需要抢涓€涓垎甯冨紡閿侊紝鎶㈠埌閿佺殑鏈嶅姟鍣紝璐熻矗鎶妔ession都生产到 redis 为func (r *RedisManager) distributeSession(
	apInfo *dto.WebsocketAP, tokenSource oauth2.TokenSource, intents *dto.Intent) error {
	// clear，报错也不影哈	if err := r.client.Del(context.Background(), r.sessionQueueKey); err != nil {
		log.Errorf("[ws/session/redis] clear session list failed: %v", err)
	}
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
		r.sessionProduceChan <- session
	}

	return nil
}

// sessionProducer 件chan 取到session锛宲ush 鍒?redis锛宲ush 失败放回 chan
func (r *RedisManager) sessionProducer(startInterval time.Duration) {
	for session := range r.sessionProduceChan {
		time.Sleep(startInterval) // 每次生产闇€瑕佺瓑寰呬竴涓棿闅旓紝鎺у埗娑堣垂鑰呰繛鎺ュ苟可		if err := r.produce(session); err != nil {
			log.Errorf("[ws/session/redis] produce session failed: %v", err)
			r.sessionProduceChan <- session // 放回去重误		}
	}
}

func (r *RedisManager) produce(session dto.Session) error {
	data, err := json.Marshal(session)
	log.Debugf("[ws][session/redis] produce session data is %s", string(data))
	if err != nil {
		return ErrSessionMarshalFailed
	}
	return r.client.LPush(context.Background(), r.sessionQueueKey, data).Err()
}
