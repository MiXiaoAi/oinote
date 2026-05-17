<template>
  <div class="flex h-full overflow-hidden">
    <!-- Sidebar -->
    <div class="hidden lg:block w-64 bg-base-200 flex flex-col h-full border-r border-base-300 shrink-0">
      <div class="h-12 border-b border-base-300 flex items-center px-4 font-bold cursor-pointer hover:bg-base-300 transition-colors" 
           @click="router.push('/')">
        <FileText class="w-5 h-5 mr-2" />
        <span class="text-lg tracking-tight">oinote</span>
      </div>
      <div class="flex-1 p-4">
        <div class="text-sm text-base-content/50 mb-2">后台管理</div>
        <ul class="menu p-0 w-full">
          <li>
            <a :class="{ 'active': activeTab === 'stats' }" @click="activeTab = 'stats'; loadTabData()">
              <BarChart3 class="w-4 h-4" />
              系统统计
            </a>
          </li>
          <li>
            <a :class="{ 'active': activeTab === 'users' }" @click="activeTab = 'users'; loadTabData()">
              <Users class="w-4 h-4" />
              用户管理
            </a>
          </li>
          <li>
            <a :class="{ 'active': activeTab === 'notes' }" @click="activeTab = 'notes'; loadTabData()">
              <FileText class="w-4 h-4" />
              笔记管理
            </a>
          </li>
          <li>
            <a :class="{ 'active': activeTab === 'channels' }" @click="activeTab = 'channels'; loadTabData()">
              <Hash class="w-4 h-4" />
              频道管理
            </a>
          </li>
          <li>
            <a :class="{ 'active': activeTab === 'ai' }" @click="activeTab = 'ai'; loadAISettings()">
              <Bot class="w-4 h-4" />
              AI 配置
            </a>
          </li>
        </ul>
      </div>
      <div class="p-4 border-t border-base-300">
        <router-link to="/" class="btn btn-ghost btn-sm w-full justify-start">
          <ArrowLeft class="w-4 h-4 mr-2" />
          返回首页
        </router-link>
      </div>
    </div>

    <!-- Main Content -->
    <main class="flex-1 bg-base-100 flex flex-col min-w-0 overflow-hidden">
      <!-- Toast Notification -->
      <div v-if="notification" class="absolute top-4 left-1/2 -translate-x-1/2 z-[100] transition-all duration-300 pointer-events-none">
        <div :class="`alert alert-${notification.type} shadow-lg border-none bg-neutral text-neutral-content py-2 px-6 flex items-center gap-2 min-w-[200px] justify-center`">
          <span class="font-medium">{{ notification.message }}</span>
        </div>
      </div>

      <!-- Mobile Header -->
      <div class="lg:hidden h-12 border-b border-base-300 flex items-center px-4 justify-between shrink-0">
        <div class="flex items-center gap-2">
          <router-link to="/" class="btn btn-ghost btn-sm btn-square">
            <ArrowLeft class="w-5 h-5" />
          </router-link>
          <span class="font-bold">后台管理</span>
        </div>
        <select v-model="activeTab" class="select select-bordered select-sm" @change="loadTabData">
          <option value="stats">统计</option>
          <option value="users">用户</option>
          <option value="notes">笔记</option>
          <option value="channels">频道</option>
          <option value="ai">AI</option>
        </select>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-auto p-4 lg:p-6">
        <!-- 系统统计 Tab -->
        <div v-if="activeTab === 'stats'">
          <h2 class="text-lg font-bold mb-4 hidden lg:block">系统总览</h2>
          <div v-if="loadingStats" class="flex items-center justify-center py-16">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="space-y-6">
            <!-- 主要统计卡片 -->
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
              <div class="stat bg-gradient-to-br from-primary/10 to-primary/5 rounded-xl p-5 border border-primary/20">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="stat-title text-base-content/60">用户总数</div>
                    <div class="stat-value text-primary text-3xl mt-1">{{ stats.user_count }}</div>
                    <div class="stat-desc text-primary/60 mt-1">
                      <span v-if="stats.today_new_users > 0" class="flex items-center gap-1">
                        <TrendingUp class="w-3 h-3" />
                        今日 +{{ stats.today_new_users }}
                      </span>
                      <span v-else>今日无新增</span>
                    </div>
                  </div>
                  <div class="w-12 h-12 rounded-xl bg-primary/20 flex items-center justify-center">
                    <Users class="w-6 h-6 text-primary" />
                  </div>
                </div>
              </div>

              <div class="stat bg-gradient-to-br from-secondary/10 to-secondary/5 rounded-xl p-5 border border-secondary/20">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="stat-title text-base-content/60">笔记总数</div>
                    <div class="stat-value text-secondary text-3xl mt-1">{{ stats.note_count }}</div>
                    <div class="stat-desc text-secondary/60 mt-1">
                      <span v-if="stats.today_new_notes > 0" class="flex items-center gap-1">
                        <TrendingUp class="w-3 h-3" />
                        今日 +{{ stats.today_new_notes }}
                      </span>
                      <span v-else>今日无新增</span>
                    </div>
                  </div>
                  <div class="w-12 h-12 rounded-xl bg-secondary/20 flex items-center justify-center">
                    <FileText class="w-6 h-6 text-secondary" />
                  </div>
                </div>
              </div>

              <div class="stat bg-gradient-to-br from-accent/10 to-accent/5 rounded-xl p-5 border border-accent/20">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="stat-title text-base-content/60">频道总数</div>
                    <div class="stat-value text-accent text-3xl mt-1">{{ stats.channel_count }}</div>
                    <div class="stat-desc text-accent/60 mt-1">协作空间</div>
                  </div>
                  <div class="w-12 h-12 rounded-xl bg-accent/20 flex items-center justify-center">
                    <Hash class="w-6 h-6 text-accent" />
                  </div>
                </div>
              </div>

              <div class="stat bg-gradient-to-br from-info/10 to-info/5 rounded-xl p-5 border border-info/20">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="stat-title text-base-content/60">消息总数</div>
                    <div class="stat-value text-info text-3xl mt-1">{{ stats.message_count }}</div>
                    <div class="stat-desc text-info/60 mt-1">
                      <span v-if="stats.today_new_messages > 0" class="flex items-center gap-1">
                        <TrendingUp class="w-3 h-3" />
                        今日 +{{ stats.today_new_messages }}
                      </span>
                      <span v-else>今日无新增</span>
                    </div>
                  </div>
                  <div class="w-12 h-12 rounded-xl bg-info/20 flex items-center justify-center">
                    <MessageSquare class="w-6 h-6 text-info" />
                  </div>
                </div>
              </div>
            </div>

            <!-- 次要统计 -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div class="bg-base-200 rounded-xl p-5 border border-base-300">
                <div class="flex items-center gap-3 mb-3">
                  <div class="w-10 h-10 rounded-lg bg-success/20 flex items-center justify-center">
                    <Paperclip class="w-5 h-5 text-success" />
                  </div>
                  <div>
                    <div class="text-sm text-base-content/60">附件文件</div>
                    <div class="text-xl font-bold">{{ stats.attachment_count }}</div>
                  </div>
                </div>
                <div class="text-xs text-base-content/50">
                  占用空间：{{ formatFileSize(stats.total_file_size) }}
                </div>
              </div>

              <div class="bg-base-200 rounded-xl p-5 border border-base-300">
                <div class="flex items-center gap-3 mb-3">
                  <div class="w-10 h-10 rounded-lg bg-warning/20 flex items-center justify-center">
                    <BookOpen class="w-5 h-5 text-warning" />
                  </div>
                  <div>
                    <div class="text-sm text-base-content/60">频道笔记</div>
                    <div class="text-xl font-bold">{{ stats.channel_note_count }}</div>
                  </div>
                </div>
                <div class="w-full bg-base-300 rounded-full h-2 mt-2">
                  <div class="bg-warning rounded-full h-2 transition-all duration-500" 
                       :style="{ width: `${stats.note_count > 0 ? (stats.channel_note_count / stats.note_count * 100) : 0}%` }"></div>
                </div>
                <div class="text-xs text-base-content/50 mt-1">
                  占比：{{ stats.note_count > 0 ? Math.round(stats.channel_note_count / stats.note_count * 100) : 0 }}%
                </div>
              </div>

              <div class="bg-base-200 rounded-xl p-5 border border-base-300">
                <div class="flex items-center gap-3 mb-3">
                  <div class="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center">
                    <UserCheck class="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <div class="text-sm text-base-content/60">个人笔记</div>
                    <div class="text-xl font-bold">{{ stats.personal_note_count }}</div>
                  </div>
                </div>
                <div class="w-full bg-base-300 rounded-full h-2 mt-2">
                  <div class="bg-primary rounded-full h-2 transition-all duration-500" 
                       :style="{ width: `${stats.note_count > 0 ? (stats.personal_note_count / stats.note_count * 100) : 0}%` }"></div>
                </div>
                <div class="text-xs text-base-content/50 mt-1">
                  占比：{{ stats.note_count > 0 ? Math.round(stats.personal_note_count / stats.note_count * 100) : 0 }}%
                </div>
              </div>
            </div>

            <!-- 最近活动 -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <!-- 最近注册用户 -->
              <div class="bg-base-200 rounded-xl border border-base-300 overflow-hidden">
                <div class="px-5 py-4 border-b border-base-300 flex items-center justify-between">
                  <h3 class="font-semibold flex items-center gap-2">
                    <UserPlus class="w-4 h-4 text-primary" />
                    最近注册用户
                  </h3>
                  <button class="btn btn-ghost btn-xs" @click="activeTab = 'users'; loadTabData()">
                    查看全部
                    <ChevronRight class="w-3 h-3" />
                  </button>
                </div>
                <div class="divide-y divide-base-300">
                  <div v-for="user in stats.recent_users" :key="user.id" 
                       class="px-5 py-3 flex items-center gap-3 hover:bg-base-300/50 transition-colors">
                    <div class="avatar">
                      <div class="rounded-full w-9 h-9 bg-neutral text-neutral-content flex items-center justify-center text-sm font-medium shrink-0 overflow-hidden">
                        <span>{{ (user.nickname || user.username || '?').charAt(0).toUpperCase() }}</span>
                      </div>
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="font-medium text-sm truncate">{{ user.nickname || user.username }}</div>
                      <div class="text-xs text-base-content/50">@{{ user.username }}</div>
                    </div>
                    <div class="text-xs text-base-content/50">
                      {{ formatDate(user.created_at) }}
                    </div>
                  </div>
                  <div v-if="!stats.recent_users || stats.recent_users.length === 0" 
                       class="px-5 py-8 text-center text-base-content/50 text-sm">
                    暂无用户
                  </div>
                </div>
              </div>

              <!-- 最近创建笔记 -->
              <div class="bg-base-200 rounded-xl border border-base-300 overflow-hidden">
                <div class="px-5 py-4 border-b border-base-300 flex items-center justify-between">
                  <h3 class="font-semibold flex items-center gap-2">
                    <FilePlus class="w-4 h-4 text-secondary" />
                    最近创建笔记
                  </h3>
                  <button class="btn btn-ghost btn-xs" @click="activeTab = 'notes'; loadTabData()">
                    查看全部
                    <ChevronRight class="w-3 h-3" />
                  </button>
                </div>
                <div class="divide-y divide-base-300">
                  <div v-for="note in stats.recent_notes" :key="note.id" 
                       class="px-5 py-3 hover:bg-base-300/50 transition-colors">
                    <div class="flex items-center gap-3">
                      <div class="w-9 h-9 rounded-lg flex items-center justify-center"
                           :class="note.channel_id ? 'bg-success/20' : 'bg-base-300'">
                        <FileText class="w-4 h-4" :class="note.channel_id ? 'text-success' : 'text-base-content/60'" />
                      </div>
                      <div class="flex-1 min-w-0">
                        <router-link :to="`/note/${note.id}`" class="font-medium text-sm hover:underline truncate block">
                          {{ note.title || '无标题' }}
                        </router-link>
                        <div class="text-xs text-base-content/50 flex items-center gap-2">
                          <span>{{ note.owner?.nickname || note.owner?.username || '未知' }}</span>
                          <span v-if="note.channel_id" class="badge badge-success badge-xs">频道</span>
                        </div>
                      </div>
                      <div class="text-xs text-base-content/50">
                        {{ formatDate(note.created_at) }}
                      </div>
                    </div>
                  </div>
                  <div v-if="!stats.recent_notes || stats.recent_notes.length === 0" 
                       class="px-5 py-8 text-center text-base-content/50 text-sm">
                    暂无笔记
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 用户管理 Tab -->
        <div v-if="activeTab === 'users'">
          <h2 class="text-lg font-bold mb-4 hidden lg:block">用户管理</h2>
          <div v-if="loadingUsers" class="flex items-center justify-center py-16">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="bg-base-200 rounded-lg border border-base-300 overflow-hidden">
            <div class="overflow-x-auto">
              <table class="table table-zebra w-full">
                <thead>
                  <tr class="bg-base-300">
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
                    <td class="font-mono">{{ user.id }}</td>
                    <td>{{ user.username }}</td>
                    <td>{{ user.nickname || '-' }}</td>
                    <td>
                      <select class="select select-bordered select-xs w-28"
                              @change="updateUserRole(user.id, $event.target.value)"
                              :disabled="updatingRole === user.id || user.id === authStore.user?.id">
                        <option value="member" :selected="user.role === 'member'">普通用户</option>
                        <option value="admin" :selected="user.role === 'admin'">管理员</option>
                      </select>
                      <span v-if="updatingRole === user.id" class="loading loading-spinner loading-xs ml-2"></span>
                    </td>
                    <td>{{ formatDate(user.created_at) }}</td>
                    <td>
                      <button @click="deleteUser(user.id)" 
                              class="btn btn-xs btn-error text-white" 
                              :disabled="deletingUser === user.id || user.id === authStore.user?.id">
                        {{ deletingUser === user.id ? '删除中...' : '删除' }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="users.length === 0" class="text-center py-12 text-base-content/50">
              暂无用户
            </div>
          </div>
        </div>

        <!-- 笔记管理 Tab -->
        <div v-if="activeTab === 'notes'">
          <h2 class="text-lg font-bold mb-4 hidden lg:block">笔记管理</h2>
          <div v-if="loadingNotes" class="flex items-center justify-center py-16">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="bg-base-200 rounded-lg border border-base-300 overflow-hidden">
            <div class="overflow-x-auto">
              <table class="table table-zebra w-full">
                <thead>
                  <tr class="bg-base-300">
                    <th>ID</th>
                    <th>标题</th>
                    <th>作者</th>
                    <th>类型</th>
                    <th>创建时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="note in notes" :key="note.id">
                    <td class="font-mono">{{ note.id }}</td>
                    <td>
                      <router-link :to="`/note/${note.id}`" class="hover:underline">
                        {{ note.title || '无标题' }}
                      </router-link>
                    </td>
                    <td>{{ note.owner?.nickname || note.owner?.username || '-' }}</td>
                    <td>
                      <span v-if="note.channel_id" class="badge badge-success badge-sm">频道笔记</span>
                      <span v-else class="badge badge-ghost badge-sm">个人笔记</span>
                    </td>
                    <td>{{ formatDate(note.created_at) }}</td>
                    <td>
                      <button @click="deleteNote(note.id)" 
                              class="btn btn-xs btn-error text-white" 
                              :disabled="deletingNote === note.id">
                        {{ deletingNote === note.id ? '删除中...' : '删除' }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="notes.length === 0" class="text-center py-12 text-base-content/50">
              暂无笔记
            </div>
          </div>
        </div>

        <!-- 频道管理 Tab -->
        <div v-if="activeTab === 'channels'">
          <h2 class="text-lg font-bold mb-4 hidden lg:block">频道管理</h2>
          <div v-if="loadingChannels" class="flex items-center justify-center py-16">
            <span class="loading loading-spinner loading-lg text-neutral"></span>
          </div>
          <div v-else class="bg-base-200 rounded-lg border border-base-300 overflow-hidden">
            <div class="overflow-x-auto">
              <table class="table table-zebra w-full">
                <thead>
                  <tr class="bg-base-300">
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
                  <tr v-for="channel in channels" :key="channel.id">
                    <td class="font-mono">{{ channel.id }}</td>
                    <td>
                      <router-link :to="`/channel/${channel.id}`" class="hover:underline flex items-center gap-2">
                        <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: channel.theme_color }"></div>
                        {{ channel.name }}
                      </router-link>
                    </td>
                    <td>{{ channel.owner?.nickname || channel.owner?.username || '-' }}</td>
                    <td>{{ channel.member_count || 0 }}</td>
                    <td>
                      <button @click="toggleChannelPublic(channel)" 
                              class="btn btn-xs text-white" 
                              :class="channel.is_public ? 'btn-success' : 'btn-warning'">
                        {{ channel.is_public ? '公开' : '私密' }}
                      </button>
                    </td>
                    <td>{{ formatDate(channel.created_at) }}</td>
                    <td>
                      <button @click="deleteChannel(channel.id)" 
                              class="btn btn-xs btn-error text-white" 
                              :disabled="deletingChannel === channel.id">
                        {{ deletingChannel === channel.id ? '删除中...' : '删除' }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="channels.length === 0" class="text-center py-12 text-base-content/50">
              暂无频道
            </div>
          </div>
        </div>

        <!-- AI 配置 Tab -->
        <div v-if="activeTab === 'ai'">
          <h2 class="text-lg font-bold mb-4 hidden lg:block">AI 配置</h2>
          <div class="bg-base-200 rounded-lg border border-base-300 p-6">
            <div class="max-w-lg space-y-4">
              <div>
                <label class="label">
                  <span class="label-text font-medium">OpenAI URL</span>
                </label>
                <input v-model="aiConfig.openai_url" type="text" class="input input-bordered w-full" placeholder="https://api.openai.com/v1" />
              </div>
              <div>
                <label class="label">
                  <span class="label-text font-medium">API Key</span>
                </label>
                <input v-model="aiConfig.api_key" type="password" class="input input-bordered w-full" placeholder="sk-..." />
              </div>
              <div>
                <label class="label">
                  <span class="label-text font-medium">Model</span>
                </label>
                <input v-model="aiConfig.model" type="text" class="input input-bordered w-full" placeholder="gpt-3.5-turbo" />
              </div>
              <div class="pt-4">
                <button class="btn btn-neutral" @click="saveAISettings" :disabled="savingAI">
                  <span v-if="savingAI" class="loading loading-spinner loading-sm"></span>
                  {{ savingAI ? '保存中...' : '保存配置' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import api from '../api/axios';
import { Settings, Users, FileText, Hash, ArrowLeft, BarChart3, Bot, TrendingUp, MessageSquare, Paperclip, BookOpen, UserCheck, UserPlus, FilePlus, ChevronRight } from 'lucide-vue-next';

const router = useRouter();
const authStore = useAuthStore();

// Notification
const notification = ref(null);
const showNotification = (message, type = 'success') => {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 2000);
};

// Tab state
const activeTab = ref('stats');

// Stats
const stats = ref({ 
  user_count: 0, 
  note_count: 0, 
  channel_count: 0,
  message_count: 0,
  attachment_count: 0,
  total_file_size: 0,
  recent_users: [],
  recent_notes: [],
  today_new_users: 0,
  today_new_notes: 0,
  today_new_messages: 0,
  channel_note_count: 0,
  personal_note_count: 0
});
const loadingStats = ref(false);

// Users
const users = ref([]);
const loadingUsers = ref(false);
const updatingRole = ref(null);
const deletingUser = ref(null);

// Notes
const notes = ref([]);
const loadingNotes = ref(false);
const deletingNote = ref(null);

// Channels
const channels = ref([]);
const loadingChannels = ref(false);
const deletingChannel = ref(null);

// AI Config
const aiConfig = ref({ openai_url: '', api_key: '', model: '' });
const savingAI = ref(false);

// 检查管理员权限
const checkAdmin = () => {
  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    showNotification('需要管理员权限', 'error');
    router.push('/');
    return false;
  }
  return true;
};

// 加载统计数据
const loadStats = async () => {
  loadingStats.value = true;
  try {
    const res = await api.get('/admin/stats');
    stats.value = res.data;
  } catch (err) {
    showNotification('加载统计信息失败', 'error');
  } finally {
    loadingStats.value = false;
  }
};

// 加载用户列表
const loadUsers = async () => {
  loadingUsers.value = true;
  try {
    const res = await api.get('/admin/users');
    users.value = res.data;
  } catch (err) {
    showNotification('加载用户列表失败', 'error');
  } finally {
    loadingUsers.value = false;
  }
};

// 更新用户角色
const updateUserRole = async (userId, role) => {
  if (userId === authStore.user?.id) {
    showNotification('不能更改自己的角色', 'error');
    return;
  }
  updatingRole.value = userId;
  try {
    await api.put(`/admin/users/${userId}/role`, { role });
    showNotification('角色更新成功', 'success');
    await loadUsers();
  } catch (err) {
    showNotification(err.response?.data?.error || '更新失败', 'error');
  } finally {
    updatingRole.value = null;
  }
};

// 删除用户
const deleteUser = async (userId) => {
  if (userId === authStore.user?.id) {
    showNotification('不能删除自己', 'error');
    return;
  }
  if (!confirm('确定要删除这个用户吗？此操作不可恢复！')) return;
  
  deletingUser.value = userId;
  try {
    await api.delete(`/admin/users/${userId}`);
    showNotification('删除成功', 'success');
    await loadUsers();
    await loadStats();
  } catch (err) {
    showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingUser.value = null;
  }
};

// 加载笔记列表
const loadNotes = async () => {
  loadingNotes.value = true;
  try {
    const res = await api.get('/admin/notes');
    notes.value = res.data;
  } catch (err) {
    showNotification('加载笔记列表失败', 'error');
  } finally {
    loadingNotes.value = false;
  }
};

// 删除笔记
const deleteNote = async (noteId) => {
  if (!confirm('确定要删除这个笔记吗？此操作不可恢复！')) return;
  
  deletingNote.value = noteId;
  try {
    await api.delete(`/admin/notes/${noteId}`);
    showNotification('删除成功', 'success');
    await loadNotes();
    await loadStats();
  } catch (err) {
    showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingNote.value = null;
  }
};

// 加载频道列表
const loadChannels = async () => {
  loadingChannels.value = true;
  try {
    const res = await api.get('/admin/channels');
    channels.value = res.data;
  } catch (err) {
    showNotification('加载频道列表失败', 'error');
  } finally {
    loadingChannels.value = false;
  }
};

// 切换频道公开状态
const toggleChannelPublic = async (channel) => {
  try {
    await api.put(`/admin/channels/${channel.id}/public`, { is_public: !channel.is_public });
    showNotification(channel.is_public ? '已设为私密' : '已设为公开', 'success');
    await loadChannels();
  } catch (err) {
    showNotification(err.response?.data?.error || '切换失败', 'error');
  }
};

// 删除频道
const deleteChannel = async (channelId) => {
  if (!confirm('确定要删除这个频道吗？此操作不可恢复！')) return;
  
  deletingChannel.value = channelId;
  try {
    await api.delete(`/channels/${channelId}`);
    showNotification('删除成功', 'success');
    await loadChannels();
    await loadStats();
  } catch (err) {
    showNotification(err.response?.data?.error || '删除失败', 'error');
  } finally {
    deletingChannel.value = null;
  }
};

// 加载 AI 配置
const loadAISettings = async () => {
  try {
    const res = await api.get('/admin/ai-config');
    aiConfig.value = {
      openai_url: res.data.openai_url || '',
      api_key: res.data.api_key || '',
      model: res.data.model || ''
    };
  } catch (err) {
    aiConfig.value = { openai_url: '', api_key: '', model: '' };
  }
};

// 保存 AI 配置
const saveAISettings = async () => {
  savingAI.value = true;
  try {
    await api.put('/admin/ai-config', {
      openai_url: aiConfig.value.openai_url || '',
      api_key: aiConfig.value.api_key || '',
      model: aiConfig.value.model || ''
    });
    showNotification('保存成功', 'success');
  } catch (err) {
    showNotification(err.response?.data?.error || '保存失败', 'error');
  } finally {
    savingAI.value = false;
  }
};

// 根据 Tab 加载数据
const loadTabData = () => {
  switch (activeTab.value) {
    case 'stats': loadStats(); break;
    case 'users': loadUsers(); break;
    case 'notes': loadNotes(); break;
    case 'channels': loadChannels(); break;
  }
};

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('zh-CN');
};

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
};

onMounted(() => {
  if (checkAdmin()) {
    loadStats();
  }
});
</script>