<template>
  <div class="w-64 bg-base-200 flex flex-col h-full border-r border-base-300">
    <!-- Header -->
    <div class="h-12 border-b border-base-300 flex items-center px-4 font-bold cursor-pointer hover:bg-base-300 transition-colors shrink-0" 
         @click="router.push('/')">
      <FileText class="w-5 h-5 mr-2" />
      <span class="text-lg tracking-tight">oinote</span>
      <router-link v-if="authStore.isAuthenticated && authStore.user?.role === 'admin'" 
              to="/admin" 
              class="ml-auto btn btn-ghost btn-xs btn-square" 
              title="后台管理"
              @click.stop>
        <Settings class="w-4 h-4" />
      </router-link>
    </div>

    <!-- Navigation -->
    <div class="flex-1 overflow-y-auto custom-scrollbar px-2 space-y-4">
      <!-- Notes Section - 仅登录用户可见 -->
      <div v-if="authStore.isAuthenticated">
        <div 
          class="px-3 py-2 text-xs font-medium text-base-content/50 uppercase tracking-wider flex items-center justify-between leading-none hover:bg-base-300/80 transition-colors rounded -mx-2 px-2 cursor-pointer"
          @click="router.push('/notes')"
        >
          <div class="flex items-center gap-1 hover:text-base-content/80 transition-colors group flex-1">
            <FileText class="w-3 h-3" />
            <span>笔记</span>
          </div>
          <button
            class="btn btn-ghost btn-xs btn-square"
            @click.stop="$emit('create-note')"
          >
            <Plus class="w-4 h-4" />
          </button>
        </div>
        <div class="max-h-60 overflow-y-auto">
          <ul class="menu p-0 w-full space-y-1">
            <li
              v-for="note in notes"
              :key="note.id"
              @contextmenu.prevent="handleContextMenu($event, note, 'note')"
              :class="['border rounded-md transition-all', isNoteActive(note.id) ? 'border-base-400 bg-base-100' : 'border-transparent hover:border-base-300/50']"
            >
              <router-link
                :to="{ name: 'note-editor', params: { id: note.id } }"
                :class="['group flex items-center gap-2 py-2.5', isNoteActive(note.id) ? 'text-base-content' : 'text-base-content/70']"
              >
                <FileText class="w-4 h-4 shrink-0" />
                <span class="truncate flex-1 font-medium">{{ note.title || '无标题' }}</span>
              </router-link>
            </li>
          </ul>
        </div>
      </div>

      <!-- Channels Section - 仅登录用户可见 -->
      <div v-if="authStore.isAuthenticated">
        <div 
          class="px-3 py-2 text-xs font-medium text-base-content/50 uppercase tracking-wider flex items-center justify-between leading-none hover:bg-base-300/80 transition-colors rounded -mx-2 px-2 cursor-pointer"
          @click="router.push('/channels')"
        >
          <div class="flex items-center gap-1 hover:text-base-content/80 transition-colors group flex-1">
            <Hash class="w-3 h-3" />
            <span>频道</span>
          </div>
          <button
            class="btn btn-ghost btn-xs btn-square"
            @click.stop="$emit('create-channel')"
          >
            <Plus class="w-4 h-4" />
          </button>
        </div>
        <div class="max-h-60 overflow-y-auto">
          <ul class="menu p-0 w-full space-y-1">
            <li
              v-for="channel in channels"
              :key="channel.id"
              @contextmenu.prevent="handleContextMenu($event, channel, 'channel')"
              :class="['border rounded-md transition-all', isChannelActive(channel.id) ? 'border-base-400 bg-base-100' : 'border-transparent hover:border-base-300/50']"
            >
              <a
                @click="navigateToChannel(channel.id)"
                :class="['group flex items-center gap-2 py-2.5 cursor-pointer', isChannelActive(channel.id) ? 'text-base-content' : 'text-base-content/70']"
              >
                <Hash class="w-4 h-4 shrink-0" />
                <span class="truncate flex-1 font-medium">{{ channel.name }}</span>
              </a>
            </li>
          </ul>
        </div>
      </div>

      <!-- 访客内容 -->
      <div v-if="!authStore.isAuthenticated" class="space-y-4">
        <!-- 欢迎信息 -->
        <div class="px-3 py-2 text-xs font-medium text-base-content/50 uppercase tracking-wider">
          <span class="hover:bg-base-300/30 hover:text-base-content/80 cursor-pointer transition-colors rounded px-1 -mx-1 py-1 -my-1">欢迎使用</span>
        </div>
        <div class="px-3 py-4 bg-base-300/30 rounded-lg">
          <div class="text-xs text-base-content/60 mb-3">
            简洁高效的笔记与协作平台
          </div>
          <div class="space-y-2">
            <div class="flex items-center gap-2 text-xs text-base-content/70">
              <FileText class="w-3 h-3" />
              <span>创建和管理笔记</span>
            </div>
            <div class="flex items-center gap-2 text-xs text-base-content/70">
              <Hash class="w-3 h-3" />
              <span>实时协作频道</span>
            </div>
            <div class="flex items-center gap-2 text-xs text-base-content/70">
              <Users class="w-3 h-3" />
              <span>团队协作</span>
            </div>
          </div>
        </div>

        <!-- 快速链接 -->
        <div class="px-3 py-2 text-xs font-medium text-base-content/50 uppercase tracking-wider">
          <span class="hover:bg-base-300/30 hover:text-base-content/80 cursor-pointer transition-colors rounded px-1 -mx-1 py-1 -my-1">快速开始</span>
        </div>
        <div class="space-y-1">
          <router-link
            to="/login"
            class="block px-3 py-2 text-sm hover:bg-base-300 rounded-lg transition-colors"
          >
            <div class="flex items-center gap-2">
              <User class="w-4 h-4 text-base-content/70" />
              <span>登录账户</span>
            </div>
          </router-link>
          <button
            @click="showAbout = true"
            class="w-full text-left px-3 py-2 text-sm hover:bg-base-300 rounded-lg transition-colors"
          >
            <div class="flex items-center gap-2">
              <Info class="w-4 h-4 text-base-content/70" />
              <span>关于应用</span>
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <div
      v-if="showContextMenu"
      class="fixed z-50 bg-base-100 shadow-xl rounded-lg border border-base-300 py-1 min-w-[160px] text-sm"
      :style="{ top: `${contextMenuY}px`, left: `${contextMenuX}px` }"
    >
      <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="openManageModal">
        <Edit class="w-4 h-4" /> 管理
      </button>
      <button class="w-full text-left px-4 py-2 hover:bg-error/10 flex items-center gap-2 text-error hover:text-error/80" @click="handleDelete">
        <Trash2 class="w-4 h-4" /> 删除
      </button>
    </div>

    <!-- Footer -->
    <div class="shrink-0 bg-base-300/30 p-2 space-y-2">
      <!-- Tools -->
      <div class="flex items-center justify-between px-2 py-1">
        <span class="text-xs font-medium text-base-content/70">设置与外观</span>
        <div class="flex items-center gap-1">
          <router-link to="/approvals" class="btn btn-ghost btn-xs btn-circle relative" title="消息通知">
            <Bell class="w-4 h-4" />
            <span v-if="hasUnreadNotifications" class="absolute top-0 right-0 w-2 h-2 bg-error rounded-full"></span>
          </router-link>
          <button class="btn btn-ghost btn-xs btn-circle" title="设置" @click="openSettings">
            <Settings class="w-4 h-4" />
          </button>
          <button class="btn btn-ghost btn-xs btn-circle" @click="toggleTheme" title="切换主题">
            <Sun v-if="themeStore.theme === 'light'" class="w-4 h-4" />
            <Moon v-else class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- User Info -->
      <div class="flex items-center p-2 rounded-lg bg-base-300/50 space-x-3">
        <div class="avatar">
          <div class="rounded-full w-9 h-9 shadow-sm overflow-hidden bg-neutral text-neutral-content flex items-center justify-center">
            <img v-if="authStore.user?.avatar" :src="getFileUrl(authStore.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
            <span v-else class="text-sm font-bold">{{ avatarChar }}</span>
          </div>
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-bold truncate">{{ authStore.user?.nickname || authStore.user?.username || '访客' }}</div>
          <div class="text-[10px] text-base-content/50 truncate">{{ authStore.isAuthenticated ? '在线' : '离线' }}</div>
        </div>
        <div class="flex items-center">
          <button v-if="authStore.isAuthenticated" @click="handleLogout" 
                  class="btn btn-ghost btn-xs btn-square tooltip tooltip-top" data-tip="退出">
            <LogOut class="w-4 h-4" />
          </button>
          <router-link v-else to="/login" class="btn btn-neutral btn-xs">登录</router-link>
        </div>
      </div>
    </div>

    <dialog :open="showSettings" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box">
        <h3 class="font-bold text-lg">个人设置</h3>
        <div class="py-4 space-y-3">
          <div>
            <label class="label">
              <span class="label-text">昵称</span>
            </label>
            <input v-model="settingsNickname" type="text" class="input input-bordered w-full" placeholder="输入昵称" />
          </div>
          <div>
            <label class="label">
              <span class="label-text">头像</span>
            </label>
            <div class="flex items-center gap-4">
              <div class="avatar">
                <div class="w-20 h-20 rounded-full bg-neutral text-neutral-content flex items-center justify-center overflow-hidden border-2 border-base-300">
                  <img
                    v-if="avatarPreviewUrl || authStore.user?.avatar"
                    :src="avatarPreviewUrl || getFileUrl(authStore.user?.avatar)"
                    alt="头像预览"
                    class="w-full h-full object-cover"
                  />
                  <span v-else class="text-2xl font-bold">{{ avatarChar }}</span>
                </div>
              </div>
              <div class="flex-1">
                <input
                  ref="avatarInputRef"
                  type="file"
                  accept="image/*"
                  class="file-input file-input-bordered w-full"
                  @change="handleAvatarSelected"
                  :disabled="isUploadingAvatar"
                />
                <div v-if="isUploadingAvatar" class="mt-2">
                  <div class="flex items-center justify-between text-xs mb-1">
                    <span>上传中...</span>
                    <span>{{ avatarUploadProgress }}%</span>
                  </div>
                  <progress class="progress progress-sm w-full" :value="avatarUploadProgress" max="100"></progress>
                </div>
              </div>
            </div>
          </div>
          <div>
            <label class="label">
              <span class="label-text">简介</span>
            </label>
            <textarea v-model="settingsBio" class="textarea textarea-bordered w-full min-h-[96px]" placeholder="写点简介..."></textarea>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeSettings" :disabled="savingSettings">取消</button>
          <button class="btn btn-neutral" @click="saveSettings" :disabled="savingSettings">
            {{ savingSettings ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </dialog>

    <ManageModal
      :open="showManageModal"
      :type="editType"
      :item="editingItem"
      @close="closeManageModal"
      @save="saveManage"
    />

    <!-- About Dialog -->
    <dialog :open="showAbout" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box">
        <h3 class="font-bold text-lg flex items-center gap-2">
          <FileText class="w-5 h-5" />
          关于 oinote
        </h3>
        <div class="py-4 space-y-4">
          <div>
            <h4 class="font-medium mb-2">应用介绍</h4>
            <p class="text-sm text-base-content/70">
              oinote 是一个简洁高效的笔记与协作平台，旨在为个人和团队提供便捷的知识管理和实时协作体验。
            </p>
          </div>
          
          <div>
            <h4 class="font-medium mb-2">主要功能</h4>
            <ul class="text-sm text-base-content/70 space-y-1">
              <li class="flex items-center gap-2">
                <FileText class="w-3 h-3" />
                <span>创建和管理个人笔记</span>
              </li>
              <li class="flex items-center gap-2">
                <Hash class="w-3 h-3" />
                <span>实时协作频道交流</span>
              </li>
              <li class="flex items-center gap-2">
                <Users class="w-3 h-3" />
                <span>团队协作与知识共享</span>
              </li>
              <li class="flex items-center gap-2">
                <Settings class="w-3 h-3" />
                <span>个性化设置与主题切换</span>
              </li>
            </ul>
          </div>
          
          <div>
            <h4 class="font-medium mb-2">技术栈</h4>
            <div class="text-sm text-base-content/70">
              <span class="badge badge-ghost mr-1">Vue 3</span>
              <span class="badge badge-ghost mr-1">Go</span>
              <span class="badge badge-ghost mr-1">Fiber</span>
              <span class="badge badge-ghost mr-1">SQLite</span>
              <span class="badge badge-ghost">DaisyUI</span>
            </div>
          </div>
          
          <div class="text-xs text-base-content/50 pt-2 border-t border-base-300">
            <p>版本 1.1.0 | © 2026 MiXiaoAi</p>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="showAbout = false">关闭</button>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount, inject } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useThemeStore } from '../stores/theme';
