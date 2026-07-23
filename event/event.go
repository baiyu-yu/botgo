// Package event 事件处理注册
package event

import (
	"encoding/json"
	"sync"

	"github.com/tidwall/gjson" // 由于回包的d 类型不确定，gjson 用于从回包json中提可d 并进行针瀵规€х殑瑙ｆ瀽

	"github.com/sealdice/botgo/dto"
)

var eventParseFuncMapLock = new(sync.RWMutex)
var eventParseFuncMap = map[dto.OPCode]map[dto.EventType]eventParseFunc{
	dto.WSDispatchEvent: {
		dto.EventGuildCreate: guildHandler,
		dto.EventGuildUpdate: guildHandler,
		dto.EventGuildDelete: guildHandler,

		dto.EventChannelCreate: channelHandler,
		dto.EventChannelUpdate: channelHandler,
		dto.EventChannelDelete: channelHandler,

		dto.EventGuildMemberAdd:    guildMemberHandler,
		dto.EventGuildMemberUpdate: guildMemberHandler,
		dto.EventGuildMemberRemove: guildMemberHandler,

		dto.EventMessageCreate: messageHandler,
		dto.EventMessageDelete: messageDeleteHandler,

		dto.EventMessageReactionAdd:    messageReactionHandler,
		dto.EventMessageReactionRemove: messageReactionHandler,

		dto.EventAtMessageCreate:     atMessageHandler,
		dto.EventPublicMessageDelete: publicMessageDeleteHandler,

		dto.EventDirectMessageCreate: directMessageHandler,
		dto.EventDirectMessageDelete: directMessageDeleteHandler,

		dto.EventAudioStart:  audioHandler,
		dto.EventAudioFinish: audioHandler,
		dto.EventAudioOnMic:  audioHandler,
		dto.EventAudioOffMic: audioHandler,

		dto.EventMessageAuditPass:   messageAuditHandler,
		dto.EventMessageAuditReject: messageAuditHandler,

		dto.EventForumThreadCreate: threadHandler,
		dto.EventForumThreadUpdate: threadHandler,
		dto.EventForumThreadDelete: threadHandler,
		dto.EventForumPostCreate:   postHandler,
		dto.EventForumPostDelete:   postHandler,
		dto.EventForumReplyCreate:  replyHandler,
		dto.EventForumReplyDelete:  replyHandler,
		dto.EventForumAuditResult:  forumAuditHandler,

		dto.EventInteractionCreate:    interactionHandler,
		dto.EventGroupAtMessageCreate: groupAtMessageHandler,
		dto.EventGroupMessageCreate:   groupMessageHandler,
		dto.EventC2CMessageCreate:     c2cMessageHandler,
		dto.EventSubscribeMsgStatus:   subscribeStatusHandler,
		dto.EventC2CFriendAdd:         c2cFriendAddHandler,
		dto.EventC2CFriendDel:         c2cFriendDelHandler,
		dto.EventGroupAddRobot:        groupAddRobotHandler,
		dto.EventGroupDelRobot:        groupDelRobotHandler,
		dto.EventGroupMemberAdd:       groupMemberAddHandler,
		dto.EventGroupMemberRemove:    groupMemberRemoveHandler,
		dto.EventEnterAIO:             enterAIOHandler,
	},
}

// RegisterHandler 注册回调事件处理器
func RegisterHandler(opCode dto.OPCode, eventType dto.EventType, handler eventParseFunc) {
	eventParseFuncMapLock.Lock()
	defer eventParseFuncMapLock.Unlock()
	if eventParseFuncMap[opCode] == nil {
		eventParseFuncMap[opCode] = make(map[dto.EventType]eventParseFunc)
	}
	eventParseFuncMap[opCode][eventType] = handler
}

func getHandler(opCode dto.OPCode, eventType dto.EventType) (eventParseFunc, bool) {
	eventParseFuncMapLock.RLock()
	defer eventParseFuncMapLock.RUnlock()
	f, ok := eventParseFuncMap[opCode][eventType]
	return f, ok
}

type eventParseFunc func(event *dto.WSPayload, message []byte) error

// ParseAndHandle 处理回调事件
func ParseAndHandle(payload *dto.WSPayload) error {
	// 指定类型的handler
	if h, ok := getHandler(payload.OPCode, payload.Type); ok {
		return h(payload, payload.RawMessage)
	}
	// 透传handler
	hStruct := getHandlers(payload)
	if hStruct.Plain != nil {
		return hStruct.Plain(payload, payload.RawMessage)
	}
	return nil
}

