<template>
  <div class="w-64 bg-base-200 flex flex-col h-full border-r border-base-300">
    <!-- Header -->
    <div class="h-12 border-b border-base-300 flex items-center px-4 font-bold cursor-pointer hover:bg-base-300 transition-colors shrink-0" 
         @click="router.push('/')">
      <FileText class="w-5 h-5 mr-2" />
      <span class="text-lg tracking-tight">oinote</span>
      <button v-if="authStore.isAuthenticated && authStore.user?.role === 'admin'" 
              @click.stop="showAdminPanel = true" 
              class="ml-auto btn btn-ghost btn-xs btn-square" 
              title="后台管理">
        <Settings class="w-4 h-4" />
      </button>
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
          <router-link to="/approvals" class="btn btn-ghost btn-xs btn-circle" title="消息通知">
            <Bell class="w-4 h-4" />
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
            <p>版本 1.0.0 | © 2024 oinote</p>
          </div>
        </div>
        <div class="modal-action">
          <button class="btn" @click="showAbout = false">关闭</button>
        </div>
      </div>
    </dialog>

    <!-- Admin Panel Dialog -->
    <dialog :open="showAdminPanel" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg flex items-center gap-2">
          <Settings class="w-5 h-5" />
          后台管理
        </h3>
        
        <!-- Tabs -->
        <div class="tabs tabs-boxed mb-4">
          <a class="tab" :class="{ 'tab-active': adminTab === 'stats' }" @click="adminTab = 'stats'; loadAdminData()">系统统计</a>
          <a class="tab" :class="{ 'tab-active': adminTab === 'users' }" @click="adminTab = 'users'; loadAdminData()">用户管理</a>
          <a class="tab" :class="{ 'tab-active': adminTab === 'notes' }" @click="adminTab = 'notes'; loadAdminData()">笔记管理</a>
          <a class="tab" :class="{ 'tab-active': adminTab === 'channels' }" @click="adminTab = 'channels'; loadAdminData()">频道管理</a>
          <a class="tab" :class="{ 'tab-active': adminTab === 'ai' }" @click="adminTab = 'ai'; openAISettings()">AI 配置</a>
        </div>

        <!-- 系统统计 Tab -->
        <div v-if="adminTab === 'stats'" class="py-4">
          <div v-if="loadingStats" class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="grid grid-cols-3 gap-4">
            <div class="stat bg-base-200 rounded-lg p-4">
              <div class="stat-figure">
                <Users class="w-8 h-8" />
              </div>
              <div class="stat-title">用户总数</div>
              <div class="stat-value">{{ stats.user_count }}</div>
              <div class="stat-desc">注册用户</div>
            </div>
            <div class="stat bg-base-200 rounded-lg p-4">
              <div class="stat-figure">
                <FileText class="w-8 h-8" />
              </div>
              <div class="stat-title">笔记总数</div>
              <div class="stat-value">{{ stats.note_count }}</div>
              <div class="stat-desc">创建的笔记</div>
            </div>
            <div class="stat bg-base-200 rounded-lg p-4">
              <div class="stat-figure">
                <Hash class="w-8 h-8" />
              </div>
              <div class="stat-title">频道总数</div>
              <div class="stat-value">{{ stats.channel_count }}</div>
              <div class="stat-desc">协作频道</div>
            </div>
          </div>
        </div>

        <!-- 用户管理 Tab -->
        <div v-if="adminTab === 'users'" class="py-4">
          <div v-if="loadingUsers" class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="table table-zebra w-full">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>用户名</th>
                  <th>昵称</th>
                  <th>角色</th>
                  <th>注册时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users" :key="user.id">
                  <td>{{ user.id }}</td>
                  <td>{{ user.username }}</td>
                  <td>{{ user.nickname || '-' }}</td>
                  <td>
                    <select class="select select-bordered select-xs w-24" 
                            @change="updateUserRole(user.id, $event.target.value)"
                            :disabled="updatingRole === user.id">
                      <option value="member" :selected="user.role === 'member'">普通用户</option>
                      <option value="admin" :selected="user.role === 'admin'">管理员</option>
                    </select>
                    <span v-if="updatingRole === user.id" class="loading loading-spinner loading-xs ml-2"></span>
                  </td>
                  <td>{{ formatDate(user.created_at) }}</td>
                  <td>
                    <button @click="deleteUser(user.id)" class="btn btn-xs btn-error" style="color: white;" :disabled="deletingUser === user.id">
                      {{ deletingUser === user.id ? '删除中...' : '删除' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 笔记管理 Tab -->
        <div v-if="adminTab === 'notes'" class="py-4">
          <div v-if="loadingNotes" class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="table table-zebra w-full">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>标题</th>
                  <th>作者</th>
                  <th>类型</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="note in adminNotes" :key="note.id">
                  <td>{{ note.id }}</td>
                  <td>{{ note.title || '无标题' }}</td>
                  <td>{{ note.owner?.nickname || note.owner?.username || '-' }}</td>
                  <td>
                    <button v-if="note.channel_id" class="badge badge-success badge-sm">频道笔记</button>
                    <button v-else class="badge badge-ghost badge-sm">个人笔记</button>
                  </td>
                  <td>{{ formatDate(note.created_at) }}</td>
                  <td>
                    <button @click="deleteNote(note.id)" class="btn btn-xs btn-error" style="color: white;" :disabled="deletingNote === note.id">
                      {{ deletingNote === note.id ? '删除中...' : '删除' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-if="adminNotes.length === 0" class="text-center py-8 text-base-content/50">
              暂无笔记
            </div>
          </div>
        </div>

        <!-- 频道管理 Tab -->
        <div v-if="adminTab === 'channels'" class="py-4">
          <div v-if="loadingChannels" class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="table table-zebra w-full">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>频道名称</th>
                  <th>所有者</th>
                  <th>成员数</th>
                  <th>类型</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="channel in adminChannels" :key="channel.id">
                  <td>{{ channel.id }}</td>
                  <td>
                    <div class="flex items-center gap-2">
                      <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: channel.theme_color }"></div>
                      {{ channel.name }}
                    </div>
                  </td>
                  <td>{{ channel.owner?.nickname || channel.owner?.username || '-' }}</td>
                  <td>{{ channel.member_count || 0 }}</td>
                  <td>
                    <button @click="toggleChannelPublic(channel)" class="btn btn-xs" :class="channel.is_public ? 'btn-success' : 'btn-primary'" style="color: white;">
                      {{ channel.is_public ? '公开' : '私密' }}
                    </button>
                  </td>
                  <td>{{ formatDate(channel.created_at) }}</td>
                  <td>
                    <button @click="deleteChannel(channel.id)" class="btn btn-xs btn-error" style="color: white;" :disabled="deletingChannel === channel.id">
                      {{ deletingChannel === channel.id ? '删除中...' : '删除' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-if="adminChannels.length === 0" class="text-center py-8 text-base-content/50">
              暂无频道
            </div>
          </div>
        </div>

        <!-- AI 配置 Tab -->
        <div v-if="adminTab === 'ai'" class="py-4 space-y-3">
          <div>
            <label class="label">
              <span class="label-text">OpenAI URL</span>
            </label>
            <input v-model="aiConfig.openai_url" type="text" class="input input-bordered w-full" />
          </div>
          <div>
            <label class="label">
              <span class="label-text">API Key</span>
            </label>
            <input v-model="aiConfig.api_key" type="password" class="input input-bordered w-full" />
          </div>
          <div>
            <label class="label">
              <span class="label-text">Model</span>
            </label>
            <input v-model="aiConfig.model" type="text" class="input input-bordered w-full" />
          </div>
        </div>

        <div class="modal-action">
          <button class="btn" @click="showAdminPanel = false" :disabled="savingAISettings">取消</button>
          <button v-if="adminTab === 'ai'" class="btn btn-neutral" @click="saveAISettings" :disabled="savingAISettings">
            {{ savingAISettings ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount, inject, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useThemeStore } from '../stores/theme';
import api from '../api/axios';
import eventBus from '../utils/eventBus';
import ManageModal from './ManageModal.vue';
import { getFileUrl } from '../utils/urlHelper';
import {
  Hash, FileText, Sun, Moon, LogOut, Settings,
  Edit, Trash2, Users, Info, User, Plus, ChevronRight, Bell, Bot
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
const showAdminPanel = ref(false);
const adminTab = ref('stats');
const aiConfig = ref({
  openai_url: '',
  api_key: '',
  model: ''
});
const savingAISettings = ref(false);

// 系统统计相关
const stats = ref({
  user_count: 0,
  note_count: 0,
  channel_count: 0
});
const loadingStats = ref(false);

// 用户管理相关
const users = ref([]);
const loadingUsers = ref(false);
const updatingRole = ref(null);
const deletingUser = ref(null);

// 笔记管理相关
const adminNotes = ref([]);
const loadingNotes = ref(false);
const deletingNote = ref(null);

// 频道管理相关
const adminChannels = ref([]);
const loadingChannels = ref(false);
const deletingChannel = ref(null);

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

const openAISettings = async () => {
  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    if (notification) notification.showNotification('需要管理员权限', 'error');
    return;
  }

  adminTab.value = 'ai';

  try {
    const res = await api.get('/admin/ai-config');
    // 使用 nextTick 确保响应式更新
    await new Promise(resolve => setTimeout(resolve, 0));
    aiConfig.value = {
      openai_url: res.data.openai_url || '',
      api_key: res.data.api_key || '',
      model: res.data.model || ''
    };
  } catch (err) {
    console.error('AI配置加载失败:', err);
    // 配置不存在是正常的，使用空配置
    aiConfig.value = { openai_url: '', api_key: '', model: '' };
  }
  showAdminPanel.value = true;
};

const saveAISettings = async () => {
  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    if (notification) notification.showNotification('需要管理员权限', 'error');
    return;
  }

  savingAISettings.value = true;
  try {
    await api.put('/admin/ai-config', {
      openai_url: aiConfig.value.openai_url || '',
      api_key: aiConfig.value.api_key || '',
      model: aiConfig.value.model || ''
    });
    if (notification) notification.showNotification('保存成功', 'success');
    showAdminPanel.value = false;
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '保存失败', 'error');
  } finally {
    savingAISettings.value = false;
  }
};

const loadStats = async () => {
  loadingStats.value = true;
  try {
    const res = await api.get('/admin/stats');
    stats.value = res.data;
  } catch (err) {
    if (notification) notification.showNotification('加载统计信息失败', 'error');
  } finally {
    loadingStats.value = false;
  }
};

const loadUsers = async () => {
  loadingUsers.value = true;
  try {
    const res = await api.get('/admin/users');
    users.value = res.data;
  } catch (err) {
    if (notification) notification.showNotification('加载用户列表失败', 'error');
  } finally {
    loadingUsers.value = false;
  }
};

const loadChannels = async () => {
  loadingChannels.value = true;
  try {
    const res = await api.get('/admin/channels');
    adminChannels.value = res.data;
  } catch (err) {
    if (notification) notification.showNotification('加载频道列表失败', 'error');
  } finally {
    loadingChannels.value = false;
  }
};

const toggleChannelPublic = async (channel) => {
  try {
    await api.put(`/admin/channels/${channel.id}/public`, { is_public: !channel.is_public });
    if (notification) notification.showNotification(channel.is_public ? '已设为私密' : '已设为公开', 'success');
    // 重新加载频道列表
    await loadChannels();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '切换失败', 'error');
  }
};

const deleteChannel = async (channelId) => {
  if (!confirm('确定要删除这个频道吗？此操作不可恢复！')) {
    return;
  }

  deletingChannel.value = channelId;
  try {
    await api.delete(`/channels/${channelId}`);
    if (notification) notification.showNotification('删除成功', 'success');
    // 重新加载频道列表
    await loadChannels();
    // 重新加载统计信息
    await loadStats();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingChannel.value = null;
  }
};

const updateUserRole = async (userId, role) => {
  updatingRole.value = userId;
  try {
    await api.put(`/admin/users/${userId}/role`, { role });
    if (notification) notification.showNotification('角色更新成功', 'success');
    // 重新加载用户列表
    await loadUsers();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '更新失败', 'error');
  } finally {
    updatingRole.value = null;
  }
};

const deleteUser = async (userId) => {
  if (!confirm('确定要删除这个用户吗？此操作不可恢复！')) {
    return;
  }

  deletingUser.value = userId;
  try {
    await api.delete(`/admin/users/${userId}`);
    if (notification) notification.showNotification('删除成功', 'success');
    // 重新加载用户列表
    await loadUsers();
    // 重新加载统计信息
    await loadStats();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingUser.value = null;
  }
};

const loadNotes = async () => {
  loadingNotes.value = true;
  try {
    const res = await api.get('/admin/notes');
    adminNotes.value = res.data;
  } catch (err) {
    if (notification) notification.showNotification('加载笔记列表失败', 'error');
  } finally {
    loadingNotes.value = false;
  }
};

const deleteNote = async (noteId) => {
  if (!confirm('确定要删除这个笔记吗？此操作不可恢复！')) {
    return;
  }

  deletingNote.value = noteId;
  try {
    await api.delete(`/admin/notes/${noteId}`);
    if (notification) notification.showNotification('删除成功', 'success');
    // 重新加载笔记列表
    await loadNotes();
    // 重新加载统计信息
    await loadStats();
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingNote.value = null;
  }
};

const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN');
};

const loadAdminData = async () => {
  if (adminTab.value === 'stats') {
    await loadStats();
  } else if (adminTab.value === 'users') {
    await loadUsers();
  } else if (adminTab.value === 'notes') {
    await loadNotes();
  } else if (adminTab.value === 'channels') {
    await loadChannels();
  }
};

// 监听 adminTab 变化，加载相应数据
const adminTabChanged = () => {
  if (adminTab.value === 'stats') {
    loadStats();
  } else if (adminTab.value === 'users') {
    loadUsers();
  } else if (adminTab.value === 'notes') {
    loadNotes();
  } else if (adminTab.value === 'channels') {
    loadChannels();
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
});

// 监听 adminTab 变化
watch(adminTab, () => {
  adminTabChanged();
});

// 监听 showAdminPanel 变化
watch(showAdminPanel, (newVal) => {
  if (newVal) {
    adminTabChanged();
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu);
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