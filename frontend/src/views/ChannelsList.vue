<template>
  <div class="bg-base-100">
    <div class="max-w-4xl mx-auto px-4 py-6 pb-20 lg:pb-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-bold">频道</h1>
        <button
          v-if="authStore.isAuthenticated"
          @click="$emit('create-channel')"
          class="btn btn-neutral btn-sm"
        >
          新建频道
        </button>
      </div>

      <div v-if="loading" class="text-center py-12 text-base-content/50">
        <div class="text-sm">加载中...</div>
      </div>

      <div v-else-if="channels.length === 0" class="text-center py-12 text-base-content/40">
        <div class="text-sm">{{ authStore.isAuthenticated ? '暂无频道' : '暂无公开频道' }}</div>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="ch in channels"
          :key="ch.id"
          class="card bg-base-100 border border-base-300 shadow-sm hover:shadow-md transition-shadow hover:border-base-400"
        >
          <div class="card-body p-4">
            <div class="flex items-start gap-2 mb-2">
              <div class="w-8 h-8 bg-neutral text-neutral-content rounded-full flex items-center justify-center text-xs font-bold shrink-0">
                {{ ch.name.charAt(0).toUpperCase() }}
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="font-medium text-base-content mb-1 truncate text-sm">{{ ch.name }}</h3>
                <div class="flex items-center gap-2 text-xs text-base-content/50">
                  <span>{{ ch.owner?.nickname || ch.owner?.username || '未知' }}</span>
                  <span class="shrink-0">{{ formatDate(ch.created_at) }}</span>
                </div>
              </div>
            </div>
            <p class="text-sm text-base-content/70 mb-2 line-clamp-2">{{ ch.description || '暂无描述' }}</p>
            <div v-if="ch.tags" class="flex flex-wrap gap-1 mb-2">
              <span
                v-for="tag in parseTags(ch.tags)"
                :key="tag"
                class="badge badge-ghost badge-xs text-xs"
              >
                {{ tag }}
              </span>
            </div>
            <div v-if="ch.is_public" class="flex items-center gap-1 text-xs text-base-content/50 mb-2">
              <Globe class="w-3 h-3" />
              <span>公开</span>
            </div>
            <div class="flex gap-2">
              <router-link
                v-if="ch.is_member || !authStore.isAuthenticated"
                :to="{ name: 'channel', params: { id: ch.id } }"
                class="btn btn-sm btn-neutral flex-1"
              >
                进入频道
              </router-link>
              <button
                v-else-if="ch.is_pending"
                disabled
                class="btn btn-sm btn-ghost flex-1 opacity-50"
              >
                待审核
              </button>
              <button
                v-else-if="ch.is_invited"
                @click="acceptInvitation(ch.id)"
                class="btn btn-sm btn-primary flex-1"
              >
                接受邀请
              </button>
              <button
                v-else-if="!ch.is_public"
                disabled
                class="btn btn-sm btn-ghost flex-1 opacity-50"
                title="该频道不支持申请"
              >
                不可申请
              </button>
              <button
                v-else
                @click="joinChannel(ch.id)"
                :disabled="ch._joining"
                class="btn btn-sm btn-neutral flex-1"
              >
                {{ ch._joining ? '申请中...' : '申请加入' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import api from '../api/axios';
import { useAuthStore } from '../stores/auth';
import wsClient from '../utils/websocket';
import { Globe } from 'lucide-vue-next';

const emit = defineEmits(['create-channel']);
const authStore = useAuthStore();
const channels = ref([]);
const loading = ref(true);

const loadChannels = async () => {
  loading.value = true;
  try {
    if (authStore.isAuthenticated) {
      const res = await api.get('/channels');
      // 为用户已加入的频道添加is_member标记
      channels.value = (res.data || []).map(ch => ({
        ...ch,
        is_member: true,
        is_pending: false,
        is_invited: false
      }));
    } else {
      const res = await api.get('/public/channels');
      channels.value = res.data || [];
    }
  } catch (error) {
    console.error('Failed to fetch channels:', error);
    channels.value = [];
  } finally {
    loading.value = false;
  }
};

const isChannelMember = (channelId) => {
  const channel = channels.value.find(ch => ch.id === channelId);
  return channel && channel.is_member;
};

const joinChannel = async (channelId) => {
  // 检查频道是否公开
  const channel = channels.value.find(ch => ch.id === channelId);
  if (!channel) {
    alert('频道不存在');
    return;
  }

  if (!channel.is_public) {
    alert('该频道为私有频道，无法直接申请加入');
    return;
  }

  if (!authStore.isAuthenticated) {
    alert('请先登录后再申请加入频道');
    return;
  }

  try {
    if (channel) {
      channel._joining = true;
    }

    // 后端从 URL 参数中获取频道 ID，不需要发送请求体
    const response = await api.post(`/channels/${channelId}/join`);

    if (channel) {
      channel._joining = false;
    }

    if (response.status === 200 || response.status === 201) {
      alert('申请已提交，等待管理员审核');
      // 刷新频道列表，显示已申请状态
      await loadChannels();
    } else {
      alert('申请失败，请稍后重试');
    }
  } catch (error) {
    console.error('Failed to join channel:', error);
    const channel = channels.value.find(ch => ch.id === channelId);
    if (channel) {
      channel._joining = false;
    }
    
    // 显示更具体的错误信息
    if (error.response?.status === 404) {
      alert('频道不存在');
    } else if (error.response?.status === 400) {
      alert(error.response.data?.error || '该频道不支持申请或您已申请过');
    } else if (error.response?.status === 403) {
      alert('无权申请加入该频道');
    } else {
      alert('申请失败：' + (error.response?.data?.error || error.message || '未知错误'));
    }
  }
};

const acceptInvitation = async (channelId) => {
  try {
    // 查找邀请记录
    const res = await api.get('/channels/approvals/pending');
    const invitation = res.data.find(item => item.channel?.id === channelId && item.status === 'invited');
    if (invitation) {
      await api.post(`/channels/approvals/${invitation.id}/accept`);
      alert('已成功加入频道');
      await loadChannels();
    }
  } catch (error) {
    console.error('Failed to accept invitation:', error);
    alert(error.response?.data?.error || '接受邀请失败');
  }
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
};

const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const handleWsMessage = async (message) => {
  if (message.type === 'channel') {
    if (message.action === 'create' || message.action === 'update') {
      await loadChannels();
    } else if (message.action === 'delete') {
      channels.value = channels.value.filter(c => {
        const channelId = String(c.id);
        const dataId = String(message.data.id);
        return channelId !== dataId;
      });
    }
  }
};

onMounted(async () => {
  await loadChannels();

  const setupWebSocketListeners = () => {
    if (localStorage.getItem('userId')) {
      wsClient.on('channel_create', handleWsMessage);
      wsClient.on('channel_update', handleWsMessage);
      wsClient.on('channel_delete', handleWsMessage);
    }
  };

  setupWebSocketListeners();
  wsClient.on('connected', setupWebSocketListeners);
});

onBeforeUnmount(() => {
  wsClient.off('channel_create', handleWsMessage);
  wsClient.off('channel_update', handleWsMessage);
  wsClient.off('channel_delete', handleWsMessage);
  wsClient.off('connected', () => {});
});
</script>