func getHandlers(payload *dto.WSPayload) *Handlers {
	if payload != nil && payload.Session != nil && payload.Session.AppID != "" {
		handlersMu.RLock()
		h, ok := DefaultHandlersMap[payload.Session.AppID]
		handlersMu.RUnlock()
		if ok {
			return h
		}
	}
	return &DefaultHandlers
}

// ParseData 解析数据
func ParseData(message []byte, target interface{}) error {
	data := gjson.Get(string(message), "d")
	return json.Unmarshal([]byte(data.String()), target)
}

func guildHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGuildData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Guild != nil {
		return h.Guild(payload, data)
	}
	return nil
}

func channelHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSChannelData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Channel != nil {
		return h.Channel(payload, data)
	}
	return nil
}

func guildMemberHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGuildMemberData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GuildMember != nil {
		return h.GuildMember(payload, data)
	}
	return nil
}

func messageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Message != nil {
		return h.Message(payload, data)
	}
	return nil
}

func messageDeleteHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSMessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.MessageDelete != nil {
		return h.MessageDelete(payload, data)
	}
	return nil
}

func messageReactionHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSMessageReactionData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.MessageReaction != nil {
		return h.MessageReaction(payload, data)
	}
	return nil
}

func atMessageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.ATMessage != nil {
		return h.ATMessage(payload, data)
	}
	return nil
}

func groupAtMessageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupATMessage != nil {
		return h.GroupATMessage(payload, data)
	}
	return nil
}

func groupMessageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupMessage != nil {
		return h.GroupMessage(payload, data)
	}
	return nil
}

func c2cMessageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSC2CMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.C2CMessage != nil {
		return h.C2CMessage(payload, data)
	}
	return nil
}

func subscribeStatusHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSSubscribeMsgStatus{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.SubscribeMsgStatus != nil {
		return h.SubscribeMsgStatus(payload, data)
	}
	return nil
}

func c2cFriendDelHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSC2CFriendData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.C2CFriend != nil {
		return h.C2CFriend(payload, data)
	}
	return nil
}

func c2cFriendAddHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSC2CFriendData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.C2CFriend != nil {
		return h.C2CFriend(payload, data)
	}
	return nil
}

func publicMessageDeleteHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSPublicMessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.PublicMessageDelete != nil {
		return h.PublicMessageDelete(payload, data)
	}
	return nil
}

func directMessageHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSDirectMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.DirectMessage != nil {
		return h.DirectMessage(payload, data)
	}
	return nil
}

func directMessageDeleteHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSDirectMessageDeleteData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.DirectMessageDelete != nil {
		return h.DirectMessageDelete(payload, data)
	}
	return nil
}

func audioHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSAudioData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Audio != nil {
		return h.Audio(payload, data)
	}
	return nil
}

func threadHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSThreadData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Thread != nil {
		return h.Thread(payload, data)
	}
	return nil
}

func postHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSPostData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Post != nil {
		return h.Post(payload, data)
	}
	return nil
}

func replyHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSReplyData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Reply != nil {
		return h.Reply(payload, data)
	}
	return nil
}

func forumAuditHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSForumAuditData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.ForumAudit != nil {
		return h.ForumAudit(payload, data)
	}
	return nil
}

func messageAuditHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSMessageAuditData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.MessageAudit != nil {
		return h.MessageAudit(payload, data)
	}
	return nil
}

func interactionHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSInteractionData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.Interaction != nil {
		return h.Interaction(payload, data)
	}
	return nil
}

func enterAIOHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSEnterAIOData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.EnterAIO != nil {
		return h.EnterAIO(payload, data)
	}
	return nil
}

func groupAddRobotHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupRobotEventData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupAddRobot != nil {
		return h.GroupAddRobot(payload, data)
	}
	return nil
}

func groupDelRobotHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupRobotEventData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupDelRobot != nil {
		return h.GroupDelRobot(payload, data)
	}
	return nil
}

func groupMemberAddHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupMemberAddData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupMemberAdd != nil {
		return h.GroupMemberAdd(payload, data)
	}
	return nil
}

func groupMemberRemoveHandler(payload *dto.WSPayload, message []byte) error {
	data := &dto.WSGroupMemberRemoveData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	h := getHandlers(payload)
	if h.GroupMemberRemove != nil {
		return h.GroupMemberRemove(payload, data)
	}
	return nil
}
