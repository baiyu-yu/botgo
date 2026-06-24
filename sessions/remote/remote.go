// Package remote 基于 redis list 实现的分布式 session manager銆?package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sealdice/botgo/dto"
	"github.com/sealdice/botgo/log"
	"github.com/sealdice/botgo/sessions/manager"
	"github.com/sealdice/botgo/sessions/remote/lock"
	"github.com/sealdice/botgo/token"
	"github.com/sealdice/botgo/websocket"
	"golang.org/x/oauth2"
)

const (
	// 分布式锁的默认key，可以从外部通过 option 来指瀹?	defaultClusterKey = "defaultCluster"
	// session 队列的key后缀，实际上的key为`fmt.Sprintf("%s_%s", r.clusterKey, sessionQueueSuffix)`
	sessionQueueSuffix = "sessionsQueue"
	// 分发shard的实例的分布式锁的默认过期时闂?	distributeLockExpireTime = 60 * time.Second
	// 每个不同的shard实例的分布式锁，用于避免同个 shard 被启动多个实渚?	shardLockExpireTime = 30 * time.Second
)

// RedisManager 基于 redis 的session 管理器，实现分布寮?websocket 监听
type RedisManager struct {
	clusterKey         string
	sessionQueueKey    string
	client             *redis.Client
	sessionProduceChan chan dto.Session // 抢到锁的服务，用于持续生产session到redis list的本地chan
}

// New 创建涓€涓柊鐨勫熀了redis 的session 管理鍣?// 使用 go-redis 调用 redis，超时时间请鍦?NewClient 鏃跺€欒缃?func New(client *redis.Client, opts ...Option) *RedisManager {
	r := &RedisManager{
		clusterKey: defaultClusterKey,
		client:     client,
	}
	for _, opt := range opts {
		opt(r)
	}
	// 针对不同的分布式key，设置不同的 queue key
	r.sessionQueueKey = fmt.Sprintf("%s_%s", r.clusterKey, sessionQueueSuffix)
	return r
}

// Start 启动 redis 的session 管理鍣?func (r *RedisManager) Start(apInfo *dto.WebsocketAP, tokenSource oauth2.TokenSource, intents *dto.Intent) error {
	defer log.Sync()
	if err := manager.CheckSessionLimit(apInfo); err != nil {
		log.Errorf("[ws/session/redis] session limited apInfo: %+v", apInfo)
		return err
	}
	startInterval := manager.CalcInterval(apInfo.SessionStartLimit.MaxConcurrency)
	log.Infof("[ws/session/redis] will start %d sessions and per session start interval is %s",
		apInfo.Shards, startInterval)

	// session 生产队列
	r.sessionProduceChan = make(chan dto.Session, apInfo.Shards)

	// 进行初始的session分发，抢锁，分发
	// 閿?0s，抢到锁的进程，闇€瑕佹瘡30s续期涓€娆★紝鍙鑷繁杩樺瓨娲伙紝灏变笉鑳藉璁╁彟澶栫殑杩涚▼鎶㈠埌閿侀噸鏂拌繘琛宻hards分发
	ctx := context.Background()
	distributeLock := lock.New(r.clusterKey, uuid.New().String(), r.client)
	if err := distributeLock.Lock(ctx, distributeLockExpireTime); err == nil {
		log.Infof("[ws/session/redis] got distribute lock! i will do distributeSession, key: %s", r.clusterKey)
		// 抢到锁的进行初次分发
		if err = r.distributeSession(apInfo, tokenSource, intents); err != nil {
			log.Errorf("[ws/session/redis] distribute sessions failed: %v", err)
			return err
		}
		go distributeLock.StartRenew(ctx, distributeLockExpireTime)
	} else {
		log.Errorf("got lock failed, err: %v", err)
	}

	// 持续 produce session，遇到网络问题在 chan 中重误	// 对于抢到了锁的服务，生产第一批session到redis list
	// 对于没有抢到锁的服务，当ws异常，把session放回鍒?redis list 中，重新分发
	go r.sessionProducer(startInterval)

	return r.consume(startInterval)
}