import api from '../api/axios';
import eventBus from '../utils/eventBus';
import ManageModal from './ManageModal.vue';
import { getFileUrl } from '../utils/urlHelper';
import {
  Hash, FileText, Sun, Moon, LogOut, Settings,
  Edit, Trash2, Users, Info, User, Plus, Bell
} from 'lucide-vue-next';

const props = defineProps({
  channels: {
    type: Array,
    default: () => []
  },
  notes: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['create-channel', 'create-note', 'logout', 'delete-note', 'rename-note', 'delete-channel', 'rename-channel']);

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const themeStore = useThemeStore();

// 判断笔记是否激活
const isNoteActive = (noteId) => {
  return route.name === 'note-editor' && route.params.id == noteId;
};

// 判断频道是否激活
const isChannelActive = (channelId) => {
  return route.name === 'channel' && route.params.id == channelId;
};

// Context Menu State
const showContextMenu = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuItem = ref(null);
const contextMenuType = ref(''); // 'note' or 'channel'

const avatarChar = computed(() => {
  const name = authStore.user?.username || '访客';
  return name.charAt(0).toUpperCase();
});

const showSettings = ref(false);
const savingSettings = ref(false);
const avatarUploadProgress = ref(0);
const isUploadingAvatar = ref(false);
const settingsNickname = ref('');
const settingsBio = ref('');
const avatarInputRef = ref(null);
const selectedAvatarFile = ref(null);
const avatarPreviewUrl = ref(null);

const notification = inject('notification');

const showManageModal = ref(false);
const editType = ref('');
const editingItem = ref(null);
const showAbout = ref(false);

// 未读消息红点
const hasUnreadNotifications = ref(false);

// 检查是否有未读消息
const checkUnreadNotifications = async () => {
  if (!authStore.isAuthenticated) {
    hasUnreadNotifications.value = false;
    return;
  }
  try {
    const res = await api.get('/channels/approvals/pending');
    hasUnreadNotifications.value = (res.data || []).length > 0;
  } catch (err) {
    hasUnreadNotifications.value = false;
  }
};

const openManageModal = () => {
  if (!contextMenuItem.value) return;
  editingItem.value = contextMenuItem.value;
  editType.value = contextMenuType.value;
  showManageModal.value = true;
  closeContextMenu();
};

const closeManageModal = () => {
  showManageModal.value = false;
  editingItem.value = null;
};

const saveManage = async (payload) => {
  if (!editingItem.value) return;
  try {
    if (editType.value === 'note') {
      const res = await api.put(`/notes/${editingItem.value.id}`, payload);
      eventBus.emit('note-updated', {
        id: editingItem.value.id,
        title: payload.title,
        is_public: payload.is_public,
        tags: payload.tags,
      });
      if (notification) notification.showNotification('保存成功', 'success');
      editingItem.value = res.data;
    } else if (editType.value === 'channel') {
      emit('rename-channel', {
        ...editingItem.value,
        name: payload.title,
        description: payload.description,
        is_public: payload.is_public,
        tags: payload.tags
      });
      if (notification) notification.showNotification('保存成功', 'success');
    }
    closeManageModal();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '保存失败', 'error');
  }
};

