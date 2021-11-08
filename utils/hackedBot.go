package utils

import (
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
)

// HackedBot 拦截一些方法方便测试
type HackedBot struct {
	Bot        **miraiBot.Bot
	testGroups []*client.GroupInfo
}

func (h *HackedBot) valid() bool {
	if h == nil || h.Bot == nil || *h.Bot == nil || !(*h.Bot).Online {
		return false
	}
	return true
}

func (h *HackedBot) FindFriend(uin int64) *client.FriendInfo {
	if !h.valid() {
		return nil
	}
	return (*h.Bot).FindFriend(uin)
}

func (h *HackedBot) FindGroup(code int64) *client.GroupInfo {
	if !h.valid() {
		for _, gi := range h.testGroups {
			if gi.Code == code {
				return gi
			}
		}
		return nil
	}
	return (*h.Bot).FindGroup(code)
}

func (h *HackedBot) SolveFriendRequest(req *client.NewFriendRequest, accept bool) {
	if !h.valid() {
		return
	}
	(*h.Bot).SolveFriendRequest(req, accept)
}

func (h *HackedBot) SolveGroupJoinRequest(i interface{}, accept, block bool, reason string) {
	if !h.valid() {
		return
	}
	(*h.Bot).SolveGroupJoinRequest(i, accept, block, reason)
}

func (h *HackedBot) GetGroupList() []*client.GroupInfo {
	if !h.valid() {
		return h.testGroups
	}
	return (*h.Bot).GroupList
}

func (h *HackedBot) GetFriendList() []*client.FriendInfo {
	if !h.valid() {
		return nil
	}
	return (*h.Bot).FriendList
}

func (h *HackedBot) IsOnline() bool {
	return h.valid()
}

var hackedBot = &HackedBot{Bot: &miraiBot.Instance}

func GetBot() *HackedBot {
	return hackedBot
}

// TESTAddGroup 仅可用于测试
func (h *HackedBot) TESTAddGroup(groupCode int64) {
	for _, g := range h.testGroups {
		if g.Code == groupCode {
			return
		}
	}
	h.testGroups = append(h.testGroups, &client.GroupInfo{
		Uin:  groupCode,
		Code: groupCode,
	})
}

// TESTAddMember 仅可用于测试
func (h *HackedBot) TESTAddMember(groupCode int64, uin int64, permission client.MemberPermission) {
	h.TESTAddGroup(groupCode)
	for _, g := range h.testGroups {
		if g.Code != groupCode {
			continue
		}
		for _, m := range g.Members {
			if m.Uin == uin {
				return
			}
		}
		g.Members = append(g.Members, &client.GroupMemberInfo{
			Group:      g,
			Uin:        uin,
			Permission: permission,
		})
	}
}

// TESTClear 仅可用于测试
func (h *HackedBot) TESTClear() {
	h.testGroups = nil
}
