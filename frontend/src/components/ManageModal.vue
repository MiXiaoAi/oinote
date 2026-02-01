<template>
  <dialog :open="open" class="modal modal-bottom sm:modal-middle">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg mb-2">管理{{ type === 'channel' ? '频道' : '笔记' }}</h3>
      <div class="space-y-3">
        <div>
          <div class="flex items-center gap-3">
            <label class="label py-1 flex-1">
              <span class="label-text text-xs">{{ type === 'channel' ? '频道名称' : '标题' }}</span>
            </label>
            <span class="text-xs text-base-content/60 whitespace-nowrap">
              公开
            </span>
          </div>
          <div class="flex items-center gap-3">
            <input
              v-model="formData.title"
              type="text"
              class="input input-bordered flex-1 input-sm pr-12"
              placeholder="输入名称"
              @keyup.enter="handleSave"
            />
            <input type="checkbox" v-model="formData.isPublic" class="toggle toggle-sm" />
          </div>
        </div>
        <div v-if="type === 'channel'">
          <label class="label py-1">
            <span class="label-text text-xs">描述</span>
          </label>
          <textarea
            v-model="formData.description"
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
              v-model="tagInput"
              type="text"
              class="input input-bordered input-sm w-full"
              placeholder="输入标签后按回车添加"
              @keyup.enter.prevent="handleAddTag"
            />
            <div v-if="tags.length" class="flex flex-wrap gap-2">
              <span
                v-for="tag in tags"
                :key="tag"
                class="badge badge-ghost cursor-pointer hover:bg-base-300"
                :class="{ 'bg-base-200': selectedTags.includes(tag) }"
                @click="toggleTag(tag)"
              >
                {{ tag }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="type === 'channel'" class="border-t border-base-300 pt-3">
          <div
            class="font-medium text-sm mb-2 cursor-pointer hover:text-base-content/70 flex items-center gap-2"
            @click="showMembers = !showMembers"
          >
            <span>成员管理</span>
            <svg
              class="w-4 h-4 transition-transform"
              :class="{ 'rotate-180': showMembers }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
          <div v-if="showMembers" class="mt-3 space-y-3">
            <div>
              <label class="label py-1">
                <span class="label-text text-xs">邀请用户</span>
              </label>
              <div class="flex items-center gap-2">
                <input
                  v-model="inviteUsername"
                  type="text"
                  class="input input-bordered input-sm w-full"
                  placeholder="输入用户名"
                  @keyup.enter="handleInviteMember"
                />
                <button class="btn btn-neutral btn-sm" @click="handleInviteMember" :disabled="inviteLoading || !inviteUsername.trim()">
                  邀请
                </button>
              </div>
            </div>
            <div>
              <div class="text-sm font-medium mb-2">成员列表</div>
              <div v-if="loading" class="text-xs text-base-content/50">加载中...</div>
              <div v-else-if="members.length === 0" class="text-xs text-base-content/50">暂无成员</div>
              <ul v-else class="space-y-2 max-h-48 overflow-y-auto">
                <li v-for="member in members" :key="member.id" class="flex items-center justify-between gap-2">
                  <div class="min-w-0">
                    <div class="text-sm truncate">
                      {{ member.user?.nickname || member.user?.username || '成员' }}
                    </div>
                    <div class="text-[10px] text-base-content/50 truncate">
                      {{ member.role }}
                    </div>
                  </div>
                  <button
                    class="btn btn-ghost btn-xs text-error hover:bg-error/10 hover:text-error/80"
                    @click="handleRemoveMember(member)"
                    :disabled="isRemoveDisabled(member) || removeLoadingId === member.user_id"
                  >
                    移出
                  </button>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-action">
        <button class="btn" @click="handleClose" :disabled="saving">取消</button>
        <button class="btn btn-neutral" @click="handleSave" :disabled="saving || !formData.title.trim()">保存</button>
      </div>
    </div>
  </dialog>
</template>

<script setup>
import { ref, computed, watch, inject } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import api from '../api/axios';

const props = defineProps({
  open: Boolean,
  type: {
    type: String,
    default: 'note',
  },
  item: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['close', 'save']);

const router = useRouter();
const authStore = useAuthStore();
const notification = inject('notification');

const saving = ref(false);
const loading = ref(false);
const inviteLoading = ref(false);
const removeLoadingId = ref(null);
const inviteUsername = ref('');
const members = ref([]);

const formData = ref({
  title: '',
  description: '',
  isPublic: false,
  tagsInput: '',
});

const tagInput = ref('');
const selectedTags = ref([]);
const showMembers = ref(false);

const allTags = ref(new Set());

const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const tags = computed(() => parseTags(formData.value.tagsInput));

const handleAddTag = () => {
  const tag = tagInput.value.trim();
  if (tag && !tags.value.includes(tag)) {
    const newTags = [...tags.value, tag];
    formData.value.tagsInput = newTags.join(',');
    allTags.value.add(tag);
    selectedTags.value = newTags;
  }
  tagInput.value = '';
};

const toggleTag = (tag) => {
  if (selectedTags.value.includes(tag)) {
    selectedTags.value = selectedTags.value.filter(t => t !== tag);
  } else {
    selectedTags.value = [...selectedTags.value, tag];
  }
  formData.value.tagsInput = selectedTags.value.join(',');
};

const updateSelectedTags = () => {
  selectedTags.value = tags.value;
  tags.value.forEach(tag => allTags.value.add(tag));
};

const isRemoveDisabled = (member) => {
  if (!member) return true;
  if (member.role === 'owner') return true;
  if (member.user_id && authStore.user?.id && member.user_id === authStore.user.id) return true;
  return false;
};

const fetchMembers = async () => {
  if (!props.item || props.type !== 'channel') return;
  loading.value = true;
  try {
    const res = await api.get(`/channels/${props.item.id}`);
    members.value = res.data?.members || [];
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '加载成员失败', 'error');
  } finally {
    loading.value = false;
  }
};

const handleInviteMember = async () => {
  if (!props.item || !inviteUsername.value.trim()) return;
  inviteLoading.value = true;
  try {
    await api.post('/channels/invite', {
      channel_id: props.item.id,
      username: inviteUsername.value.trim(),
    });
    inviteUsername.value = '';
    await fetchMembers();
    if (notification) notification.showNotification('邀请已发送', 'success');
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '邀请失败', 'error');
  } finally {
    inviteLoading.value = false;
  }
};

