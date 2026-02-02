<template>
  <div class="flex h-full overflow-hidden">
    <!-- Desktop Sidebar -->
    <div class="hidden lg:block w-64 bg-base-200 flex flex-col h-full border-r border-base-300">
      <Sidebar 
        :channels="userChannels" 
        :notes="personalNotes"
        @create-channel="showCreateChannel = true"
        @create-note="showCreateNote = true"
        @delete-note="handleDeleteNote"
        @rename-note="handleRenameNote"
        @delete-channel="handleDeleteChannel"
        @rename-channel="handleRenameChannel"
      />
    </div>

    <!-- Main Content -->
    <main class="flex-1 bg-base-100 flex flex-col min-w-0 relative">
      <!-- Toast Notification -->
      <div v-if="notification" class="absolute top-4 left-1/2 -translate-x-1/2 z-[100] transition-all duration-300 pointer-events-none">
        <div :class="`alert alert-${notification.type} shadow-lg border-none bg-neutral text-neutral-content py-2 px-6 flex items-center gap-2 min-w-[200px] justify-center`">
          <span class="font-medium">{{ notification.message }}</span>
        </div>
      </div>
      <header class="h-12 border-b border-base-300 flex items-center px-4 justify-between shrink-0 relative z-10">
        <div class="flex items-center space-x-2 min-w-0 w-1/4">
          <template v-if="isChannelRoute">
            <div v-if="currentChannelName" class="flex items-center gap-2">
              <Hash class="w-4 h-4" />
              <span class="font-bold text-base-content truncate">{{ currentChannelName }}</span>
            </div>
            <span v-else class="font-bold text-base-content/50 truncate">加载频道中...</span>
          </template>
          <template v-else-if="currentNoteInfo.isChannelNote">
            <div class="flex items-center gap-2 px-2 py-1 bg-neutral/10 rounded-lg text-base-content text-sm">
              <Hash class="w-4 h-4" />
              <span class="font-medium">{{ currentNoteInfo.channelName || '' }}</span>
            </div>
          </template>
          <template v-else-if="currentNoteInfo.ownerId">
            <div class="flex items-center gap-2 px-2 py-1 bg-base-200 rounded-lg text-sm text-base-content">
              <User class="w-4 h-4" />
              <span class="font-medium">{{ currentNoteInfo.ownerName || '用户' }}</span>
            </div>
          </template>
          <template v-else>
            <span class="text-base-content/50">#</span>
            <span class="font-bold text-base-content truncate">{{ currentDocName }}</span>
          </template>
        </div>
        <div class="absolute left-1/2 -translate-x-1/2" v-if="isChannelRoute">
          <div class="inline-flex bg-base-200 rounded-lg px-1 py-1 gap-1">
            <button
              class="btn btn-xs rounded-md w-16"
              :class="channelViewMode === 'chat' ? 'btn-neutral' : 'btn-ghost'"
              @click="channelViewMode = 'chat'"
            >
              聊天
            </button>
            <button
              class="btn btn-xs rounded-md w-16"
              :class="channelViewMode === 'notes' ? 'btn-neutral' : 'btn-ghost'"
              @click="channelViewMode = 'notes'"
            >
              笔记
            </button>
          </div>
        </div>
        <div class="flex items-center space-x-2 w-1/4 justify-end">
          <button
            v-if="isChannelRoute"
            class="btn btn-neutral btn-sm"
            @click="openChannelFiles"
          >
            文件
          </button>
          <button
            class="btn btn-ghost btn-sm btn-circle"
            @click="openSearch"
          >
            <Search class="w-4 h-4" />
          </button>
        </div>
      </header>
      <div class="flex-1 overflow-auto relative pb-12 lg:pb-0">
        <router-view @create-channel="showCreateChannel = true" />
      </div>
    </main>

    <!-- Mobile Bottom Navigation -->
    <div class="lg:hidden fixed bottom-0 left-0 right-0 bg-base-200 border-t border-base-300 z-50">
      <div class="flex items-center justify-around py-2">
        <!-- Home -->
        <button
          class="btn btn-ghost btn-sm btn-square"
          title="主页"
          @click="goToHome"
        >
          <Home class="w-5 h-5" />
        </button>

        <!-- Notes Section -->
        <router-link
          to="/notes"
          class="btn btn-ghost btn-sm btn-square"
          title="笔记"
        >
          <FileText class="w-5 h-5" />
        </router-link>

        <!-- Channels Section -->
        <router-link
          to="/channels"
          class="btn btn-ghost btn-sm btn-square"
          title="频道"
        >
          <Hash class="w-5 h-5" />
        </router-link>

        <!-- User Profile -->
        <div class="relative user-menu">
          <div class="avatar">
            <div
              class="rounded-full w-8 h-8 bg-neutral text-neutral-content flex items-center justify-center text-xs cursor-pointer hover:bg-neutral-focus transition-colors overflow-hidden"
              @click="showUserMenu = !showUserMenu"
            >
              <img v-if="authStore.user?.avatar" :src="getFileUrl(authStore.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
              <span v-else>{{ avatarChar }}</span>
            </div>
          </div>

          <!-- User Menu -->
          <div v-if="showUserMenu" class="absolute bottom-full right-0 mb-2 w-48 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50 user-menu" @click.stop>
            <div class="p-3 border-b border-base-300">
              <p class="font-bold text-sm">{{ authStore.user?.nickname || authStore.user?.username }}</p>
              <p class="text-xs text-base-content/60">@{{ authStore.user?.username }}</p>
            </div>
            <div class="p-1">
              <router-link
                v-if="authStore.isAuthenticated"
                to="/approvals"
                class="w-full text-left px-3 py-2 text-sm hover:bg-base-200 rounded-lg flex items-center gap-2"
                @click="showUserMenu = false"
              >
                <Bell class="w-4 h-4" />
                通知
              </router-link>
              <button
                v-if="authStore.isAuthenticated"
                @click="openSettings"
                class="w-full text-left px-3 py-2 text-sm hover:bg-base-200 rounded-lg flex items-center gap-2"
              >
                <Settings class="w-4 h-4" />
                设置
              </button>
              <button
                v-if="authStore.isAuthenticated"
                @click="logout"
                class="w-full text-left px-3 py-2 text-sm text-error hover:bg-error/10 rounded-lg flex items-center gap-2"
              >
                <LogOut class="w-4 h-4" />
                退出
              </button>
              <button
                v-else
                @click="goToLogin"
                class="w-full text-left px-3 py-2 text-sm hover:bg-base-200 rounded-lg flex items-center gap-2"
              >
                <LogIn class="w-4 h-4" />
                登录
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile Notes Drawer -->
    <MobileDrawer 
      :show="showNotesDrawer" 
      title="笔记" 
      @close="showNotesDrawer = false"
    >
      <div class="space-y-2">
        <button
          v-if="authStore.isAuthenticated"
          @click="createNewNote"
          class="w-full btn btn-neutral btn-sm"
        >
          <Plus class="w-4 h-4 mr-1" />
          新建笔记
        </button>
        <div v-if="personalNotes.length === 0" class="text-center text-base-content/50 py-8">
          暂无笔记
        </div>
        <div v-else class="space-y-1">
          <div
            v-for="note in personalNotes"
            :key="note.id"
            @click="openNote(note)"
            class="p-3 rounded-lg hover:bg-base-300 cursor-pointer flex items-center gap-3"
          >
            <FileText class="w-4 h-4 shrink-0" />
            <div class="flex-1 min-w-0">
              <div class="font-medium truncate">{{ note.title || '无标题' }}</div>
              <div class="text-xs text-base-content/50">
                {{ note.updated_at ? new Date(note.updated_at).toLocaleString() : '' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </MobileDrawer>

    <!-- Mobile Channels Drawer -->
    <MobileDrawer 
      :show="showChannelsDrawer" 
      title="频道" 
      @close="showChannelsDrawer = false"
    >
      <div class="space-y-2">
        <button
          v-if="authStore.isAuthenticated"
          @click="showCreateChannel = true; showChannelsDrawer = false"
          class="w-full btn btn-neutral btn-sm"
        >
          <Plus class="w-4 h-4 mr-1" />
          新建频道
        </button>
        <div v-if="userChannels.length === 0" class="text-center text-base-content/50 py-8">
          暂无频道
        </div>
        <div v-else class="space-y-1">
          <div
            v-for="channel in userChannels"
            :key="channel.id"
            @click="openChannel(channel)"
            class="p-3 rounded-lg hover:bg-base-300 cursor-pointer flex items-center gap-3"
          >
            <Hash class="w-4 h-4 shrink-0" />
            <div class="flex-1 min-w-0">
              <div class="font-medium truncate">{{ channel.name }}</div>
            </div>
          </div>
        </div>
      </div>
    </MobileDrawer>

    <!-- Create Channel Modal -->
    <dialog :open="showCreateChannel" class="modal modal-bottom sm:modal-middle" @click="handleModalClick($event, 'showCreateChannel')">
      <div class="modal-box">
        <h3 class="font-bold text-lg">新建频道</h3>
        <div class="space-y-3">
          <div>
            <div class="flex items-center gap-3">
              <label class="label py-1 flex-1">
                <span class="label-text text-xs">频道名称</span>
              </label>
              <span class="text-xs text-base-content/60 whitespace-nowrap">
                公开
              </span>
            </div>
            <div class="flex items-center gap-3">
              <input
                v-model="newChannelName"
                type="text"
                class="input input-bordered flex-1 input-sm"
                placeholder="输入频道名称"
                @keyup.enter="createChannel"
                :class="{ 'input-error': channelError }"
              />
              <input type="checkbox" v-model="newChannelIsPublic" class="toggle toggle-sm" />
            </div>
            <div v-if="channelError" class="text-error text-sm mt-1">{{ channelError }}</div>
          </div>
          <div>
            <label class="label py-1">
              <span class="label-text text-xs">描述</span>
            </label>
            <textarea
              v-model="newChannelDescription"
              class="textarea textarea-bordered w-full h-20 text-sm"
              placeholder="输入频道描述"
            ></textarea>
          </div>
          <div>
            <label class="label py-1">
              <span class="label-text text-xs">标签</span>
            </label>
            <div class="space-y-2">
              <input
                v-model="newChannelTagInput"
                type="text"
                class="input input-bordered input-sm w-full"
                placeholder="输入标签后按回车添加"
                @keyup.enter="handleAddChannelTag"
              />
              <div v-if="newChannelTags.length" class="flex flex-wrap gap-2">
                <span
                  v-for="tag in newChannelTags"
                  :key="tag"
                  class="badge badge-ghost cursor-pointer hover:bg-base-300"
                  @click="removeChannelTag(tag)"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeChannelModal" :disabled="creatingChannel">取消</button>
          <button class="btn btn-neutral" @click="createChannel" :disabled="creatingChannel || !newChannelName.trim()">
            {{ creatingChannel ? '创建中..' : '创建' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- Create Note Modal -->
    <dialog :open="showCreateNote" class="modal modal-bottom sm:modal-middle" @click="handleModalClick($event, 'showCreateNote')">
      <div class="modal-box">
        <h3 class="font-bold text-lg">新建笔记</h3>
        <div class="space-y-3">
          <div>
            <div class="flex items-center gap-3">
              <label class="label py-1 flex-1">
                <span class="label-text text-xs">笔记标题</span>
              </label>
              <span class="text-xs text-base-content/60 whitespace-nowrap">
                公开
              </span>
            </div>
            <div class="flex items-center gap-3">
              <input
                v-model="newNoteTitle"
                type="text"
                placeholder="输入笔记标题"
                class="input input-bordered flex-1 input-sm"
                @keyup.enter="createNote"
                :class="{ 'input-error': noteError }"
              />
              <input type="checkbox" v-model="newNoteIsPublic" class="toggle toggle-sm" />
            </div>
            <div v-if="noteError" class="text-error text-sm mt-1">{{ noteError }}</div>
          </div>
          <div>
            <label class="label py-1">
              <span class="label-text text-xs">标签</span>
            </label>
            <div class="space-y-2">
              <input
                v-model="newNoteTagInput"
                type="text"
                class="input input-bordered input-sm w-full"
                placeholder="输入标签后按回车添加"
                @keyup.enter="handleAddNoteTag"
              />
              <div v-if="newNoteTags.length" class="flex flex-wrap gap-2">
                <span
                  v-for="tag in newNoteTags"
                  :key="tag"
                  class="badge badge-ghost cursor-pointer hover:bg-base-300"
                  @click="removeNoteTag(tag)"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeNoteModal" :disabled="creatingNote">取消</button>
          <button class="btn btn-neutral" @click="createNote" :disabled="creatingNote || !newNoteTitle.trim()">
            {{ creatingNote ? '创建中..' : '创建' }}
          </button>
        </div>
      </div>
    </dialog>

    <!-- Search Modal -->
    <dialog :open="showSearchModal" class="modal modal-bottom sm:modal-middle" @click="handleModalClick($event, 'showSearchModal')">
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">搜索</h3>
        <div class="form-control mb-4">
          <div class="join w-full">
            <input
              v-model="searchQuery"
              type="text"
              class="input input-bordered join-item flex-1"
              placeholder="输入搜索关键词..."
              @input="debouncedSearch"
            />
            <button
              class="btn btn-neutral join-item"
              :disabled="searching"
            >
              搜索
            </button>
          </div>
        </div>
        <div class="max-h-96 overflow-y-auto">
          <div v-if="searching" class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else-if="searchResults.length === 0 && searchQuery" class="text-center py-8 text-base-content/50">
            未找到相关内容
          </div>
          <div v-else-if="searchResults.length === 0 && !searchQuery" class="text-center py-8 text-base-content/50">
            请输入搜索关键词
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="item in searchResults"
              :key="item.id"
              @click="handleSearchResultClick(item)"
              class="p-3 rounded-lg border border-base-300 hover:border-base-content hover:shadow-md cursor-pointer transition-all"
            >
              <!-- 笔记类型 -->
              <template v-if="item.type === 'note'">
                <div class="flex items-start gap-2">
                  <FileText class="w-5 h-5 text-base-content/30 shrink-0 mt-0.5" />
                  <div class="flex-1 min-w-0">
                    <div class="font-semibold text-sm truncate">{{ item.title || '无标题' }}</div>
                    <div v-if="item.matchReasons && item.matchReasons.length" class="mt-1 flex flex-wrap gap-1">
                      <span v-for="reason in item.matchReasons" :key="reason" class="badge badge-neutral badge-xs">{{ reason }}匹配</span>
                    </div>
                    <div class="flex items-center gap-2 mt-1">
                      <span class="text-[10px] text-base-content/60">
                        {{ item.owner?.username || '未知用户' }}
                      </span>
                      <span v-if="item.channel_id" class="badge badge-ghost badge-xs">频道笔记</span>
                    </div>
                    <div v-if="item.tags" class="mt-1 flex flex-wrap gap-1">
                      <span v-for="tag in parseTags(item.tags)" :key="tag" class="badge badge-ghost badge-xs">{{ tag }}</span>
                    </div>
                    <div class="text-[10px] text-base-content/40 mt-1">
                      {{ item.updated_at ? new Date(item.updated_at).toLocaleString() : '' }}
                    </div>
                  </div>
                </div>
              </template>
              <!-- 频道类型 -->
              <template v-else-if="item.type === 'channel'">
                <div class="flex items-start gap-2">
                  <Hash class="w-5 h-5 text-base-content/30 shrink-0 mt-0.5" />
                  <div class="flex-1 min-w-0">
                    <div class="font-semibold text-sm truncate">{{ item.name }}</div>
                    <div v-if="item.matchReasons && item.matchReasons.length" class="mt-1 flex flex-wrap gap-1">
                      <span v-for="reason in item.matchReasons" :key="reason" class="badge badge-neutral badge-xs">{{ reason }}匹配</span>
                    </div>
                    <div v-if="item.description" class="text-xs text-base-content/60 truncate mt-1">
                      {{ item.description }}
                    </div>
                    <div v-if="item.tags" class="mt-1 flex flex-wrap gap-1">
                      <span v-for="tag in parseTags(item.tags)" :key="tag" class="badge badge-ghost badge-xs">{{ tag }}</span>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="closeSearch">关闭</button>
        </div>
      </div>
    </dialog>

    <!-- Settings Modal -->
    <dialog :open="showSettings" class="modal modal-bottom sm:modal-middle" @click="handleModalClick($event, 'showSettings')">
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
              <span class="label-text">个人简介</span>
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
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, computed, reactive, provide } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { Plus, FileText, Hash, Settings, User, Home, Bell, LogOut, LogIn, Search } from 'lucide-vue-next';
import api from '../api/axios';
import Sidebar from '../components/Sidebar.vue';
import MobileDrawer from '../components/MobileDrawer.vue';
import eventBus from '../utils/eventBus';
import wsClient from '../utils/websocket';
import { getFileUrl } from '../utils/urlHelper';

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const userChannels = ref([]);
const personalNotes = ref([]);
const currentDocName = ref('内容');
const currentNoteInfo = reactive({
  isChannelNote: false,
  ownerId: null,
  ownerName: '',
  channelName: ''
});
const showCreateChannel = ref(false);
const newChannelName = ref('');
const newChannelIsPublic = ref(false);
const newChannelDescription = ref('');
const newChannelTagInput = ref('');
const newChannelTags = ref([]);
const creatingChannel = ref(false);
const channelError = ref('');

const showCreateNote = ref(false);
const newNoteTitle = ref('');
const newNoteIsPublic = ref(false);
const newNoteTags = ref([]);
const newNoteTagInput = ref('');
const creatingNote = ref(false);
const noteError = ref('');
const notification = ref(null);
const showNotesDrawer = ref(false);
const showChannelsDrawer = ref(false);
const showUserMenu = ref(false);
const showSettings = ref(false);
const settingsNickname = ref('');
const settingsBio = ref('');
const selectedAvatarFile = ref(null);
const isUploadingAvatar = ref(false);
const avatarUploadProgress = ref(0);
const savingSettings = ref(false);
const avatarInputRef = ref(null);
const avatarPreviewUrl = ref(null);
const autoRefreshInterval = ref(null);

const channelViewMode = ref('chat');
const channelHeaderBridge = reactive({ openFiles: null });
const currentChannel = ref(null);
const showSearchModal = ref(false);
const searchQuery = ref('');
const searchResults = ref([]);
const searching = ref(false);

const isChannelRoute = computed(() => route.name === 'channel');

const isHomeRoute = computed(() => route.name === 'home');

const currentChannelName = computed(() => {
  if (!isChannelRoute.value) return '';
  // 优先使用从API获取的当前频道信息
  if (currentChannel.value && currentChannel.value.id === Number(route.params.id)) {
    return currentChannel.value.name;
  }
  // 回退到用户频道列表
  const id = Number(route.params.id);
  const found = userChannels.value.find((ch) => ch.id === id);
  return found ? found.name : '';
});

// 获取当前频道信息（包括公开频道）
const fetchCurrentChannel = async () => {
  if (!isChannelRoute.value) {
    currentChannel.value = null;
    return;
  }
  try {
    const id = route.params.id;
    const res = await api.get(`/channels/${id}`);
    currentChannel.value = res.data.channel;
  } catch (e) {
    console.error('Failed to fetch current channel:', e);
    currentChannel.value = null;
  }
};

const avatarChar = computed(() => {
  const name = authStore.user?.username || '访客';
  return name.charAt(0).toUpperCase();
});

const showNotification = (message, type = 'success') => {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 2000);
};

provide('notification', { showNotification });
provide('channelViewMode', channelViewMode);
provide('channelHeaderBridge', channelHeaderBridge);

const openChannelFiles = () => {
  if (channelHeaderBridge.openFiles) {
    channelHeaderBridge.openFiles();
  }
};

const openSearch = () => {
  showSearchModal.value = true;
  searchQuery.value = '';
  searchResults.value = [];
};

const closeSearch = () => {
  showSearchModal.value = false;
  searchQuery.value = '';
  searchResults.value = [];
};

const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    return;
  }

  searching.value = true;
  try {
    const query = searchQuery.value.trim().toLowerCase();
    
    // 分别请求笔记和频道，避免一个失败影响另一个
    const notesPromise = api.get('/notes/search', { params: { q: searchQuery.value.trim() } }).catch(() => ({ data: [] }));
    const channelsPromise = api.get('/public/channels').catch(() => ({ data: [] }));
    
    const [notesRes, channelsRes] = await Promise.all([notesPromise, channelsPromise]);

    const results = [];

    // 处理笔记搜索结果
    (notesRes.data || []).forEach(note => {
      const reasons = [];
      if (note.title && note.title.toLowerCase().includes(query)) {
        reasons.push('标题');
      }
      if (note.content && note.content.toLowerCase().includes(query)) {
        reasons.push('内容');
      }
      if (note.tags && note.tags.toLowerCase().includes(query)) {
        reasons.push('标签');
      }
      results.push({
        type: 'note',
        ...note,
        matchReasons: reasons
      });
    });

    // 处理频道搜索结果（只搜索公开频道）
    (channelsRes.data || []).forEach(channel => {
      const reasons = [];
      if (channel.name && channel.name.toLowerCase().includes(query)) {
        reasons.push('名称');
      }
      if (channel.description && channel.description.toLowerCase().includes(query)) {
        reasons.push('描述');
      }
      if (channel.tags && channel.tags.toLowerCase().includes(query)) {
        reasons.push('标签');
      }
      // 只有匹配的频道才添加到结果中
      if (reasons.length > 0) {
        results.push({
          type: 'channel',
          ...channel,
          matchReasons: reasons
        });
      }
    });

    searchResults.value = results;
  } catch (err) {
    if (notification && notification.value?.showNotification) {
      notification.value.showNotification(err.response?.data?.error || '搜索失败', 'error');
    }
    searchResults.value = [];
  } finally {
    searching.value = false;
  }
};

