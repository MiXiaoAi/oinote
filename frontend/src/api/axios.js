import axios from 'axios';
import { getApiBaseUrl } from '../utils/urlHelper';

const api = axios.create({
    baseURL: getApiBaseUrl(),
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// 添加响应拦截器，检查响应中是否包含错误信息
api.interceptors.response.use(
    (response) => {
        // 检查响应中是否包含错误信息
        if (response.data && response.data.error) {
            // 如果响应中包含错误信息，抛出错误
            const error = new Error(response.data.error);
            error.response = response;
            return Promise.reject(error);
        }
        return response;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default api;
