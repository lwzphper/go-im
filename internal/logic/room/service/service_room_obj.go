package service

import (
	"go-im/internal/connect"
	roomType "go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go-im/pkg/util"
)

/**
 * @Description: 房间对象管理
 */

// 新建房间
func (s *Service) newRoom(roomId uint64, name string) *Room {
	s.newRoomLock.Lock()
	defer s.newRoomLock.Unlock()

	if r := s.getRoom(roomId); r != nil {
		return r
	}

	r := &Room{
		RoomId:  roomId,
		name:    name,
		clients: make(map[uint64]*connect.Node),
	}
	s.roomsManager[roomId] = r
	return r
}

// 获取房间
func (s *Service) getRoom(roomId uint64) *Room {
	r, ok := s.roomsManager[roomId]
	if ok {
		return r
	}

	// 本地不存在，从Redis中获取
	if name := s.roomCache.GetName(roomId); name != "" {
		return s.newRoom(roomId, name)
	}

	return nil
}

// 加入房间
func (s *Service) joinRoom(r *Room, n *connect.Node, username string) {
	r.joinLock.Lock()
	defer r.joinLock.Unlock()

	s.roomUserCache.Create(r.RoomId, n.UserId, username)
	s.userServiceCache.Create(r.RoomId, n.UserId, n.ServerAddr)

	if _, ok := r.clients[n.UserId]; !ok {
		r.clients[n.UserId] = n
	}
	n.RoomId = r.RoomId
}

// 推送消息到房间
func (s *Service) pushRoom(r *Room, data *roomType.QueueMsgData) {
	r.pushLock.Lock()
	defer func() {
		r.pushLock.Unlock()
		if err := recover(); err != nil {
			logger.Errorf("Room.Push error: %v", err)
		}
	}()

	for uid, node := range r.clients {
		if data.FromUid == uid {
			continue
		}
		node.DataQueue <- data.MarshalOutput(node.RoomId)
	}
}

// 离开房间
func (s *Service) handleLeaveRoom(r *Room, conn *connect.Node) {
	r.leaveLock.Lock()
	defer r.leaveLock.Unlock()

	s.roomUserCache.Remove(r.RoomId, conn.UserId)
	s.userServiceCache.Remove(r.RoomId, conn.UserId)
	delete(r.clients, conn.UserId)

	conn.RoomId = 0
}

// 获取房间用户列表
func (s *Service) roomUserList(roomId uint64) roomType.UserList {
	userIdNameMap := s.roomUserCache.GetAll(roomId)

	var userData = roomType.UserList{}
	for userId, name := range userIdNameMap {
		uid, _ := util.StringToUint64(userId)
		if uid > 0 {
			userData = append(userData, roomType.UserItem{
				Id:   uid,
				Name: name,
			})
		}
	}
	return userData
}
