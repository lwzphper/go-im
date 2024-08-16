package connect

import (
	"github.com/redis/go-redis/v9"
	"go-im/internal/connect/repo"
	"go-im/pkg/logger"
	pkgRedis "go-im/pkg/redis"
	util "go-im/pkg/util"
	"sync"
)

var (
	roomsManager = make(map[uint64]*Room)
	roomLock     sync.Mutex
)

// GetRoom 获取房间
func GetRoom(roomId uint64) *Room {
	if r, ok := roomsManager[roomId]; ok {
		return r
	}

	// 本机不存在，判断其他服务是否已创建房间
	if roomName := repo.RoomCache.GetName(roomId); roomName != "" {
		return newRoom(roomId, roomName)
	}

	return nil
}

// 新建房间
func newRoom(roomId uint64, name string) *Room {
	roomLock.Lock()
	defer roomLock.Unlock()

	if r, ok := roomsManager[roomId]; ok {
		return r
	}

	r := &Room{
		RoomId:   roomId,
		name:     name,
		clients:  make(map[uint64]*Node),
		rdClient: pkgRedis.C(pkgRedis.NAME_DEFAULT),
	}
	roomsManager[roomId] = r
	return r
}

type Room struct {
	RoomId   uint64 // 房间ID
	name     string // 房间名称
	clients  map[uint64]*Node
	lock     sync.RWMutex
	rdClient *redis.Client
}

// Join 加入房间
func (r *Room) Join(conn *Node, username string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	repo.RoomUserCache.Create(r.RoomId, conn.UserId, username)
	repo.UserServiceCache.Create(r.RoomId, conn.UserId, conn.ServerAddr)

	if _, ok := r.clients[conn.UserId]; !ok {
		r.clients[conn.UserId] = conn
	}
	conn.RoomId = r.RoomId
}

// Leave 离开房间
func (r *Room) Leave(conn *Node) {
	r.lock.Lock()
	defer r.lock.Unlock()

	repo.RoomUserCache.Remove(r.RoomId, conn.UserId)
	repo.UserServiceCache.Remove(r.RoomId, conn.UserId)
	delete(r.clients, conn.UserId)

	conn.RoomId = 0
}

// Close 关闭房间
func (r *Room) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, node := range r.clients {
		node.Close()
	}

	repo.UserServiceCache.DeleteRoom(r.RoomId)
	repo.RoomUserCache.DeleteRoom(r.RoomId)
	repo.RoomCache.Remove(r.RoomId)
	delete(roomsManager, r.RoomId)
}

// Push 推送消息到房间
func (r *Room) Push(data *QueueMsgData) {
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
		if err := recover(); err != nil {
			logger.Errorf("Room.Push error: %v", err)
		}
	}()

	for uid, node := range r.clients {
		if data.FromUid == uid {
			continue
		}
		node.DataQueue <- data
	}
}

// 获取房间用户列表
func (r *Room) getUserList() UserList {
	userIdNameMap := repo.RoomUserCache.GetAll(r.RoomId)

	var userData = UserList{}
	for userId, name := range userIdNameMap {
		uid, _ := util.StringToUint64(userId)
		if uid > 0 {
			userData = append(userData, UserItem{
				Id:   uid,
				Name: name,
			})
		}
	}
	return userData
}

// CheckUserIsLogin 判断用户是否已登录
func (r *Room) CheckUserIsLogin(userId uint64) bool {
	return repo.RoomUserCache.Exists(r.RoomId, userId)
}
