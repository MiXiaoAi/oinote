<template>
  <div class="min-h-screen bg-base-100">
    <!-- Header -->
    <div class="bg-base-200/50 border-b border-base-300">
      <div class="max-w-4xl mx-auto px-4 py-4">
        <h1 class="text-2xl font-bold text-base-content">消息通知</h1>
      </div>
    </div>

    <!-- Main Content -->
    <div class="max-w-4xl mx-auto px-4 py-6">
      <div v-if="loading" class="flex flex-col items-center justify-center h-64 text-base-content/40">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-neutral mb-4"></div>
        <div>加载中...</div>
      </div>

      <div v-else-if="approvals.length === 0" class="text-center py-16 text-base-content/40">
        <div class="text-lg mb-2">暂无消息通知</div>
        <div class="text-sm">当有新的申请或通知时，会显示在这里</div>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="approval in approvals"
          :key="approval.id"
          class="card bg-base-100 border border-base-300 shadow-sm"
        >
          <div class="card-body p-4">
            <div class="flex items-start gap-4">
              <!-- User Avatar -->
              <div class="avatar">
                <div class="w-12 h-12 bg-neutral text-neutral-content rounded-full flex items-center justify-center text-lg font-bold">
                  {{ approval.status === 'invited'
                    ? (approval.channel?.owner?.username?.charAt(0) || '?')
                    : (approval.user?.username?.charAt(0) || '?') }}
                </div>
              </div>

              <!-- User Info -->
                          <div class="flex-1">
                            <div class="flex items-center gap-2 mb-1">
                              <h3 class="font-bold text-base-content">
                                {{ approval.status === 'invited' ? (approval.channel?.owner?.nickname || approval.channel?.owner?.username) : (approval.user?.nickname || approval.user?.username) }}
                              </h3>
                              <span class="text-sm text-base-content/60">
                                @{{ approval.status === 'invited' ? approval.channel?.owner?.username : approval.user?.username }}
                              </span>
                            </div>
                            <p class="text-sm text-base-content/70">
                              {{ approval.status === 'invited' ? `邀请您加入频道: ${approval.channel?.name}` : `申请加入频道: ${approval.channel?.name}` }}
                            </p>
                            <p class="text-xs text-base-content/50 mt-1">
                              {{ approval.status === 'invited' ? '邀请时间' : '申请时间' }}: {{ formatDate(approval.joined_at) }}
                            </p>
                          </div>
              
                          <!-- Actions -->
                          <div class="flex gap-2">
                            <button
                              @click="approval.status === 'invited' ? acceptInvitation(approval.id) : approveRequest(approval.id)"
                              :disabled="approval._processing"
                              class="btn btn-sm btn-success text-white"
                            >
                              {{ approval._processing ? '处理中...' : (approval.status === 'invited' ? '接受' : '批准') }}
                            </button>
                            <button
                              @click="approval.status === 'invited' ? rejectInvitation(approval.id) : rejectRequest(approval.id)"
                              :disabled="approval._processing"
                              class="btn btn-sm btn-error text-white"
                            >
                              {{ approval._processing ? '处理中...' : '拒绝' }}
                            </button>
                          </div>            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import api from '../api/axios';

const router = useRouter();
const authStore = useAuthStore();
const approvals = ref([]);
const loading = ref(true);

const loadApprovals = async () => {
  // 检查是否已登录
  if (!authStore.isAuthenticated) {
    router.push('/login');
    return;
  }

  try {
    const res = await api.get('/channels/approvals/pending');
    approvals.value = res.data || [];
  } catch (error) {
    // 如果返回401，可能是token过期，重定向到登录页
    if (error.response?.status === 401) {
      authStore.logout();
      router.push('/login');
    }
    console.error('Failed to load approvals:', error);
    approvals.value = [];
  } finally {
    loading.value = false;
  }
};

const approveRequest = async (approvalId) => {
  try {
    const approval = approvals.value.find(a => a.id === approvalId);
    if (approval) {
      approval._processing = true;
    }

    await api.post('/channels/approvals/approve', {
      member_record_id: approvalId,
      action: 'approve_request'
    });

    // 重新加载列表
    await loadApprovals();
  } catch (error) {
    console.error('Failed to approve request:', error);
    const approval = approvals.value.find(a => a.id === approvalId);
    if (approval) {
      approval._processing = false;
    }
    alert(error.response?.data?.error || '批准失败');
  }
};

const acceptInvitation = async (invitationId) => {
  try {
    const approval = approvals.value.find(a => a.id === invitationId);
    if (approval) {
      approval._processing = true;
    }

    await api.post(`/channels/approvals/${invitationId}/accept`);

    // 重新加载列表
    await loadApprovals();
  } catch (error) {
    console.error('Failed to accept invitation:', error);
    const approval = approvals.value.find(a => a.id === invitationId);
    if (approval) {
      approval._processing = false;
    }
    alert(error.response?.data?.error || '接受邀请失败');
  }
};

const rejectRequest = async (approvalId) => {
  if (!confirm('确定要拒绝此申请吗？')) return;

  try {
    const approval = approvals.value.find(a => a.id === approvalId);
    if (approval) {
      approval._processing = true;
    }

    await api.delete(`/channels/approvals/${approvalId}`);

    // 重新加载列表
    await loadApprovals();
  } catch (error) {
    console.error('Failed to reject request:', error);
    const approval = approvals.value.find(a => a.id === approvalId);
    if (approval) {
      approval._processing = false;
    }
    alert(error.response?.data?.error || '拒绝失败');
  }
};

const rejectInvitation = async (invitationId) => {
  if (!confirm('确定要拒绝此邀请吗？')) return;

  try {
    const approval = approvals.value.find(a => a.id === invitationId);
    if (approval) {
      approval._processing = true;
    }

    await api.delete(`/channels/approvals/${invitationId}`);

    // 重新加载列表
    await loadApprovals();
  } catch (error) {
    console.error('Failed to reject invitation:', error);
    const approval = approvals.value.find(a => a.id === invitationId);
    if (approval) {
      approval._processing = false;
    }
    alert(error.response?.data?.error || '拒绝邀请失败');
  }
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

onMounted(() => {
  loadApprovals();
});
</script>