const openSettings = async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login');
    return;
  }

  try {
    await authStore.refreshMe();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '加载个人信息失败', 'error');
  }

  settingsNickname.value = authStore.user?.nickname || '';
  settingsBio.value = authStore.user?.bio || '';
  selectedAvatarFile.value = null;
  avatarPreviewUrl.value = null;
  if (avatarInputRef.value) avatarInputRef.value.value = '';
  showSettings.value = true;
};

const closeSettings = () => {
  showSettings.value = false;
  // 清理预览 URL
  if (avatarPreviewUrl.value) {
    URL.revokeObjectURL(avatarPreviewUrl.value);
    avatarPreviewUrl.value = null;
  }
};

const handleAvatarSelected = (e) => {
  const file = e.target?.files?.[0];
  selectedAvatarFile.value = file || null;
  // 创建预览 URL
  if (file) {
    avatarPreviewUrl.value = URL.createObjectURL(file);
  } else {
    avatarPreviewUrl.value = null;
  }
};

const uploadAvatar = async (file) => {
  const form = new FormData();
  form.append('file', file);
  form.append('type', 'avatar');
  isUploadingAvatar.value = true;
  avatarUploadProgress.value = 0;

  try {
    const res = await api.post('/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          avatarUploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        }
      },
    });
    return res.data?.file_path;
  } finally {
    isUploadingAvatar.value = false;
    avatarUploadProgress.value = 0;
  }
};

