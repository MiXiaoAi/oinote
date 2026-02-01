<template>
  <div class="flex-1 flex flex-col h-full bg-base-100 p-4">
    <div class="flex items-center gap-4 mb-4">
      <input
        v-model="title"
        type="text"
        :placeholder="loadingNote ? '' : '无标题'"
        :disabled="loadingNote || isReadonly"
        class="input input-ghost text-3xl font-bold flex-1 focus:bg-base-200"
        :class="{ 'cursor-default': isReadonly }"
      />
      <template v-if="canEdit">
        <button @click="saveNote" class="btn btn-neutral" :disabled="saving">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </template>
      <template v-else>
        <span class="text-xs text-base-content/50">只读模式</span>
      </template>
    </div>
    
    <!-- Editor Toolbar (仅编辑模式显示) -->
    <div v-if="canEdit" class="border border-base-300 rounded-t-lg bg-base-200 p-2 flex flex-wrap items-center gap-2">
      <!-- Text Formatting -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().toggleBold().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('bold'), 'bg-base-100 hover:bg-base-300': !editor.isActive('bold') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="粗体">
          <Bold class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleItalic().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('italic'), 'bg-base-100 hover:bg-base-300': !editor.isActive('italic') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="斜体">
          <Italic class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleUnderline().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('underline'), 'bg-base-100 hover:bg-base-300': !editor.isActive('underline') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="下划线">
          <UnderlineIcon class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleStrike().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('strike'), 'bg-base-100 hover:bg-base-300': !editor.isActive('strike') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="删除线">
          <Strikethrough class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleCode().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('code'), 'bg-base-100 hover:bg-base-300': !editor.isActive('code') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="代码">
          <Code class="w-4 h-4" />
        </button>
      </div>

      <!-- Headings -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 1 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 1 }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="标题1">
          <Heading1 class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 2 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 2 }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="标题2">
          <Heading2 class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 3 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 3 }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="标题3">
          <Heading3 class="w-4 h-4" />
        </button>
      </div>

      <!-- Lists -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().toggleBulletList().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('bulletList'), 'bg-base-100 hover:bg-base-300': !editor.isActive('bulletList') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="无序列表">
          <List class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleOrderedList().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('orderedList'), 'bg-base-100 hover:bg-base-300': !editor.isActive('orderedList') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="有序列表">
          <ListOrdered class="w-4 h-4" />
        </button>
      </div>

      <!-- Alignment -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().setTextAlign('left').run()"
                :class="{ 'bg-neutral text-neutral-content border-neutral': isLeftAlignActive, 'bg-base-100 hover:bg-base-300': !isLeftAlignActive }"
                class="btn btn-xs btn-square border transition-all duration-200" title="左对齐">
          <AlignLeft class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().setTextAlign('center').run()"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive({ textAlign: 'center' }), 'bg-base-100 hover:bg-base-300': !editor.isActive({ textAlign: 'center' }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="居中对齐">
          <AlignCenter class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().setTextAlign('right').run()"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive({ textAlign: 'right' }), 'bg-base-100 hover:bg-base-300': !editor.isActive({ textAlign: 'right' }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="右对齐">
          <AlignRight class="w-4 h-4" />
        </button>
      </div>

      <!-- Other -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()"
                class="btn btn-xs btn-square border transition-all duration-200" title="插入表格">
          <TableIcon class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleBlockquote().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('blockquote'), 'bg-base-100 hover:bg-base-300': !editor.isActive('blockquote') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="引用">
          <Quote class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().toggleCodeBlock().run()" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('codeBlock'), 'bg-base-100 hover:bg-base-300': !editor.isActive('codeBlock') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="代码块">
          <Code2 class="w-4 h-4" />
        </button>
      </div>

      <!-- History -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().undo().run()"
                :disabled="!editor.can().undo()"
                :class="{ 'bg-base-100 hover:bg-base-300': editor.can().undo(), 'bg-base-200 opacity-50 cursor-not-allowed': !editor.can().undo() }"
                class="btn btn-xs btn-square border transition-all duration-200" title="撤销">
          <Undo class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().redo().run()"
                :disabled="!editor.can().redo()"
                :class="{ 'bg-base-100 hover:bg-base-300': editor.can().redo(), 'bg-base-200 opacity-50 cursor-not-allowed': !editor.can().redo() }"
                class="btn btn-xs btn-square border transition-all duration-200" title="重做">
          <Redo class="w-4 h-4" />
        </button>
      </div>

      <!-- Upload -->
      <div class="flex items-center gap-1">
        <input
          ref="imageInputRef"
          type="file"
          accept="image/*"
          class="hidden"
          @change="handleImageSelected"
        />
        <button @click="triggerImageUpload" :disabled="isUploading" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="插入图片">
          <ImageIcon class="w-4 h-4" />
        </button>
        <input
          ref="fileInputRef"
          type="file"
          class="hidden"
          @change="handleFileSelected"
        />
        <button @click="triggerFileUpload" :disabled="isUploading" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="插入附件">
          <Paperclip class="w-4 h-4" />
        </button>
      </div>

      <!-- Upload Progress -->
      <div v-if="isUploading" class="flex items-center gap-2 ml-2">
        <span class="text-xs text-primary font-medium">上传中 {{ uploadProgress }}%</span>
        <div class="w-24">
          <progress class="progress progress-primary w-full h-2" :value="uploadProgress" max="100"></progress>
        </div>
      </div>
    </div>

    <!-- Table Toolbar (Conditional) -->
    <div v-if="editor && editor.isActive('table')" class="border-x border-b border-base-300 bg-base-200 p-2 flex flex-wrap items-center gap-2 text-sm shadow-sm">
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().addColumnBefore().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="前插列">
          <ArrowLeftToLine class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().addColumnAfter().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="后插列">
          <ArrowRightToLine class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().deleteColumn().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="删列">
          <Trash2 class="w-4 h-4" />
          <span class="text-[10px] absolute -bottom-1">列</span>
        </button>
      </div>
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().addRowBefore().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="前插行">
          <ArrowUpToLine class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().addRowAfter().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="后插行">
          <ArrowDownToLine class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().deleteRow().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="删行">
          <Trash2 class="w-4 h-4" />
          <span class="text-[10px] absolute -bottom-1">行</span>
        </button>
      </div>
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @click="editor.chain().focus().mergeCells().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="合并单元格">
          <Merge class="w-4 h-4" />
        </button>
        <button @click="editor.chain().focus().splitCell().run()" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200" title="拆分单元格">
          <Split class="w-4 h-4" />
        </button>
      </div>
      <div class="flex items-center gap-1">
        <button @click="editor.chain().focus().deleteTable().run()" class="btn btn-xs btn-neutral btn-outline" title="删除表格">
          <Trash2 class="w-4 h-4 mr-1" />
          删除表格
        </button>
      </div>
    </div>

    <div class="flex-1 min-h-0 border border-base-300 overflow-hidden relative flex"
         :class="[canEdit ? 'border-t-0 rounded-b-lg' : 'rounded-lg', { 'dragging': isDragging }]"
         @contextmenu.prevent="handleContextMenu"
         @dragover.prevent="handleDragOver"
         @dragleave.prevent="handleDragLeave"
         @drop.prevent="handleDrop">
      <!-- Drag overlay -->
      <div v-if="isDragging && canEdit" class="absolute inset-0 bg-primary/10 border-2 border-dashed border-primary rounded-lg z-10 flex items-center justify-center pointer-events-none">
        <div class="text-center">
          <svg class="w-16 h-16 mx-auto mb-2 text-primary opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <p class="text-lg font-medium text-primary">拖放文件到此处</p>
          <p class="text-sm text-base-content/60 mt-1">图片将直接插入，其他文件将作为附件</p>
        </div>
      </div>
      <editor-content 
        :editor="editor" 
        class="flex-1 overflow-auto pt-0 px-4 pb-4 prose max-w-none prose-neutral" 
        @paste="handlePaste"
      />
    </div>

    <!-- Context Menu -->
    <div v-if="showContextMenu && canEdit" 
         class="fixed z-50 bg-base-100 shadow-xl rounded-lg border border-base-200 py-1 min-w-[180px] text-sm"
         :style="{ top: `${contextMenuY}px`, left: `${contextMenuX}px` }">
      
      <!-- Table Menu -->
      <template v-if="contextMenuType === 'table'">
        <div class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">表格操作</div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addColumnBefore().run()">
          <ArrowLeftToLine class="w-4 h-4" /> 前插列
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addColumnAfter().run()">
          <ArrowRightToLine class="w-4 h-4" /> 后插列
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addRowBefore().run()">
          <ArrowUpToLine class="w-4 h-4" /> 前插行
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addRowAfter().run()">
          <ArrowDownToLine class="w-4 h-4" /> 后插行
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().mergeCells().run()">
          <Merge class="w-4 h-4" /> 合并单元格
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().splitCell().run()">
          <Split class="w-4 h-4" /> 拆分单元格
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2 text-base-content" @click="editor.chain().focus().deleteTable().run()">
          <Trash2 class="w-4 h-4" /> 删除表格
        </button>
      </template>

      <!-- Image Menu -->
      <template v-else-if="contextMenuType === 'image'">
        <div class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">图片操作</div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editImage">
          <Edit3 class="w-4 h-4" /> 编辑
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2 text-error" @click="deleteImage">
          <Trash2 class="w-4 h-4" /> 删除
        </button>
      </template>

      <!-- Selection Menu -->
      <template v-else-if="contextMenuType === 'selection'">
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleBold().run()">
          <Bold class="w-4 h-4" /> 粗体
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleItalic().run()">
          <Italic class="w-4 h-4" /> 斜体
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleUnderline().run()">
          <UnderlineIcon class="w-4 h-4" /> 下划线
        </button>
        <div class="divider my-0"></div>
        <!-- Note: Copy/Cut/Paste usually require browser permissions or native API access which might not work directly in all contexts, but basic execCommand might work -->
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="navigator.clipboard.writeText(editor.state.selection.content().content.textBetween(0, editor.state.selection.content().content.size, '\n'))">
           <ClipboardCopy class="w-4 h-4" /> 复制
        </button>
      </template>

      <!-- Default Menu -->
      <template v-else>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().undo().run()" :disabled="!editor.can().undo()" :class="{'opacity-50': !editor.can().undo()}">
          <Undo class="w-4 h-4" /> 撤销
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().redo().run()" :disabled="!editor.can().redo()" :class="{'opacity-50': !editor.can().redo()}">
          <Redo class="w-4 h-4" /> 重做
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().selectAll().run()">
          <Code2 class="w-4 h-4" /> 全选
        </button>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { Editor, EditorContent } from '@tiptap/vue-3';