// 防抖函数
let searchTimeout = null;
const debouncedSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = setTimeout(() => {
    handleSearch();
  }, 300);
};

const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const openSearchResult = (note) => {
  closeSearch();
  router.push({ name: 'note-editor', params: { id: note.id } });
};

const handleSearchResultClick = (item) => {
  closeSearch();
  if (item.type === 'note') {
    router.push({ name: 'note-editor', params: { id: item.id } });
  } else if (item.type === 'channel') {
    router.push({ name: 'channel', params: { id: item.id } });
  }
};

const openSettings = async () => {
  showUserMenu.value = false;
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

const handleModalClick = (event, modalRefName) => {
  if (event.target.tagName === 'DIALOG') {
    if (modalRefName === 'showCreateChannel') showCreateChannel.value = false;
    else if (modalRefName === 'showCreateNote') showCreateNote.value = false;
    else if (modalRefName === 'showSearchModal') showSearchModal.value = false;
    else if (modalRefName === 'showSettings') showSettings.value = false;
  }
};

const goToHome = () => {
  showUserMenu.value = false;
  router.push('/');
};

const scrollToSection = (section) => {
  if (route.name !== 'home') {
    router.push('/');
    // 等待路由切换完成后滚动
    setTimeout(() => {
      const element = document.getElementById(section);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth' });
      }
    }, 100);
  } else {
    const element = document.getElementById(section);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  }
};