const saveSettings = async () => {
  if (!authStore.isAuthenticated) return;
  savingSettings.value = true;
  try {
    let avatarPath = authStore.user?.avatar || '';
    if (selectedAvatarFile.value) {
      avatarPath = await uploadAvatar(selectedAvatarFile.value);
    }
    const payload = {
      nickname: settingsNickname.value,
      avatar: avatarPath,
      bio: settingsBio.value,
    };
    await authStore.updateMe(payload);
    if (notification) notification.showNotification('保存成功', 'success');
    // 清理预览 URL
    if (avatarPreviewUrl.value) {
      URL.revokeObjectURL(avatarPreviewUrl.value);
      avatarPreviewUrl.value = null;
    }
    showSettings.value = false;
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '保存失败', 'error');
  } finally {
    savingSettings.value = false;
  }
};

const toggleTheme = () => {
  themeStore.toggleTheme();
};

const navigateToChannel = (id) => {
  router.push({ name: 'channel', params: { id } });
};

const handleLogout = () => {
  authStore.logout();
  router.push('/login');
};

const handleContextMenu = (e, item, type) => {
  // Prevent default context menu
  // e.preventDefault(); // Moved to template
  
  contextMenuItem.value = item;
  contextMenuType.value = type;
  
  // Calculate position
  const menuWidth = 160;
  const menuHeight = 100;
  let x = e.clientX;
  let y = e.clientY;
  
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 5;
  }
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 5;
  }
  
  contextMenuX.value = x;
  contextMenuY.value = y;
  showContextMenu.value = true;
};