import StarterKit from '@tiptap/starter-kit';
import TextAlign from '@tiptap/extension-text-align';
import Underline from '@tiptap/extension-underline';
import Link from '@tiptap/extension-link';
import Image from '@tiptap/extension-image';
import { Table } from '@tiptap/extension-table';
import { TableCell } from '@tiptap/extension-table-cell';
import { TableHeader } from '@tiptap/extension-table-header';
import { TableRow } from '@tiptap/extension-table-row';
import {
  Bold, Italic, Underline as UnderlineIcon, Strikethrough, Code,
  Heading1, Heading2, Heading3,
  List, ListOrdered,
  AlignLeft, AlignCenter, AlignRight,
  Quote, Code2,
  Undo, Redo,
  Table as TableIcon,
  ArrowLeftToLine, ArrowRightToLine, ArrowUpToLine, ArrowDownToLine,
  Trash2, Merge, Split, X,
  Clipboard, Scissors, ClipboardCopy,
  Image as ImageIcon, Paperclip, Edit3
} from 'lucide-vue-next';
import api from '../api/axios';
import eventBus from '../utils/eventBus';
import { getFileUrl } from '../utils/urlHelper';
import { inject } from 'vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const title = ref('');
const saving = ref(false);
const lastSavedContent = ref('');
const notification = inject('notification');
const loadingNote = ref(false);
const noteOwnerId = ref(null);
const noteChannelId = ref(null);
const noteOwnerName = ref('');
const noteChannelName = ref('');
let loadSeq = 0;

