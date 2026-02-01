<template>
  <div class="h-full bg-base-100 text-base-content" :data-theme="themeStore.theme">
    <router-view />
  </div>
</template>

<script setup>
import { useThemeStore } from './stores/theme';
import { onMounted } from 'vue';

const themeStore = useThemeStore();

onMounted(() => {
  themeStore.initTheme();
});
</script>
