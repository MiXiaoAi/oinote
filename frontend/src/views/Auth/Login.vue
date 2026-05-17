<template>
  <div class="min-h-screen flex items-center justify-center bg-base-200">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="card-title justify-center text-2xl font-bold mb-4 cursor-pointer hover:text-neutral transition-colors" @click="router.push('/')">
          🗒 oinote
        </div>
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="form-control">
            <label class="label"><span class="label-text">用户名</span></label>
            <input v-model="username" type="text" class="input input-bordered w-full" required />
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">密码</span></label>
            <input v-model="password" type="password" class="input input-bordered w-full" required />
          </div>
          <div class="form-control mt-6">
            <button class="btn btn-neutral w-full" :disabled="loading">
              {{ loading ? '登录中...' : '登录' }}
            </button>
          </div>
        </form>
        <div class="text-center mt-4 space-y-2 text-sm text-base-content">
          <div>
            <button @click="showChangePassword = true" class="link link-neutral text-base-content">修改密码</button>
          </div>
          <div>
            没有账号？<router-link to="/register" class="link link-neutral text-base-content">立即注册</router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Change Password Modal -->
    <div v-if="showChangePassword" class="modal modal-open">
      <div class="modal-box max-w-sm">
        <div class="flex items-center justify-between mb-6">
          <h3 class="font-bold text-lg">修改密码</h3>
          <button @click="closePasswordModal" class="btn btn-ghost btn-sm btn-circle">
            <X class="h-4 w-4" />
          </button>
        </div>
        
        <form @submit.prevent="handleChangePassword" class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">用户名</span>
            </label>
            <input
              v-model="changeUsername"
              type="text"
              placeholder="请输入用户名"
              class="input input-bordered input-sm w-full focus:ring-2 focus:ring-neutral"
              required
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">当前密码</span>
            </label>
            <input
              v-model="currentPassword"
              type="password"
              placeholder="请输入当前密码"
              class="input input-bordered input-sm w-full focus:ring-2 focus:ring-neutral"
              required
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">新密码</span>
            </label>
            <input
              v-model="newPassword"
              type="password"
              placeholder="请输入新密码"
              class="input input-bordered input-sm w-full focus:ring-2 focus:ring-neutral"
              required
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">确认新密码</span>
            </label>
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
              class="input input-bordered input-sm w-full focus:ring-2 focus:ring-neutral"
              required
            />
          </div>

          <div v-if="passwordError" class="alert alert-error text-sm py-2">
            <AlertTriangle class="stroke-current shrink-0 h-4 w-4" />
            <span>{{ passwordError }}</span>
          </div>

          <div class="modal-action pt-4">
            <button type="button" @click="closePasswordModal" class="btn btn-outline btn-sm" :disabled="changingPassword">
              取消
            </button>
            <button type="submit" class="btn btn-neutral btn-sm" :disabled="changingPassword">
              <span v-if="changingPassword" class="loading loading-spinner loading-xs"></span>
              {{ changingPassword ? '修改中...' : '确认修改' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Toast Notification -->
    <div v-if="notification" class="toast toast-top toast-center z-50">
      <div :class="`alert ${notification.type === 'success' ? 'bg-success text-white' : 'alert-' + notification.type}`">
        <span>{{ notification.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../../stores/auth';
import { useRouter } from 'vue-router';
import api from '../../api/axios';
import { X, AlertTriangle } from 'lucide-vue-next';

const authStore = useAuthStore();
const router = useRouter();
const username = ref('');
const password = ref('');
const loading = ref(false);

// Password change modal state
const showChangePassword = ref(false);
const changeUsername = ref('');
const currentPassword = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const changingPassword = ref(false);
const passwordError = ref('');
const notification = ref(null);

const handleLogin = async () => {
  loading.value = true;
  try {
    await authStore.login(username.value, password.value);
    showNotification('登录成功', 'success');
    
    // 检查是否有重定向地址
    const redirect = router.currentRoute.value.query.redirect || '/';
    setTimeout(() => router.push(redirect), 1000);
  } catch (err) {
    showNotification('登录失败: ' + (err.response?.data?.error || '未知错误'), 'error');
  } finally {
    loading.value = false;
  }
};

const handleChangePassword = async () => {
  if (!changeUsername.value.trim() || !currentPassword.value || !newPassword.value || !confirmPassword.value) {
    passwordError.value = '请填写所有字段';
    return;
  }

  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = '两次输入的密码不一致';
    return;
  }

  if (newPassword.value.length < 6) {
    passwordError.value = '新密码至少需要6个字符';
    return;
  }

  changingPassword.value = true;
  passwordError.value = '';

  try {
    await api.post('/auth/change-password', {
      username: changeUsername.value.trim(),
      current_password: currentPassword.value,
      new_password: newPassword.value
    });
    
    showNotification('密码修改成功', 'success');
    closePasswordModal();
  } catch (err) {
    passwordError.value = err.response?.data?.error || '密码修改失败';
  } finally {
    changingPassword.value = false;
  }
};

const showNotification = (message, type = 'info') => {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 3000);
};

const closePasswordModal = () => {
  showChangePassword.value = false;
  changeUsername.value = '';
  currentPassword.value = '';
  newPassword.value = '';
  confirmPassword.value = '';
  passwordError.value = '';
};
</script>
