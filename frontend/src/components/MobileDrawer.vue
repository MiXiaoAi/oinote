<template>
  <div v-if="show" class="fixed inset-0 z-50 lg:hidden">
    <!-- Backdrop -->
    <div 
      class="absolute inset-0 bg-black/50" 
      @click="closeDrawer"
    ></div>
    
    <!-- Drawer Content -->
    <div 
      class="absolute bottom-0 left-0 right-0 bg-base-200 rounded-t-2xl max-h-[80vh] overflow-hidden"
      @click.stop
    >
      <!-- Handle -->
      <div class="flex justify-center py-2">
        <div class="w-12 h-1 bg-base-content/30 rounded-full"></div>
      </div>
      
      <!-- Header -->
      <div class="px-4 pb-2 border-b border-base-300">
        <h3 class="text-lg font-bold">{{ title }}</h3>
      </div>
      
      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4">
        <slot></slot>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  }
});

const emit = defineEmits(['close']);

const closeDrawer = () => {
  emit('close');
};
</script>
