package types

type MsgHandler func(userId uint64, data *Input)

type MsgStrategy map[MsgMethod]MsgHandler

// 注册
func (m MsgStrategy) Register(method MsgMethod, handler MsgHandler) {
	/*if _, ok := m[method]; ok {
		return errors.New("msg method already exists")
	}*/
	m[method] = handler
}

// 获取
func (m MsgStrategy) Get(method MsgMethod) MsgHandler {
	if h, ok := m[method]; ok {
		return h
	}

	return nil
}
