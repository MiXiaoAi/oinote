import * as Y from 'yjs'

/**
 * Yjs 协同编辑客户端
 */
export class YjsCollabClient {
  constructor(noteId, userId, username, wsUrl) {
    this.noteId = noteId
    this.userId = userId
    this.username = username
    this.wsUrl = wsUrl || `ws://${window.location.hostname}:3000/ws/collab`
    
    this.doc = new Y.Doc()
    this.ytext = this.doc.getText('content')
    this.ws = null
    this.connected = false
    this.synced = false
    this.clientId = null
    
    // 用户颜色映射
    this.userColors = new Map()
    this.colorPalette = [
      '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
      '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52B788'
    ]
    this.nextColorIndex = 0
    
    // 回调函数
    this.onUpdate = null
    this.onSync = null
    this.onAwareness = null
    this.onUserLeft = null
    this.onConnect = null
    this.onDisconnect = null
    
    // 绑定文档更新事件
    this.doc.on('update', this.handleDocUpdate.bind(this))
  }
  
  /**
   * 获取用户颜色
   */
  getUserColor(userId) {
    if (!this.userColors.has(userId)) {
      const color = this.colorPalette[this.nextColorIndex % this.colorPalette.length]
      this.userColors.set(userId, color)
      this.nextColorIndex++
    }
    return this.userColors.get(userId)
  }

  /**
   * 连接到服务器
   */
  connect() {
    const url = `${this.wsUrl}?userId=${this.userId}&username=${encodeURIComponent(this.username)}&noteId=${this.noteId}`
    
    this.ws = new WebSocket(url)
    
    this.ws.onopen = () => {
      this.connected = true
      
      // 请求同步
      this.requestSync()
      
      if (this.onConnect) {
        this.onConnect()
      }
    }
    
    this.ws.onmessage = (event) => {
      this.handleMessage(event.data)
    }
    
    this.ws.onerror = (error) => {
      console.error('协同编辑 WebSocket 错误:', error)
    }
    
    this.ws.onclose = () => {
      this.connected = false
      this.synced = false
      
      if (this.onDisconnect) {
        this.onDisconnect()
      }
      
      // 尝试重连
      setTimeout(() => {
        if (!this.connected) {
          this.connect()
        }
      }, 3000)
    }
  }