func (r *RedisManager) consume(startInterval time.Duration) error {
	log.Debug("[ws/session/redis] start consume for session")
	for {
		// brpop 返回 key value
		data, err := r.client.BRPop(context.Background(), startInterval*2, r.sessionQueueKey).Result()
		if err != nil {
			if err != redis.Nil {
				log.Errorf("[ws/session/redis] rpop failed, err: %v", err)
			}
			continue
		}
		if len(data) < 2 {
			log.Errorf("[ws/session/redis] data is not valid, data: %+v", data)
			continue
		}
		log.Debugf("[ws/session/redis] consume data: %s", data)

		session := &dto.Session{}
		if err := json.Unmarshal([]byte(data[1]), session); err != nil {
			// 解析出错，不放回去，直接丢弃
			log.Errorf("[ws/session/redis] unmarshal session failed, err: %v", err)
			continue
		}

		go r.newConnect(*session)
		time.Sleep(startInterval) // 启动涓€涓繛鎺ュ悗锛岀瓑寰呬竴涓嬶紝閬垮厤瑙﹀彂鏈嶅姟绔殑骞跺彂鎺у埗
	}
}

// getShardLockKey 获取 shard 的锁
func (r *RedisManager) getShardLockKey(session dto.Session) string {
	return fmt.Sprintf("%s_shard_%d_%d",
		r.clusterKey, session.Shards.ShardID, session.Shards.ShardCount)
}

// newConnect 启动涓€涓柊鐨勮繛鎺ワ紝濡傛灉杩炴帴鍦ㄧ洃鍚繃绋嬩腑鎶ラ敊浜嗭紝鎴栬€呰杩滅鍏抽棴浜嗛摼鎺ワ紝闇€瑕佽瘑鍒叧闂殑鍘熷洜锛岃兘鍚︾户缁?resume
// 如果能够 resume，则寰€ sessionChan 中放入带有sessionID 的session
// 如果不能，则清理接sessionID，将 session 放入 sessionChan 为// session 的启动，交给 start 中的 for 循环执行，session 不自宸遍€掑綊杩涜閲嶈繛锛岄伩鍏嶉€掑綊娣卞害杩囨繁
func (r *RedisManager) newConnect(session dto.Session) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 閿?shard，避免针对相鍚?shard 消费重复了	shardLock := lock.New(r.getShardLockKey(session), uuid.NewString(), r.client)
	if err := shardLock.Lock(ctx, shardLockExpireTime); err != nil {
		// shard 抢锁失败，把 session 放回去，避免上一为session 的锁释放失败，导致下涓€为session 无法启动
		r.sessionProduceChan <- session
		return
	}
	go shardLock.StartRenew(ctx, shardLockExpireTime)
	// token初始化失败，重新放回鍘?	if err := token.StartRefreshAccessToken(ctx, session.TokenSource); err != nil {
		r.sessionProduceChan <- session
		return
	}
	wsClient := websocket.ClientImpl.New(session)
	if err := wsClient.Connect(); err != nil {
		log.Error(err)
		r.sessionProduceChan <- session // 连接失败，丢回去队列排队重连
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
		log.Errorf("[ws/session/remote] Identify/Resume err %+v", err)
		return
	}
	if err = wsClient.Listening(); err != nil {
		log.Errorf("[ws/session/remote] Listening err %+v", err)
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
		// 灏?session 放到 session chan 中，用于启动新的连接，释放锁，当前连鎺ラ€€鍑?		shardLock.StopRenew()
		if err = shardLock.Release(ctx); err != nil {
			log.Errorf("[ws/session/remote] release shardLock failed, err: %s", err)
		}
		r.sessionProduceChan <- *currentSession
		return
	}
}