// Upload state
const imageInputRef = ref(null);
const fileInputRef = ref(null);
const isUploading = ref(false);
const uploadProgress = ref(0);
const isDragging = ref(false);

// 计算属性：是否可以编辑（未登录时是访客，或登录时不是作者）
const canEdit = computed(() => {
  // 新笔记（没有ID）可以编辑
  if (!route.params.id) return true;
  // 需要用户登录
  if (!authStore.isAuthenticated) return false;
  // 需要是当前用户自己的笔记
  if (!noteOwnerId.value) return false;
  return authStore.user?.id === noteOwnerId.value;
});

// 计算属性：是否是只读模式
const isReadonly = computed(() => !canEdit.value);

// 计算属性：左对齐是否应该高亮（左对齐或没有对齐方式）
const isLeftAlignActive = computed(() => {
  return editor.isActive({ textAlign: 'left' }) ||
         (!editor.isActive({ textAlign: 'center' }) && !editor.isActive({ textAlign: 'right' }));
});

// Context Menu State
const showContextMenu = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuType = ref('default'); // 'default', 'table', 'selection'

const editor = new Editor({
  extensions: [
    StarterKit.configure({
      placeholder: '开始编写你的笔记...',
      // 排除重复的扩展
      heading: {
        levels: [1, 2, 3],
      },
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph', 'image'],
    }),
    Underline,
    Link.configure({
      openOnClick: false,
      link: {
        openOnClick: false,
      },
    }),
    Image.configure({
      inline: true,
      allowBase64: false,
      selectable: true,
      draggable: false,
      HTMLAttributes: {
        class: 'editor-image',
      },
    }),
    Table.configure({
      resizable: true,
    }),
    TableRow,
    TableHeader,
    TableCell,
  ],
  content: '',
  onUpdate: ({ editor }) => {
    // Auto-save on content change
    const currentContent = editor.getHTML();
    if (currentContent !== lastSavedContent.value && route.params.id) {
      autoSave();
    }
  },
});