  /**
   * 断开连接
   */
  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.connected = false
    this.synced = false
  }

  /**
   * 请求同步
   */
  requestSync() {
    if (!this.connected) {
      return
    }
    
    // 发送状态向量请求差异更新
    const stateVector = Y.encodeStateVector(this.doc)
    
    this.sendMessage({
      type: 'sync-step2',
      noteId: this.noteId,
      data: {
        stateVector: this.uint8ArrayToBase64(stateVector)
      }
    })
  }

  /**
   * 处理接收到的消息
   */
  handleMessage(data) {
    try {
      const message = JSON.parse(data)
      
      switch (message.type) {
        case 'welcome':
          this.clientId = message.clientId
          break
          
        case 'sync-step1':
          // 服务器发送状态向量
          this.handleSyncStep1(message)
          break
          
        case 'sync-step2':
          // 服务器发送差异更新
          this.handleSyncStep2(message)
          break
          
        case 'update':
          // 服务器广播的更新
          this.handleUpdate(message)
          break
          
        case 'awareness':
          // 其他客户端的感知状态
          this.handleAwareness(message)
          break
          
        case 'user-left':
          // 用户离开
          this.handleUserLeft(message)
          break
          
        default:
          console.warn('未知消息类型:', message.type)
      }
    } catch (error) {
      console.error('处理消息失败:', error)
    }
  }

  /**
   * 处理同步步骤1
   */
  handleSyncStep1(message) {
    // 服务器发送了状态向量，我们需要发送我们的状态向量
    const stateVector = Y.encodeStateVector(this.doc)
    
    this.sendMessage({
      type: 'sync-step2',
      noteId: this.noteId,
      data: {
        stateVector: this.uint8ArrayToBase64(stateVector)
      }
    })
  }

  /**
   * Uint8Array 转 Base64 字符串
   */
  uint8ArrayToBase64(bytes) {
    let binary = ''
    const len = bytes.byteLength
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i])
    }
    return btoa(binary)
  }

  /**
   * Base64 字符串转 Uint8Array
   */
  base64ToUint8Array(base64) {
    const binaryString = atob(base64)
    const len = binaryString.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    return bytes
  }

  /**
   * 处理同步步骤2
   */
  handleSyncStep2(message) {
    if (message.update) {
      try {
        // Go 的 JSON 序列化会将 []byte 编码为 Base64 字符串
        let update
        if (typeof message.update === 'string') {
          // Base64 字符串，需要解码
          update = this.base64ToUint8Array(message.update)
        } else if (Array.isArray(message.update)) {
          // 数字数组
          update = new Uint8Array(message.update)
        } else {
          console.error('未知的 update 格式:', typeof message.update)
          return
        }
        
        if (update.length > 0) {
          // 传递 origin 参数，标记这是远程更新
          Y.applyUpdate(this.doc, update, 'remote')
        }
        
        this.synced = true
        
        if (this.onSync) {
          this.onSync(this.ytext.toString())
        }
      } catch (error) {
        console.error('应用同步更新失败:', error)
      }
    } else {
      // 即使没有更新，也标记为已同步
      this.synced = true
      if (this.onSync) {
        this.onSync(this.ytext.toString())
      }
    }
  }

  /**
   * 处理更新
   */
  handleUpdate(message) {
    if (message.update) {
      try {
        // Go 的 JSON 序列化会将 []byte 编码为 Base64 字符串
        let update
        if (typeof message.update === 'string') {
          // Base64 字符串，需要解码
          update = this.base64ToUint8Array(message.update)
        } else if (Array.isArray(message.update)) {
          // 数字数组
          update = new Uint8Array(message.update)
        } else {
          console.error('未知的 update 格式:', typeof message.update)
          return
        }
        
        if (update.length > 0) {
          // 传递 origin 参数，标记这是远程更新
          Y.applyUpdate(this.doc, update, 'remote')
          
          // 触发 onUpdate 回调，通知编辑器更新
          if (this.onUpdate) {
            this.onUpdate(this.ytext.toString())
          }
        }
      } catch (error) {
        console.error('应用更新失败:', error)
      }
    }
  }

  /**
   * 处理感知状态
   */
  handleAwareness(message) {
    if (this.onAwareness) {
      // 添加颜色信息
      const color = this.getUserColor(message.userId)
      this.onAwareness({
        clientId: message.clientId,
        userId: message.userId,
        username: message.username,
        color: color,
        cursor: message.cursor,
        timestamp: message.timestamp
      })
    }
  }

  /**
   * 处理用户离开
   */
  handleUserLeft(message) {
    if (this.onUserLeft) {
      this.onUserLeft({
        clientId: message.clientId,
        userId: message.userId,
        username: message.username
      })
    }
  }

  /**
   * 处理文档更新
   */
  handleDocUpdate(update, origin) {
    // 如果更新来自远程（origin 为 'remote'），不要发送回服务器
    // 其他情况（包括 origin 为 'local' 或 undefined）都发送到服务器
    if (origin !== 'remote' && this.connected && this.synced) {
      const updateBase64 = this.uint8ArrayToBase64(update)
      
      this.sendMessage({
        type: 'update',
        noteId: this.noteId,
        data: {
          update: updateBase64
        }
      })
    }
    
    // 通知外部更新（仅当是远程更新时，用于更新编辑器）
    if (this.onUpdate && origin === 'remote') {
      this.onUpdate(this.ytext.toString())
    }
  }

  /**
   * 发送消息到服务器
   */
  sendMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  /**
   * 插入文本
   */
  insertText(index, text) {
    this.ytext.insert(index, text)
  }

  /**
   * 删除文本
   */
  deleteText(index, length) {
    this.ytext.delete(index, length)
  }

  /**
   * 获取文本内容
   */
  getText() {
    return this.ytext.toString()
  }

  /**
   * 设置文本内容（用于同步编辑器内容到 Yjs）
   */
  setText(text) {
    const currentText = this.ytext.toString()
    if (currentText !== text) {
      // 使用事务来批量更新
      // 不传递 origin，这样 handleDocUpdate 会将更新发送到服务器
      this.doc.transact(() => {
        this.ytext.delete(0, currentText.length)
        this.ytext.insert(0, text)
      })
    }
  }

  /**
   * 发送光标位置
   */
  sendCursorPosition(from, to) {
    if (!this.connected || !this.synced) {
      return
    }
    
    this.sendMessage({
      type: 'awareness',
      noteId: this.noteId,
      data: {
        cursor: {
          from: from,
          to: to
        }
      }
    })
  }

  /**
   * 销毁客户端
   */
  destroy() {
    this.disconnect()
    this.doc.destroy()
  }
}

export default YjsCollabClient
