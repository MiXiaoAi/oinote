<template>
  <div class="min-h-screen flex items-center justify-center bg-base-200">
    <div class="card w-96 bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="card-title justify-center text-2xl font-bold mb-4 cursor-pointer hover:text-neutral transition-colors" @click="router.push('/')">
          🗒 oinote
        </div>
        <form @submit.prevent="handleRegister" class="space-y-4">
          <div class="form-control">
            <label class="label"><span class="label-text">用户名</span></label>
            <input v-model="username" type="text" class="input input-bordered w-full" required />
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">密码</span></label>
            <input v-model="password" type="password" class="input input-bordered w-full" required />
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">确认密码</span></label>
            <input v-model="confirmPassword" type="password" class="input input-bordered w-full" required />
          </div>
          <div v-if="registerError" class="text-error text-sm">{{ registerError }}</div>
          <div class="form-control mt-6">
            <button class="btn btn-neutral w-full" :disabled="loading">
              {{ loading ? '注册中...' : '注册' }}
            </button>
          </div>
        </form>
        <div class="text-center mt-4 space-y-2 text-sm text-base-content">
          <div>
            已有账号？<router-link to="/login" class="link link-neutral text-base-content">返回登录</router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast Notification -->
    <div v-if="notification" class="toast toast-top toast-center z-50">
      <div :class="`alert alert-${notification.type}`">
        <span>{{ notification.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../../stores/auth';
import { useRouter } from 'vue-router';

const authStore = useAuthStore();
const router = useRouter();
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const loading = ref(false);
const registerError = ref('');
const notification = ref(null);

const handleRegister = async () => {
  registerError.value = '';
  
  if (!username.value.trim()) {
    registerError.value = '请输入用户名';
    return;
  }

  if (username.value.trim().length < 2) {
    registerError.value = '用户名至少需要2个字符';
    return;
  }

  if (!password.value) {
    registerError.value = '请输入密码';
    return;
  }

  if (password.value.length < 6) {
    registerError.value = '密码至少需要6个字符';
    return;
  }

  if (password.value !== confirmPassword.value) {
    registerError.value = '两次输入的密码不一致';
    return;
  }

  loading.value = true;
  try {
    await authStore.register(username.value, password.value);
    showNotification('注册成功，请登录', 'success');
    setTimeout(() => router.push('/login'), 1500);
  } catch (err) {
    registerError.value = err.response?.data?.error || '注册失败';
  } finally {
    loading.value = false;
  }
};

const showNotification = (message, type = 'info') => {
  notification.value = { message, type };
  setTimeout(() => {
    notification.value = null;
  }, 3000);
};
</script>