const handleContextMenu = (e) => {
  // Prevent default browser context menu
  // e.preventDefault(); // Moved to template event modifier

  if (!editor) return;

  // Determine context type
  if (editor.isActive('table')) {
    contextMenuType.value = 'table';
  } else if (editor.isActive('image')) {
    contextMenuType.value = 'image';
  } else if (!editor.state.selection.empty) {
    contextMenuType.value = 'selection';
  } else {
    contextMenuType.value = 'default';
  }

  // Calculate position
  const menuWidth = 200; // Estimated width
  const menuHeight = 300; // Estimated height
  let x = e.clientX;
  let y = e.clientY;

  // Boundary checks
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 10;
  }
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 10;
  }

  contextMenuX.value = x;
  contextMenuY.value = y;
  showContextMenu.value = true;
};

const closeContextMenu = () => {
  showContextMenu.value = false;
};

// Global click listener to close context menu
onMounted(() => {
  window.addEventListener('click', closeContextMenu);
  // Prevent browser from opening files when dragging
  window.addEventListener('dragover', (e) => e.preventDefault());
  window.addEventListener('drop', (e) => e.preventDefault());
  loadNote();
});

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu);
  window.removeEventListener('dragover', (e) => e.preventDefault());
  window.removeEventListener('drop', (e) => e.preventDefault());
  editor.destroy();
});

