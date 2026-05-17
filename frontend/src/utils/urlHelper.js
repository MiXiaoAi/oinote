// URL生成工具函数
export const getApiBaseUrl = () => {
  // 如果在开发环境，根据当前域名确定API地址
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:3000/api';
  }
  // 生产环境或其他情况，使用当前域名的3000端口
  return `${window.location.protocol}//${window.location.hostname}:3000/api`;
};

export const getWebSocketUrl = () => {
  // 如果在开发环境，使用 ws://localhost:3000/ws
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'ws://localhost:3000/ws';
  }
  // 生产环境或其他情况，使用当前域名的3000端口和ws协议
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.hostname}:3000/ws`;
};

export const getFileUrl = (filePath) => {
  if (!filePath) return '';
  // 如果已经是完整URL，直接返回
  if (filePath.startsWith('http://') || filePath.startsWith('https://')) {
    return filePath;
  }
  
  const apiBase = getApiBaseUrl();
  const baseUrl = apiBase.replace('/api', '');
  
  // 检查是否为媒体文件（视频或音频）
  const mediaExtensions = /\.(mp4|webm|ogg|mov|mp3|wav|m4a|flac)$/i;
  if (mediaExtensions.test(filePath)) {
    // 使用支持Range请求的/media路径
    return baseUrl + '/media/' + filePath.replace(/^\/+/, '');
  }
  
  // 其他文件使用原来的/uploads路径
  return baseUrl + filePath;
};