const handleRemoveMember = async (member) => {
  if (!props.item || !member?.user_id) return;
  if (isRemoveDisabled(member)) return;
  removeLoadingId.value = member.user_id;
  try {
    await api.delete(`/channels/${props.item.id}/members/${member.user_id}`);
    members.value = members.value.filter((m) => m.user_id !== member.user_id);
    if (notification) notification.showNotification('已移出成员', 'success');
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '移出失败', 'error');
  } finally {
    removeLoadingId.value = null;
  }
};

const handleSave = async () => {
  if (!props.item || !formData.value.title.trim()) return;
  saving.value = true;
  try {
    const payload = {
      title: formData.value.title.trim(),
      is_public: !!formData.value.isPublic,
      description: formData.value.description.trim(),
      tags: selectedTags.value.join(','),
    };
    emit('save', payload);
  } catch (err) {
    if (notification) notification.showNotification(err.response?.data?.error || '保存失败', 'error');
  } finally {
    saving.value = false;
  }
};

const handleClose = () => {
  emit('close');
};

const resetForm = () => {
  formData.value = {
    title: '',
    description: '',
    isPublic: false,
    tagsInput: '',
  };
  tagInput.value = '';
  selectedTags.value = [];
  inviteUsername.value = '';
  members.value = [];
  showMembers.value = false;
};

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && props.item) {
      if (props.type === 'note') {
        formData.value.title = props.item.title || '';
        formData.value.description = '';
        formData.value.isPublic = !!props.item.is_public;
        formData.value.tagsInput = props.item.tags || '';
        updateSelectedTags();
      } else if (props.type === 'channel') {
        formData.value.title = props.item.name || '';
        formData.value.description = props.item.description || '';
        formData.value.isPublic = !!props.item.is_public;
        formData.value.tagsInput = props.item.tags || '';
        updateSelectedTags();
        fetchMembers();
      }
    } else if (!isOpen) {
      resetForm();
    }
  }
);
</script>