const autoSave = async () => {
  if (!route.params.id || saving.value) return;

  try {
    // 自动保存只保存内容和标题，不覆盖设置（避免与 Sidebar 的设置冲突）
    const data = {
      title: title.value,
      content: editor.getHTML(),
    };
    await api.put(`/notes/${route.params.id}`, data);
    lastSavedContent.value = editor.getHTML();
    eventBus.emit('note-updated', {
      id: Number(route.params.id),
      title: title.value,
    });
    // Auto-save doesn't need to show a notification to avoid distraction
  } catch (err) {
    console.error('Auto-save failed', err);
  }
};

const showSaveStatus = (message, type) => {
  if (notification) {
    notification.showNotification(message, type);
  }
};

const loadNote = async () => {
  const seq = ++loadSeq;
  loadingNote.value = true;
  try {
    if (route.params.id) {
      const res = await api.get(`/notes/${route.params.id}`);
      if (seq !== loadSeq) return;
      title.value = res.data.title;
      editor.commands.setContent(res.data.content);
      noteOwnerId.value = res.data.owner_id;
      noteChannelId.value = res.data.channel_id;

      // 获取笔记所有者信息
      if (res.data.owner_id) {
        if (res.data.owner && (res.data.owner.nickname || res.data.owner.username)) {
          noteOwnerName.value = res.data.owner.nickname || res.data.owner.username;
        } else {
          // 如果后端没有返回 owner 信息，尝试从 authStore 获取
          if (authStore.user?.id === res.data.owner_id) {
            noteOwnerName.value = authStore.user.nickname || authStore.user.username;
          }
        }
      }

      lastSavedContent.value = editor.getHTML();

      // 如果是频道笔记，获取频道名称
      if (res.data.channel_id) {
        try {
          const channelRes = await api.get(`/channels/${res.data.channel_id}`);
          noteChannelName.value = channelRes.data.channel?.name || '';
        } catch (e) {
          console.error('Failed to fetch channel name:', e);
        }
      }

      // 设置编辑器只读状态
      if (isReadonly.value) {
        editor.setEditable(false);
      } else {
        editor.setEditable(true);
      }
      // 发送笔记信息到 header
      emitNoteInfo();
      return;
    }
    // 新笔记重置作者ID
    noteOwnerId.value = null;
    noteChannelId.value = null;
    noteOwnerName.value = '';
    noteChannelName.value = '';

    if (seq !== loadSeq) return;
    title.value = '';
    editor.commands.setContent('');
    lastSavedContent.value = editor.getHTML();
    // 发送笔记信息到 header
    emitNoteInfo();
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '加载笔记失败', 'error');
    }
  } finally {
    if (seq === loadSeq) loadingNote.value = false;
  }
};

const emitNoteInfo = () => {
  eventBus.emit('note-info-changed', {
    isChannelNote: !!noteChannelId.value,
    ownerId: noteOwnerId.value,
    ownerName: noteOwnerName.value,
    channelName: noteChannelName.value
  });
};

const saveNote = async () => {
  saving.value = true;
  const startTime = Date.now();

  try {
    const data = {
      title: title.value || '无标题',
      content: editor.getHTML(),
    };

    if (route.params.id) {
      await api.put(`/notes/${route.params.id}`, data);
      showSaveStatus('保存成功', 'success');
    } else {
      const res = await api.post('/notes', data);
      showSaveStatus('创建成功', 'success');
      // 跳转到新建的笔记
      router.replace({ name: 'note-editor', params: { id: res.data.id } });
    }

    lastSavedContent.value = editor.getHTML();
    if (route.params.id) {
      eventBus.emit('note-updated', {
        id: Number(route.params.id),
        title: data.title,
      });
    }
  } catch (err) {
    showSaveStatus('保存失败: ' + (err.response?.data?.error || '未知错误'), 'error');
  } finally {
    // 确保至少显示 300ms 的保存状态
    const elapsed = Date.now() - startTime;
    if (elapsed < 300) {
      await new Promise(resolve => setTimeout(resolve, 300 - elapsed));
    }
    saving.value = false;
  }
};