const closeContextMenu = () => {
  showContextMenu.value = false;
  contextMenuItem.value = null;
};

const handleDelete = () => {
  if (contextMenuType.value === 'note') {
    emit('delete-note', contextMenuItem.value);
  } else if (contextMenuType.value === 'channel') {
    emit('delete-channel', contextMenuItem.value);
  }
  closeContextMenu();
};


// Global click listener to close context menu
onMounted(() => {
  window.addEventListener('click', closeContextMenu);
  checkUnreadNotifications();
  // 每30秒检查一次未读消息
  const interval = setInterval(checkUnreadNotifications, 30000);
  // 存储interval以便清理
  window._notificationInterval = interval;
  // 监听消息更新事件
  eventBus.on('notifications-updated', checkUnreadNotifications);
});

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu);
  if (window._notificationInterval) {
    clearInterval(window._notificationInterval);
  }
  eventBus.off('notifications-updated', checkUnreadNotifications);
});
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: hsl(var(--bc) / 0.2);
  border-radius: 10px;
}
.custom-scrollbar:hover::-webkit-scrollbar-thumb {
  background: hsl(var(--bc) / 0.3);
}

/* 确保菜单项在按下时不会变黑 */
.menu li > a:active {
  background-color: transparent !important;
  color: inherit !important;
}
</style>