const logout = () => {
  showUserMenu.value = false;
  authStore.logout();
  router.push('/');
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

const goToLogin = () => {
  showUserMenu.value = false;
  router.push('/login');
};

const createNewNote = () => {
  showNotesDrawer.value = false;
  router.push('/note');
};

const openNote = (note) => {
  showNotesDrawer.value = false;
  router.push({ name: 'note-editor', params: { id: note.id } });
};

const openChannel = (channel) => {
  showChannelsDrawer.value = false;
  router.push({ name: 'channel', params: { id: channel.id } });
};

const fetchChannels = async () => {
  try {
    if (authStore.isAuthenticated) {
      // 已登录用户获取自己的频道
      const res = await api.get('/channels');
      userChannels.value = res.data;
    } else {
      // 访客获取公开频道
      const res = await api.get('/public/channels');
      userChannels.value = res.data;
    }
  } catch (e) {
    console.error("Failed to fetch channels", e);
    userChannels.value = [];
  }
};

const handleAddChannelTag = () => {
  const tag = newChannelTagInput.value.trim();
  if (tag && !newChannelTags.value.includes(tag)) {
    newChannelTags.value.push(tag);
  }
  newChannelTagInput.value = '';
};

const removeChannelTag = (tag) => {
  newChannelTags.value = newChannelTags.value.filter(t => t !== tag);
};

const createChannel = async () => {
  if (!newChannelName.value.trim()) {
    channelError.value = '请输入频道名称';
    return;
  }

  if (newChannelName.value.trim().length < 2) {
    channelError.value = '频道名称至少需要2个字符';
    return;
  }

  creatingChannel.value = true;
  channelError.value = '';

  try {
    const res = await api.post('/channels', {
      name: newChannelName.value.trim(),
      description: newChannelDescription.value.trim(),
      is_public: newChannelIsPublic.value,
      tags: newChannelTags.value.join(',')
    });

    closeChannelModal();
    fetchChannels(); // 刷新频道列表

    // Navigate to the newly created channel
    router.push({ name: 'channel', params: { id: res.data.id } });
  } catch (err) {
    channelError.value = err.response?.data?.error || '创建频道失败';
  } finally {
    creatingChannel.value = false;
  }
};

const closeChannelModal = () => {
  showCreateChannel.value = false;
  newChannelName.value = '';
  newChannelIsPublic.value = false;
  newChannelDescription.value = '';
  newChannelTagInput.value = '';
  newChannelTags.value = [];
  channelError.value = '';
};

const handleAddNoteTag = () => {
  const tag = newNoteTagInput.value.trim();
  if (tag && !newNoteTags.value.includes(tag)) {
    newNoteTags.value.push(tag);
  }
  newNoteTagInput.value = '';
};

const removeNoteTag = (tag) => {
  newNoteTags.value = newNoteTags.value.filter(t => t !== tag);
};

const createNote = async () => {
  if (!newNoteTitle.value.trim()) {
    noteError.value = '请输入笔记标题';
    return;
  }

  creatingNote.value = true;
  noteError.value = '';

  try {
    const res = await api.post('/notes', {
      title: newNoteTitle.value.trim(),
      content: '',
      is_public: newNoteIsPublic.value,
      tags: newNoteTags.value.join(',')
    });

    closeNoteModal();
    fetchNotes(); // 刷新笔记列表
    router.push({ name: 'note-editor', params: { id: res.data.id } });
  } catch (err) {
    noteError.value = err.response?.data?.error || '创建笔记失败';
  } finally {
    creatingNote.value = false;
  }
};

const closeNoteModal = () => {
  showCreateNote.value = false;
  newNoteTitle.value = '';
  newNoteIsPublic.value = false;
  newNoteTags.value = [];
  newNoteTagInput.value = '';
  noteError.value = '';
};

const handleDeleteNote = async (note) => {
  if (!confirm(`确定要删除笔记"${note.title || '无标题'}" 吗？`)) return;
  try {
    await api.delete(`/notes/${note.id}`);
    fetchNotes();
    if (route.name === 'note-editor' && route.params.id == note.id) {
      router.push('/');
    }
    if (notification) {
      notification.showNotification('笔记已删除', 'success');
    }
  } catch (e) {
    console.error("Failed to delete note", e);
    if (notification) {
      notification.showNotification(e.response?.data?.error || '删除失败', 'error');
    }
  }
};

const handleRenameNote = async (note) => {
  try {
    await api.put(`/notes/${note.id}`, { title: note.title });
    fetchNotes();
  } catch (e) {
    console.error("Failed to rename note", e);
  }
};

const handleDeleteChannel = async (channel) => {
  if (!confirm(`确定要删除频道"${channel.name}" 吗？此操作可能会删除该频道下的所有笔记。`)) return;
  try {
    await api.delete(`/channels/${channel.id}`);
    fetchChannels();
    fetchNotes();
    if (route.name === 'channel' && route.params.id == channel.id) {
      router.push('/');
    }
    if (notification) {
      notification.showNotification('频道已删除', 'success');
    }
  } catch (e) {
    console.error("Failed to delete channel", e);
    if (notification) {
      notification.showNotification(e.response?.data?.error || '删除失败', 'error');
    }
  }
};

const handleRenameChannel = async (channel) => {
  try {
    const payload = {
      name: channel.name,
      description: channel.description || '',
      is_public: channel.is_public !== undefined ? channel.is_public : false,
      tags: channel.tags || ''
    };
    await api.put(`/channels/${channel.id}`, payload);
    fetchChannels();
  } catch (e) {
    console.error("Failed to update channel", e);
  }
};

const fetchNotes = async () => {
  try {
    if (authStore.isAuthenticated) {
      // 已登录用户获取自己的笔记
      const res = await api.get('/notes', { params: { channel_id: 0 } });
      personalNotes.value = res.data;
    } else {
      // 访客获取公开笔记
      const res = await api.get('/public/notes');
      personalNotes.value = res.data;
    }
  } catch (e) {
    console.error("Failed to fetch notes", e);
    personalNotes.value = [];
  }
};

const applyPersonalNotePatch = (payload) => {
  if (!payload || !payload.id) return;
  const target = personalNotes.value.find((note) => note.id === Number(payload.id));
  if (!target) return;
  if (typeof payload.title === 'string') target.title = payload.title;
  if (typeof payload.is_public === 'boolean') target.is_public = payload.is_public;
  if (typeof payload.tags === 'string') target.tags = payload.tags;
};

const handleNoteTitleUpdated = (payload) => {
  if (!payload || !payload.id) return;
  applyPersonalNotePatch({ id: payload.id, title: payload.title });
};

const handleNoteInfoChanged = (payload) => {
  if (!payload) return;
  currentNoteInfo.isChannelNote = payload.isChannelNote;
  currentNoteInfo.ownerId = payload.ownerId;
  currentNoteInfo.ownerName = payload.ownerName || '';
  currentNoteInfo.channelName = payload.channelName || '';
};

// WebSocket 消息处理
const handleWsMessage = (message) => {
  if (message.type === 'note') {
    if (message.action === 'create') {
      // 添加新笔记
      fetchNotes();
    } else if (message.action === 'update') {
      // 更新笔记
      applyPersonalNotePatch({
        id: message.data.id,
        title: message.data.title,
        is_public: message.data.is_public,
        tags: message.data.tags
      });
    } else if (message.action === 'delete') {
      // 删除笔记（确保类型一致的比较）
      const deleteId = String(message.data.id);
      personalNotes.value = personalNotes.value.filter(n => String(n.id) !== deleteId);
      if (route.name === 'note-editor' && route.params.id == message.data.id) {
        router.push('/');
      }
    }
  } else if (message.type === 'channel') {
    if (message.action === 'create') {
      // 添加新频道
      fetchChannels();
    } else if (message.action === 'update') {
      // 更新频道
      const channel = userChannels.value.find(c => c.id === message.data.id);
      if (channel) {
        Object.assign(channel, message.data);
      }
    } else if (message.action === 'delete') {
      // 删除频道（确保类型一致的比较）
      const deleteId = String(message.data.id);
      userChannels.value = userChannels.value.filter(c => String(c.id) !== deleteId);
      if (route.name === 'channel' && route.params.id == message.data.id) {
        router.push('/');
      }
    }
  }
};

onMounted(() => {
  // 如果已登录（有token），恢复用户状态并连接 WebSocket
  if (localStorage.getItem('token') && authStore.isAuthenticated) {
    authStore.refreshMe();
  }

  fetchChannels();
  fetchNotes();
  fetchCurrentChannel();
  eventBus.on('note-title-updated', handleNoteTitleUpdated);
  eventBus.on('note-updated', applyPersonalNotePatch);
  eventBus.on('note-info-changed', handleNoteInfoChanged);

  // 全局点击事件监听器，用于关闭用户菜单
  window.addEventListener('click', (e) => {
    if (showUserMenu.value) {
      const userMenu = e.target.closest('.user-menu');
      if (!userMenu) {
        showUserMenu.value = false;
      }
    }
  });

  // 等待 WebSocket 连接建立后再设置监听器
  const setupWebSocketListeners = () => {
    if (localStorage.getItem('userId')) {
      // 监听所有 WebSocket 消息
      wsClient.on('note_create', handleWsMessage);
      wsClient.on('note_update', handleWsMessage);
      wsClient.on('note_delete', handleWsMessage);
      wsClient.on('channel_create', handleWsMessage);
      wsClient.on('channel_update', handleWsMessage);
      wsClient.on('channel_delete', handleWsMessage);
    }
  };

  // 如果已经有userId，立即设置监听器
  setupWebSocketListeners();

  // 监听 WebSocket 连接成功事件，确保连接后设置监听器
  wsClient.on('connected', setupWebSocketListeners);

  // 30秒自动刷新笔记和频道数据（作为备用）
  autoRefreshInterval.value = setInterval(() => {
    fetchNotes();
    fetchChannels();
  }, 30000);
});

onBeforeUnmount(() => {
  eventBus.off('note-title-updated', handleNoteTitleUpdated);
  eventBus.off('note-updated', applyPersonalNotePatch);
  eventBus.off('note-info-changed', handleNoteInfoChanged);

  // 清除 WebSocket 监听器
  wsClient.off('note_create', handleWsMessage);
  wsClient.off('note_update', handleWsMessage);
  wsClient.off('note_delete', handleWsMessage);
  wsClient.off('channel_create', handleWsMessage);
  wsClient.off('channel_update', handleWsMessage);
  wsClient.off('channel_delete', handleWsMessage);
  wsClient.off('connected', () => {}); // 移除 connected 事件监听器
  // 清除自动刷新定时器
  if (autoRefreshInterval.value) {
    clearInterval(autoRefreshInterval.value);
  }
});

watch(
  () => route.fullPath,
  (newPath, oldPath) => {
    // 只在离开笔记页面时重置笔记信息
    if (oldPath.includes('/note/') && !newPath.includes('/note/')) {
      currentNoteInfo.isChannelNote = false;
      currentNoteInfo.ownerId = null;
      currentNoteInfo.ownerName = '';
      currentNoteInfo.channelName = '';
    }
    fetchNotes();
    fetchCurrentChannel();
    if (route.name === 'home') {
      currentDocName.value = '首页';
    }
    if (route.name === 'channel') {
      channelViewMode.value = 'chat';
    }
  }
);

// 监听移动端抽屉打开时刷新数据
watch(showNotesDrawer, (isOpen) => {
  if (isOpen) {
    fetchNotes();
  }
});

watch(showChannelsDrawer, (isOpen) => {
  if (isOpen) {
    fetchChannels();
  }
});
</script>