const handleExternalNoteUpdate = (payload) => {
  if (!payload || !route.params.id) return;
  if (Number(route.params.id) !== Number(payload.id)) return;
  // 只更新标题，不覆盖设置（避免与编辑器本地的设置冲突）
  if (typeof payload.title === 'string' && payload.title !== title.value) {
    title.value = payload.title;
  }
};

// Upload functions
const triggerImageUpload = () => {
  imageInputRef.value?.click();
};

const triggerFileUpload = () => {
  fileInputRef.value?.click();
};

const handleImageSelected = async (e) => {
  const file = e.target.files?.[0];
  if (!file) return;
  if (!file.type.startsWith('image/')) {
    if (notification) notification.showNotification('请选择图片文件', 'error');
    return;
  }
  await uploadFile(file, 'image');
  e.target.value = ''; // Reset input
};

const handleFileSelected = async (e) => {
  const file = e.target.files?.[0];
  if (!file) return;
  await uploadFile(file, 'file');
  e.target.value = ''; // Reset input
};

const uploadFile = async (file, type) => {
  if (!route.params.id) {
    if (notification) notification.showNotification('请先保存笔记后再上传文件', 'error');
    return;
  }

  isUploading.value = true;
  uploadProgress.value = 0;

  try {
    const form = new FormData();
    form.append('file', file);
    form.append('type', 'note_attachment');
    form.append('note_id', route.params.id);

    const uploadRes = await api.post('/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        }
      },
    });

    const attachment = uploadRes.data;

    if (type === 'image') {
      insertImage(attachment.file_path);
    } else {
      insertLink(attachment.file_name, attachment.file_path);
    }

    if (notification) notification.showNotification('上传成功', 'success');
  } catch (err) {
    console.error('Upload failed:', err);
    let errorMessage = '上传失败';
    if (err.code === 'ECONNABORTED') errorMessage = '上传超时，请重试';
    else if (err.response?.status === 413) errorMessage = '文件太大';
    else if (err.response?.status === 415) errorMessage = '不支持的文件格式';
    else if (err.response?.data?.error) errorMessage = err.response.data.error;
    if (notification) notification.showNotification(errorMessage, 'error');
  } finally {
    isUploading.value = false;
    uploadProgress.value = 0;
  }
};

const insertImage = (url) => {
  editor.chain().focus().setImage({ src: getFileUrl(url) }).setTextAlign('left').run();
};

const insertLink = (fileName, url) => {
  editor.chain().focus().insertContent(`<a href="${getFileUrl(url)}" target="_blank" class="text-primary hover:underline">${fileName}</a>`).run();
};

// Image context menu actions
const deleteImage = () => {
  if (editor.isActive('image')) {
    editor.chain().focus().deleteSelection().run();
  }
  closeContextMenu();
};

const editImage = () => {
  if (editor.isActive('image')) {
    const { src, alt } = editor.getAttributes('image');
    const newSrc = prompt('输入图片URL:', src);
    if (newSrc && newSrc !== src) {
      editor.chain().focus().updateAttributes('image', { src: newSrc, alt: newSrc }).run();
    }
  }
  closeContextMenu();
};

const insertImageFromUrl = () => {
  const url = prompt('输入图片URL:');
  if (url) {
    insertImage(url);
  }
  closeContextMenu();
};

// Drag and drop handlers
const handleDragOver = (e) => {
  if (!canEdit.value) return;
  e.preventDefault();
  isDragging.value = true;
};

const handleDragLeave = (e) => {
  if (!canEdit.value) return;
  e.preventDefault();
  const rect = e.currentTarget.getBoundingClientRect();
  const x = e.clientX;
  const y = e.clientY;
  if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
    isDragging.value = false;
  }
};

