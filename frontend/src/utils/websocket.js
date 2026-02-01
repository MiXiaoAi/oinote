import { getWebSocketUrl } from './urlHelper';

class WebSocketClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.reconnectInterval = null;
    this.reconnectDelay = 3000; // 3秒后重连
    this.maxReconnectAttempts = 10;
    this.reconnectAttempts = 0;
    this.listeners = new Map();
    this.isConnecting = false;
  }

  connect(userId) {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    if (!userId) {
      return;
    }

    this.isConnecting = true;

    try {
      const wsUrl = `${this.url}?userId=${userId}`;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.emit('connected');
      };

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
        }
      };

      this.ws.onclose = (event) => {
        this.isConnecting = false;
        this.emit('disconnected');
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        this.isConnecting = false;
        this.emit('error', error);
      };
    } catch (error) {
      this.isConnecting = false;
      this.attemptReconnect();
    }
  }

  handleMessage(message) {
    const { type, action, data } = message;

    // 触发特定类型的事件 - 传递完整的消息对象
    const eventKey = `${type}_${action}`;
    this.emit(eventKey, message);

    // 触发通用类型事件 - 传递完整的消息对象
    this.emit(type, message);
  }

  attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return;
    }

    if (this.reconnectInterval) {
      clearTimeout(this.reconnectInterval);
    }

    this.reconnectAttempts++;

    this.reconnectInterval = setTimeout(() => {
      const userId = localStorage.getItem('userId');
      if (userId) {
        this.connect(userId);
      }
    }, this.reconnectDelay);
  }

  disconnect() {
    if (this.reconnectInterval) {
      clearTimeout(this.reconnectInterval);
      this.reconnectInterval = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
    this.isConnecting = false;
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  off(event, callback) {
    if (!this.listeners.has(event)) {
      return;
    }
    const callbacks = this.listeners.get(event);
    const index = callbacks.indexOf(callback);
    if (index > -1) {
      callbacks.splice(index, 1);
    }
  }

  emit(event, data) {
    if (!this.listeners.has(event)) {
      return;
    }
    this.listeners.get(event).forEach(callback => {
      try {
        callback(data);
      } catch (error) {
      }
    });
  }
}

// 创建全局 WebSocket 客户端实例
const wsClient = new WebSocketClient(getWebSocketUrl());

// 自动连接（如果用户已登录）
export function connectWebSocket(userId) {
  if (userId) {
    wsClient.connect(userId);
  }
}

// 断开连接
export function disconnectWebSocket() {
  wsClient.disconnect();
}

export default wsClient;