const handleDrop = async (e) => {
  if (!canEdit.value) return;
  e.preventDefault();
  isDragging.value = false;

  const file = e.dataTransfer?.files?.[0];
  if (!file) return;

  const isImage = file.type.startsWith('image/');
  await uploadFile(file, isImage ? 'image' : 'file');
};

const handlePaste = async (e) => {
  if (!canEdit.value) return;
  
  const items = e.clipboardData?.items;
  if (!items) return;
  
  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    
    // 检查是否是文件（图片或其他文件）
    if (item.kind === 'file') {
      e.preventDefault();
      const file = item.getAsFile();
      if (file) {
        const isImage = file.type.startsWith('image/');
        await uploadFile(file, isImage ? 'image' : 'file');
      }
      return;
    }
  }
};

watch(() => route.params.id, () => loadNote());

watch(title, (value) => {
  if (!route.params.id) return;
  eventBus.emit('note-title-updated', { id: Number(route.params.id), title: value });
});

onMounted(() => {
  loadNote();
  eventBus.on('note-updated', handleExternalNoteUpdate);
});

onBeforeUnmount(() => {
  eventBus.off('note-updated', handleExternalNoteUpdate);
});
</script>

<style>
/* Basic table styles for editor */
.ProseMirror table {
  border-collapse: collapse;
  table-layout: fixed;
  width: 100%;
  margin: 1em 0;
  overflow: hidden;
}

.ProseMirror td,
.ProseMirror th {
  min-width: 1em;
  border: 2px solid hsl(var(--bc) / 0.3);
  padding: 3px 5px;
  vertical-align: top;
  box-sizing: border-box;
  position: relative;
}

.ProseMirror th {
  font-weight: bold;
  text-align: left;
  background-color: hsl(var(--bc) / 0.1);
}

.ProseMirror .selectedCell:after {
  z-index: 2;
  position: absolute;
  content: "";
  left: 0; right: 0; top: 0; bottom: 0;
  background: hsl(var(--pf) / 0.4);
  pointer-events: none;
}

.ProseMirror .column-resize-handle {
  position: absolute;
  right: -2px;
  top: 0;
  bottom: -2px;
  width: 4px;
  background-color: hsl(var(--pf));
  pointer-events: none;
}

/* Code block styles - 使用深色背景和浅色文字 */
.ProseMirror pre {
  background-color: #1a1a1a;
  color: #e6edf3;
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  padding: 1rem;
  border-radius: 0.5rem;
  margin: 1rem 0;
  overflow-x: auto;
  border: 1px solid hsl(var(--bc) / 0.3);
}

/* 强制代码块内文字颜色 */
.ProseMirror pre code {
  color: #e6edf3 !important;
  padding: 0;
  background: none !important;
  font-size: 0.875rem;
}

/* Inline code styles */
.ProseMirror code {
  background-color: hsl(var(--bc) / 0.15);
  color: hsl(var(--pc));
  padding: 0.2em 0.4em;
  border-radius: 0.375rem;
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 0.875em;
}

/* Dragging state */
.dragging {
  border-color: var(--fallback-p, oklch(var(--p))) !important;
  background-color: oklch(var(--b3)) !important;
  transition: all 0.2s ease;
}

/* 图片样式 - 默认较小尺寸，可调整大小 */
.ProseMirror img {
  width: 20%;
  height: auto;
  display: inline-block;
  border-radius: 0.375rem;
  margin: 0.5em 0;
  resize: both;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s ease;
  min-width: 50px;
  max-width: 100%;
  user-select: none;
  cursor: pointer;
}

.ProseMirror img:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.ProseMirror img.ProseMirror-selectednode {
  outline: 2px solid hsl(var(--pf));
  resize: both;
}

.ProseMirror img:focus {
  outline: 2px solid hsl(var(--pf));
}
</style>
