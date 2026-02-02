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
        <!-- 协同编辑状态指示器 (仅频道笔记显示) -->
        <div v-if="isCollabEnabled" class="flex items-center gap-2 px-3 py-2 rounded-lg bg-base-200">
          <!-- 状态点 -->
          <div class="w-2 h-2 rounded-full" 
               :class="{
                 'bg-success': isCollabConnected && isCollabSynced,
                 'bg-warning': isCollabConnected && !isCollabSynced,
                 'bg-error': !isCollabConnected
               }"></div>
          <!-- 状态文字 (仅桌面端显示) -->
          <span class="text-xs font-medium hidden sm:inline">
            {{ isCollabConnected && isCollabSynced ? '协同工作中' : 
               isCollabConnected ? '同步中' : '未连接' }}
          </span>
          <!-- 在线用户头像 -->
          <div v-if="isCollabConnected && isCollabSynced" class="flex -space-x-2 sm:ml-1">
            <!-- 显示自己 -->
            <div class="w-6 h-6 rounded-full flex items-center justify-center text-white text-xs font-bold border-2 border-base-100 bg-primary"
                 :title="`${authStore.user?.username || authStore.user?.nickname || '我'} (你)`">
              {{ (authStore.user?.username || authStore.user?.nickname || '我').charAt(0).toUpperCase() }}
            </div>
            <!-- 显示其他用户 -->
            <div v-for="[clientId, cursor] in Array.from(remoteCursors.entries()).slice(0, 3)" 
                 :key="clientId"
                 class="w-6 h-6 rounded-full flex items-center justify-center text-white text-xs font-bold border-2 border-base-100"
                 :style="{ backgroundColor: cursor.color }"
                 :title="cursor.username">
              {{ cursor.username.charAt(0).toUpperCase() }}
            </div>
            <div v-if="remoteCursors.size > 3"
                 class="w-6 h-6 rounded-full flex items-center justify-center bg-base-300 text-xs font-bold border-2 border-base-100"
                 :title="`还有 ${remoteCursors.size - 3} 位用户`">
              +{{ remoteCursors.size - 3 }}
            </div>
          </div>
        </div>
        
        <div class="dropdown dropdown-end">
          <button tabindex="0" class="btn btn-neutral" :disabled="summarizing || polishing">
            <Sparkles class="w-4 h-4 sm:mr-2" />
            <span class="hidden sm:inline">AI</span>
          </button>
          <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-100 rounded-box w-40">
            <li>
              <a @click="summarizeNote" :class="{ 'opacity-50': summarizing }">
                <FileText class="w-4 h-4" />
                {{ summarizing ? '总结中...' : 'AI总结' }}
              </a>
            </li>
            <li>
              <a @click="polishNote" :class="{ 'opacity-50': polishing }">
                <Wand2 class="w-4 h-4" />
                {{ polishing ? '润色中...' : 'AI润色' }}
              </a>
            </li>
          </ul>
        </div>
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
      <!-- 第一行：最常用的格式化工具 -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleBold().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('bold'), 'bg-base-100 hover:bg-base-300': !editor.isActive('bold') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="粗体">
          <Bold class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleItalic().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('italic'), 'bg-base-100 hover:bg-base-300': !editor.isActive('italic') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="斜体">
          <Italic class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleUnderline().run())" 
                class="btn btn-xs btn-square border transition-all duration-200 hidden sm:inline-flex"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('underline'), 'bg-base-100 hover:bg-base-300': !editor.isActive('underline') }"
                title="下划线">
          <UnderlineIcon class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleStrike().run())" 
                class="btn btn-xs btn-square border transition-all duration-200 hidden sm:inline-flex"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('strike'), 'bg-base-100 hover:bg-base-300': !editor.isActive('strike') }"
                title="删除线">
          <Strikethrough class="w-4 h-4" />
        </button>
      </div>

      <!-- 标题 -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleHeading({ level: 1 }).run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 1 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 1 }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="标题1">
          <Heading1 class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleHeading({ level: 2 }).run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 2 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 2 }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="标题2">
          <Heading2 class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleHeading({ level: 3 }).run())" 
                class="btn btn-xs btn-square border transition-all duration-200 hidden sm:inline-flex"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('heading', { level: 3 }), 'bg-base-100 hover:bg-base-300': !editor.isActive('heading', { level: 3 }) }"
                title="标题3">
          <Heading3 class="w-4 h-4" />
        </button>
      </div>

      <!-- 列表 -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleBulletList().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('bulletList'), 'bg-base-100 hover:bg-base-300': !editor.isActive('bulletList') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="无序列表">
          <List class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleOrderedList().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('orderedList'), 'bg-base-100 hover:bg-base-300': !editor.isActive('orderedList') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="有序列表">
          <ListOrdered class="w-4 h-4" />
        </button>
      </div>

      <!-- 上传（移动端优先显示） -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
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
        <button @click="triggerFileUpload" :disabled="isUploading" class="btn btn-xs btn-square border bg-base-100 hover:bg-base-300 transition-all duration-200 hidden sm:inline-flex" title="插入附件">
          <Paperclip class="w-4 h-4" />
        </button>
      </div>

      <!-- 历史操作 -->
      <div class="flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.undo().run())"
                :disabled="!editor.can().undo()"
                :class="{ 'bg-base-100 hover:bg-base-300': editor.can().undo(), 'bg-base-200 opacity-50 cursor-not-allowed': !editor.can().undo() }"
                class="btn btn-xs btn-square border transition-all duration-200" title="撤销">
          <Undo class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.redo().run())"
                :disabled="!editor.can().redo()"
                :class="{ 'bg-base-100 hover:bg-base-300': editor.can().redo(), 'bg-base-200 opacity-50 cursor-not-allowed': !editor.can().redo() }"
                class="btn btn-xs btn-square border transition-all duration-200" title="重做">
          <Redo class="w-4 h-4" />
        </button>
      </div>

      <!-- 桌面端额外功能 -->
      <div class="hidden sm:flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleCode().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('code'), 'bg-base-100 hover:bg-base-300': !editor.isActive('code') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="代码">
          <Code class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleBlockquote().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('blockquote'), 'bg-base-100 hover:bg-base-300': !editor.isActive('blockquote') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="引用">
          <Quote class="w-4 h-4" />
        </button>
      </div>

      <!-- 对齐（桌面端显示） -->
      <div class="hidden sm:flex items-center gap-1 border-r border-base-300 pr-2">
        <button @mousedown="handleToolbarAction($event, chain => chain.setTextAlign('left').run())"
                :class="{ 'bg-neutral text-neutral-content border-neutral': isLeftAlignActive, 'bg-base-100 hover:bg-base-300': !isLeftAlignActive }"
                class="btn btn-xs btn-square border transition-all duration-200" title="左对齐">
          <AlignLeft class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.setTextAlign('center').run())"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive({ textAlign: 'center' }), 'bg-base-100 hover:bg-base-300': !editor.isActive({ textAlign: 'center' }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="居中对齐">
          <AlignCenter class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.setTextAlign('right').run())"
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive({ textAlign: 'right' }), 'bg-base-100 hover:bg-base-300': !editor.isActive({ textAlign: 'right' }) }"
                class="btn btn-xs btn-square border transition-all duration-200" title="右对齐">
          <AlignRight class="w-4 h-4" />
        </button>
      </div>

      <!-- 表格和代码块（桌面端显示） -->
      <div class="hidden sm:flex items-center gap-1">
        <button @mousedown="handleToolbarAction($event, chain => chain.insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run())"
                class="btn btn-xs btn-square border transition-all duration-200" title="插入表格">
          <TableIcon class="w-4 h-4" />
        </button>
        <button @mousedown="handleToolbarAction($event, chain => chain.toggleCodeBlock().run())" 
                :class="{ 'bg-neutral text-neutral-content border-neutral': editor.isActive('codeBlock'), 'bg-base-100 hover:bg-base-300': !editor.isActive('codeBlock') }"
                class="btn btn-xs btn-square border transition-all duration-200" title="代码块">
          <Code2 class="w-4 h-4" />
        </button>
      </div>

      <!-- Upload Progress -->
      <div v-if="isUploading" class="flex items-center gap-2 ml-2">
        <span class="text-xs text-primary font-medium">上传中 {{ uploadProgress }}%</span>
        <div class="w-24">
          <progress class="progress progress-primary w-full h-2" :value="uploadProgress" max="100"></progress>
        </div>
      </div>

      <!-- Line Spacing -->
      <div class="flex items-center gap-2 ml-2 border-l border-base-300 pl-2">
        <span class="text-xs text-base-content/70 cursor-pointer hover:text-base-content" @click="resetLineSpacing" title="恢复默认行距">行距</span>
        <input
          type="range"
          v-model.number="lineSpacing"
          min="1"
          max="3"
          step="0.1"
          class="range range-xs w-20"
          title="调整行间距"
        />
        <span class="text-xs text-base-content/70 w-8">{{ lineSpacing }}</span>
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
          <UploadCloud class="w-16 h-16 mx-auto mb-2 text-primary opacity-60" />
          <p class="text-lg font-medium text-primary">拖放文件到此处</p>
          <p class="text-sm text-base-content/60 mt-1">图片将直接插入，其他文件将作为附件</p>
        </div>
      </div>
      <editor-content 
        :editor="editor" 
        class="flex-1 overflow-auto pt-0 px-4 pb-4 prose max-w-none prose-neutral"
        :style="{ '--line-height': lineSpacing }"
        @paste="handlePaste"
      />
    </div>

    <!-- Context Menu -->
    <div v-if="showContextMenu && canEdit" 
         class="fixed z-50 bg-base-100 shadow-xl rounded-lg border border-base-200 py-1 min-w-[180px] text-sm"
         :style="{ top: `${contextMenuY}px`, left: `${contextMenuX}px` }"
         @click.stop>
      
      <!-- Table Menu -->
      <template v-if="contextMenuType === 'table'">
        <div class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">表格操作</div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addColumnBefore().run(); closeContextMenu()">
          <ArrowLeftToLine class="w-4 h-4" /> 前插列
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addColumnAfter().run(); closeContextMenu()">
          <ArrowRightToLine class="w-4 h-4" /> 后插列
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addRowBefore().run(); closeContextMenu()">
          <ArrowUpToLine class="w-4 h-4" /> 前插行
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().addRowAfter().run(); closeContextMenu()">
          <ArrowDownToLine class="w-4 h-4" /> 后插行
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().mergeCells().run(); closeContextMenu()">
          <Merge class="w-4 h-4" /> 合并单元格
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().splitCell().run(); closeContextMenu()">
          <Split class="w-4 h-4" /> 拆分单元格
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2 text-base-content" @click="editor.chain().focus().deleteTable().run(); closeContextMenu()">
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
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleBold().run(); closeContextMenu()">
          <Bold class="w-4 h-4" /> 粗体
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleItalic().run(); closeContextMenu()">
          <Italic class="w-4 h-4" /> 斜体
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().toggleUnderline().run(); closeContextMenu()">
          <UnderlineIcon class="w-4 h-4" /> 下划线
        </button>
        <div class="divider my-0"></div>
        <!-- Note: Copy/Cut/Paste usually require browser permissions or native API access which might not work directly in all contexts, but basic execCommand might work -->
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="navigator.clipboard.writeText(editor.state.selection.content().content.textBetween(0, editor.state.selection.content().content.size, '\n')); closeContextMenu()">
           <ClipboardCopy class="w-4 h-4" /> 复制
        </button>
      </template>

      <!-- Default Menu -->
      <template v-else>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().undo().run(); closeContextMenu()" :disabled="!editor.can().undo()" :class="{'opacity-50': !editor.can().undo()}">
          <Undo class="w-4 h-4" /> 撤销
        </button>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().redo().run(); closeContextMenu()" :disabled="!editor.can().redo()" :class="{'opacity-50': !editor.can().redo()}">
          <Redo class="w-4 h-4" /> 重做
        </button>
        <div class="divider my-0"></div>
        <button class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2" @click="editor.chain().focus().selectAll().run(); closeContextMenu()">
          <Code2 class="w-4 h-4" /> 全选
        </button>
      </template>
    </div>

    <!-- AI Summary Modal -->
    <dialog id="ai_summary_modal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">AI 总结</h3>
        <div v-if="summary" class="prose prose-sm max-w-none" v-html="formatSummary(summary)"></div>
        <div v-else-if="summarizing" class="flex items-center justify-center py-8">
          <span class="loading loading-spinner loading-lg text-neutral"></span>
          <span class="ml-3 text-base-content/70">AI 正在总结中...</span>
        </div>
        <div class="modal-action">
          <form method="dialog">
            <button class="btn">关闭</button>
          </form>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button>close</button>
      </form>
    </dialog>

    <!-- AI Polish Modal -->
    <dialog id="ai_polish_modal" class="modal modal-bottom sm:modal-middle">
      <div class="modal-box max-w-3xl">
        <h3 class="font-bold text-lg mb-4">AI 润色</h3>
        <div v-if="polishedContent" class="space-y-4">
          <div>
            <label class="label">
              <span class="label-text font-medium">润色后的内容</span>
            </label>
            <div class="prose prose-sm max-w-none bg-base-200 p-4 rounded-lg max-h-96 overflow-auto" v-html="formatPolished(polishedContent)"></div>
          </div>
        </div>
        <div v-else-if="polishing" class="flex items-center justify-center py-8">
          <span class="loading loading-spinner loading-lg text-neutral"></span>
          <span class="ml-3 text-base-content/70">AI 正在润色中...</span>
        </div>
        <div class="modal-action" v-if="polishedContent">
          <form method="dialog">
            <button class="btn">取消</button>
          </form>
          <button @click="applyPolishedContent" class="btn btn-neutral">应用润色</button>
        </div>
        <div class="modal-action" v-else>
          <form method="dialog">
            <button class="btn">取消</button>
          </form>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { Editor, EditorContent, Extension } from '@tiptap/vue-3';
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
  Image as ImageIcon, Paperclip, Edit3, UploadCloud,
  Sparkles, Wand2, FileText
} from 'lucide-vue-next';
import api from '../api/axios';
import eventBus from '../utils/eventBus';
import { getFileUrl } from '../utils/urlHelper';
import { inject } from 'vue';
import { marked } from 'marked';
import { YjsCollabClient } from '../utils/yjs-collab';

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

// AI Summary state
const summarizing = ref(false);
const summary = ref('');

// AI Polish state
const polishing = ref(false);
const polishedContent = ref('');

// Editor state
const lineSpacing = ref(1.5);

// 协同编辑状态
const collabClient = ref(null);
const isCollabEnabled = computed(() => !!noteChannelId.value && authStore.isAuthenticated);
const isCollabConnected = ref(false);
const isCollabSynced = ref(false);
const isChannelMember = ref(false); // 是否是频道成员
const remoteCursors = ref(new Map()); // 存储其他用户的光标位置

// 配置 marked 以支持 GFM 和其他功能
marked.setOptions({
  breaks: true,  // 支持换行符转换为 <br>
  gfm: true,     // 启用 GitHub Flavored Markdown
});

// 计算属性：是否可以编辑
const canEdit = computed(() => {
  // 新笔记（没有ID）可以编辑
  if (!route.params.id) return true;
  // 需要用户登录
  if (!authStore.isAuthenticated) return false;
  // 如果没有加载完成，暂时不允许编辑
  if (!noteOwnerId.value) return false;
  
  // 如果是频道笔记，检查是否是频道成员
  if (noteChannelId.value) {
    return isChannelMember.value;
  }
  
  // 个人笔记：需要是当前用户自己的笔记
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

// 协同编辑同步状态
let isUpdatingFromYjs = false;
let syncTimeout = null;

// 检测是否是移动设备
const isMobile = () => {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || window.innerWidth < 768;
};

// 工具栏按钮点击处理（移动端智能聚焦）
const handleToolbarAction = (event, action) => {
  // 阻止按钮获得焦点，避免编辑器失焦
  event.preventDefault();
  
  if (isMobile()) {
    // 移动端：检查编辑器是否已经有焦点
    const editorHasFocus = editor.view.hasFocus();
    
    if (editorHasFocus) {
      // 如果编辑器已经有焦点（用户正在编辑），正常调用 focus() 保持焦点
      action(editor.chain().focus());
    } else {
      // 如果编辑器没有焦点，不调用 focus()，避免唤起输入法
      action(editor.chain());
    }
  } else {
    // 桌面端：正常调用 focus()
    action(editor.chain().focus());
  }
};

// 自定义扩展：按 Tab 键插入制表符
const TabIndent = Extension.create({
  name: 'tabIndent',
  addKeyboardShortcuts() {
    return {
      Tab: () => {
        // 插入制表符
        return this.editor.commands.insertContent('\t');
      },
    };
  },
});

const editor = new Editor({
  extensions: [
    StarterKit.configure({
      placeholder: '开始编写你的笔记...',
      // 排除重复的扩展
      heading: {
        levels: [1, 2, 3],
      },
      // StarterKit 包含了 Link 和 Underline，所以需要排除
      link: false,
      underline: false,
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
    TabIndent,
  ],
  content: '',
  onUpdate: ({ editor }) => {
    // 如果正在从 Yjs 更新，不要触发同步
    if (isUpdatingFromYjs) {
      return;
    }
    
    // 如果启用了协同编辑，同步到 Yjs
    if (isCollabEnabled.value && collabClient.value?.synced) {
      syncEditorToYjs();
    }
    
    // Auto-save on content change
    const currentContent = editor.getHTML();
    if (currentContent !== lastSavedContent.value && route.params.id) {
      autoSave();
    }
  },
  onSelectionUpdate: () => {
    // 发送光标位置
    handleSelectionUpdate();
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
  // Global click listener to close context menu
  window.addEventListener('click', closeContextMenu);
  loadNote();
  eventBus.on('note-updated', handleExternalNoteUpdate);

  // 添加键盘快捷键监听
  window.addEventListener('keydown', handleKeyboardShortcuts);
  // Prevent browser from opening files when dragging
  window.addEventListener('dragover', (e) => e.preventDefault());
  window.addEventListener('drop', (e) => e.preventDefault());
  
  // 监听编辑器滚动，更新光标位置
  nextTick(() => {
    const editorElement = document.querySelector('.ProseMirror');
    if (editorElement) {
      editorElement.addEventListener('scroll', () => {
        if (remoteCursors.value.size > 0) {
          updateCursorDecorations();
        }
      });
    }
  });
});

const handleKeyboardShortcuts = (e) => {
  // Ctrl+S 保存笔记
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault();
    if (canEdit.value) {
      saveNote();
    }
  }
};

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu);
  window.removeEventListener('dragover', (e) => e.preventDefault());
  window.removeEventListener('drop', (e) => e.preventDefault());
  window.removeEventListener('keydown', handleKeyboardShortcuts);
  eventBus.off('note-updated', handleExternalNoteUpdate);
  editor.destroy();
});

const autoSave = async () => {
  if (!route.params.id || saving.value) return;

  try {
    // 自动保存只保存内容和标题，不覆盖设置（避免与 Sidebar 的设置冲突）
    const data = {
      title: title.value,
      content: editor.getHTML(),
      line_spacing: lineSpacing.value,
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

const resetLineSpacing = () => {
  lineSpacing.value = 1.5;
  if (notification) notification.showNotification('已恢复默认行间距', 'success');
};

const loadNote = async () => {
  const seq = ++loadSeq;
  loadingNote.value = true;
  
  // 断开之前的协同编辑连接
  disconnectCollab();
  
  // 重置频道成员状态
  isChannelMember.value = false;
  
  try {
    if (route.params.id) {
      const res = await api.get(`/notes/${route.params.id}`);
      if (seq !== loadSeq) return;
      
      title.value = res.data.title;
      editor.commands.setContent(res.data.content);
      noteOwnerId.value = res.data.owner_id;
      noteChannelId.value = res.data.channel_id;
      lineSpacing.value = res.data.line_spacing || 1.5;

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

      // 如果是频道笔记，获取频道信息并检查成员权限
      if (res.data.channel_id) {
        try {
          const channelRes = await api.get(`/channels/${res.data.channel_id}`);
          noteChannelName.value = channelRes.data.channel?.name || '';
          
          // 检查当前用户是否是频道成员
          // 注意：成员列表在 channelRes.data.members，不是 channelRes.data.channel.members
          if (authStore.isAuthenticated && channelRes.data.members) {
            const currentUserId = authStore.user?.id;
            isChannelMember.value = channelRes.data.members.some(
              member => member.user_id === currentUserId && member.status === 'active'
            );
          }
        } catch (e) {
          console.error('Failed to fetch channel info:', e);
        }
      }

      // 设置编辑器只读状态
      if (isReadonly.value) {
        editor.setEditable(false);
      } else {
        editor.setEditable(true);
      }
      
      // 如果是频道笔记且用户已登录，启用协同编辑
      if (isCollabEnabled.value) {
        await nextTick();
        initCollab();
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
    isChannelMember.value = false;

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
      line_spacing: lineSpacing.value,
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

const summarizeNote = async () => {
  if (!authStore.isAuthenticated) {
    if (notification) notification.showNotification('请先登录', 'error');
    return;
  }

  const content = editor.getHTML();
  const currentTitle = title.value || '无标题';

  // 提取纯文本内容，移除图片等媒体元素
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = content;

  // 移除所有图片、视频、音频等媒体元素
  const mediaElements = tempDiv.querySelectorAll('img, video, audio, iframe, object, embed');
  mediaElements.forEach(el => el.remove());

  // 提取纯文本内容
  const textContent = tempDiv.textContent || tempDiv.innerText || '';

  if (textContent.trim().length < 10) {
    if (notification) notification.showNotification('笔记内容太少，无法总结', 'error');
    return;
  }

  summarizing.value = true;
  summary.value = '';

  // 打开模态框
  const modal = document.getElementById('ai_summary_modal');
  if (modal) modal.showModal();

  try {
    const res = await api.post('/ai/summarize', {
      title: currentTitle,
      content: textContent,
    });
    summary.value = res.data.summary;
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || 'AI 总结失败', 'error');
    }
    if (modal) modal.close();
  } finally {
    summarizing.value = false;
  }
};

const formatSummary = (text) => {
  // 使用 marked 解析 Markdown
  return marked.parse(text);
};

const polishNote = async () => {
  if (!authStore.isAuthenticated) {
    if (notification) notification.showNotification('请先登录', 'error');
    return;
  }

  const content = editor.getHTML();
  const currentTitle = title.value || '无标题';

  // 提取纯文本内容，移除图片等媒体元素
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = content;

  // 移除所有图片、视频、音频等媒体元素
  const mediaElements = tempDiv.querySelectorAll('img, video, audio, iframe, object, embed');
  mediaElements.forEach(el => el.remove());

  // 提取纯文本内容
  const textContent = tempDiv.textContent || tempDiv.innerText || '';

  if (textContent.trim().length < 10) {
    if (notification) notification.showNotification('笔记内容太少，无法润色', 'error');
    return;
  }

  polishing.value = true;
  polishedContent.value = '';

  // 打开模态框
  const modal = document.getElementById('ai_polish_modal');
  if (modal) modal.showModal();

  try {
    const res = await api.post('/ai/polish', {
      title: currentTitle,
      content: textContent,
    });
    polishedContent.value = res.data.polished;
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || 'AI 润色失败', 'error');
    }
    if (modal) modal.close();
  } finally {
    polishing.value = false;
  }
};

const applyPolishedContent = () => {
  if (polishedContent.value) {
    // 将 Markdown 转换为 HTML 后应用，保留格式
    const htmlContent = formatPolished(polishedContent.value);
    editor.commands.setContent(htmlContent);
    const modal = document.getElementById('ai_polish_modal');
    if (modal) modal.close();
    if (notification) notification.showNotification('已应用润色内容', 'success');
  }
};

const formatPolished = (text) => {
  // 使用 marked 解析 Markdown
  return marked.parse(text);
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

// 协同编辑相关函数
const initCollab = () => {
  if (!route.params.id || !authStore.user) return;
  
  try {
    collabClient.value = new YjsCollabClient(
      Number(route.params.id),
      authStore.user.id,
      authStore.user.username || authStore.user.nickname || `User${authStore.user.id}`
    );

    // 监听 Yjs 更新
    collabClient.value.onUpdate = (content) => {
      syncYjsToEditor(content);
    };

    // 监听同步完成
    collabClient.value.onSync = (content) => {
      isCollabSynced.value = true;
      
      // 同步完成后，如果服务器有内容，使用服务器内容
      // 如果服务器没有内容，发送本地内容
      if (content && content.trim()) {
        syncYjsToEditor(content);
      } else {
        const localHtml = editor.getHTML();
        if (localHtml && localHtml !== '<p></p>') {
          collabClient.value.setText(localHtml);
        }
      }
    };

    // 监听感知状态（光标位置）
    collabClient.value.onAwareness = (awareness) => {
      handleRemoteCursor(awareness);
    };

    // 监听用户离开
    collabClient.value.onUserLeft = (user) => {
      remoteCursors.value.delete(user.clientId);
      // 强制触发 Vue 响应式更新
      remoteCursors.value = new Map(remoteCursors.value);
      updateCursorDecorations();
    };

    // 监听连接状态
    collabClient.value.onConnect = () => {
      isCollabConnected.value = true;
    };

    collabClient.value.onDisconnect = () => {
      isCollabConnected.value = false;
      isCollabSynced.value = false;
      // 清除所有远程光标
      remoteCursors.value.clear();
      // 强制触发 Vue 响应式更新
      remoteCursors.value = new Map(remoteCursors.value);
      updateCursorDecorations();
    };

    // 连接到服务器
    collabClient.value.connect();
  } catch (err) {
    console.error('初始化协同编辑失败:', err);
    if (notification) {
      notification.showNotification('协同编辑连接失败', 'error');
    }
  }
};

const disconnectCollab = () => {
  if (collabClient.value) {
    collabClient.value.destroy();
    collabClient.value = null;
    isCollabConnected.value = false;
    isCollabSynced.value = false;
  }
};

const syncYjsToEditor = (content) => {
  if (isUpdatingFromYjs) {
    return;
  }
  
  // 设置标志，防止循环
  isUpdatingFromYjs = true;
  
  // 清除之前的超时
  if (syncTimeout) {
    clearTimeout(syncTimeout);
  }
  
  try {
    // 保存当前光标位置
    const { from } = editor.state.selection;
    
    // 获取当前编辑器内容
    const currentContent = editor.getHTML();
    
    // 只有当内容真的不同时才更新
    if (currentContent !== content) {
      // 直接使用 HTML 内容（Yjs 存储的是 HTML）
      editor.commands.setContent(content || '<p></p>', false);
      
      // 恢复光标位置
      try {
        const newLength = editor.state.doc.content.size;
        const safePosition = Math.min(from, Math.max(0, newLength - 1));
        editor.commands.setTextSelection(safePosition);
      } catch (e) {
        // 忽略光标位置错误
      }
    }
    
    // 更新远程光标显示
    nextTick(() => {
      updateCursorDecorations();
    });
  } finally {
    // 延迟重置标志，确保 onUpdate 事件处理完成
    syncTimeout = setTimeout(() => {
      isUpdatingFromYjs = false;
      syncTimeout = null;
    }, 100);
  }
};

const syncEditorToYjs = () => {
  if (isUpdatingFromYjs || !collabClient.value?.synced) {
    return;
  }
  
  try {
    const editorHtml = editor.getHTML();
    const yjsText = collabClient.value.getText();
    
    // 只有当内容真的不同时才同步
    if (editorHtml !== yjsText) {
      collabClient.value.setText(editorHtml);
    }
  } catch (e) {
    console.error('同步到 Yjs 失败:', e);
  }
};

// 处理远程光标
const handleRemoteCursor = (awareness) => {
  if (!awareness.cursor) {
    // 移除光标
    remoteCursors.value.delete(awareness.clientId);
  } else {
    // 更新光标
    remoteCursors.value.set(awareness.clientId, {
      userId: awareness.userId,
      username: awareness.username,
      color: awareness.color,
      from: awareness.cursor.from,
      to: awareness.cursor.to,
      timestamp: awareness.timestamp
    });
  }
  
  // 清理过期的光标（超过10秒）
  const now = Date.now() / 1000;
  for (const [clientId, cursor] of remoteCursors.value.entries()) {
    if (now - cursor.timestamp > 10) {
      remoteCursors.value.delete(clientId);
    }
  }
  
  // 强制触发 Vue 响应式更新
  // 创建一个新的 Map 来触发响应
  remoteCursors.value = new Map(remoteCursors.value);
};

// 更新光标装饰
const updateCursorDecorations = () => {
  // 移除所有现有的光标标签
  document.querySelectorAll('.remote-cursor').forEach(el => el.remove());
  
  if (!editor || remoteCursors.value.size === 0) {
    return;
  }
  
  try {
    const editorElement = document.querySelector('.ProseMirror');
    if (!editorElement) return;
    
    const docSize = editor.state.doc.content.size;
    
    // 为每个远程光标创建装饰
    for (const [clientId, cursor] of remoteCursors.value.entries()) {
      try {
        // 确保光标位置在有效范围内
        const safeFrom = Math.min(Math.max(0, cursor.from), docSize - 1);
        
        // 获取光标位置的 DOM 坐标
        const coords = editor.view.coordsAtPos(safeFrom);
        if (!coords) continue;
        
        // 获取编辑器容器的位置
        const editorRect = editorElement.getBoundingClientRect();
        
        // 计算相对位置
        const left = coords.left - editorRect.left + editorElement.scrollLeft;
        const top = coords.top - editorRect.top + editorElement.scrollTop;
        
        // 创建光标元素
        const cursorEl = document.createElement('div');
        cursorEl.className = 'remote-cursor';
        cursorEl.style.position = 'absolute';
        cursorEl.style.left = `${left}px`;
        cursorEl.style.top = `${top}px`;
        cursorEl.style.width = '2px';
        cursorEl.style.height = '20px';
        cursorEl.style.backgroundColor = cursor.color;
        cursorEl.style.zIndex = '1000';
        cursorEl.style.pointerEvents = 'none';
        
        // 创建标签元素
        const labelEl = document.createElement('div');
        labelEl.className = 'remote-cursor-label';
        labelEl.textContent = cursor.username;
        labelEl.style.position = 'absolute';
        labelEl.style.left = '0';
        labelEl.style.top = '-20px';
        labelEl.style.backgroundColor = cursor.color;
        labelEl.style.color = 'white';
        labelEl.style.padding = '2px 6px';
        labelEl.style.borderRadius = '3px';
        labelEl.style.fontSize = '12px';
        labelEl.style.whiteSpace = 'nowrap';
        labelEl.style.pointerEvents = 'none';
        
        cursorEl.appendChild(labelEl);
        editorElement.appendChild(cursorEl);
      } catch (e) {
        // 忽略单个光标的错误
        console.warn('渲染光标失败:', e);
      }
    }
  } catch (e) {
    console.error('更新光标装饰失败:', e);
  }
};

// 发送本地光标位置
const sendLocalCursor = () => {
  if (!collabClient.value?.synced) {
    return;
  }
  
  try {
    const { from, to } = editor.state.selection;
    collabClient.value.sendCursorPosition(from, to);
  } catch (e) {
    // 忽略错误
  }
};

// 监听编辑器选择变化
let cursorUpdateTimeout = null;
const handleSelectionUpdate = () => {
  if (!isCollabEnabled.value || !isCollabSynced.value) {
    return;
  }
  
  // 防抖：避免频繁发送光标位置
  if (cursorUpdateTimeout) {
    clearTimeout(cursorUpdateTimeout);
  }
  
  cursorUpdateTimeout = setTimeout(() => {
    sendLocalCursor();
  }, 200);
};

onMounted(() => {
  loadNote();
  eventBus.on('note-updated', handleExternalNoteUpdate);
});

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu);
  window.removeEventListener('dragover', (e) => e.preventDefault());
  window.removeEventListener('drop', (e) => e.preventDefault());
  window.removeEventListener('keydown', handleKeyboardShortcuts);
  eventBus.off('note-updated', handleExternalNoteUpdate);
  
  // 清除同步超时
  if (syncTimeout) {
    clearTimeout(syncTimeout);
  }
  
  // 清除光标更新超时
  if (cursorUpdateTimeout) {
    clearTimeout(cursorUpdateTimeout);
  }
  
  disconnectCollab();
  editor.destroy();
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
  padding: 0.75rem;
  border-radius: 0.375rem;
  margin: 0.5em 0;
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

/* 段落行间距 - 使用 CSS 变量 */
.ProseMirror p {
  margin: 0.5em 0;
  line-height: var(--line-height, 1.5);
}

/* 其他元素的行间距 */
.ProseMirror h1,
.ProseMirror h2,
.ProseMirror h3,
.ProseMirror h4,
.ProseMirror h5,
.ProseMirror h6,
.ProseMirror ul,
.ProseMirror ol,
.ProseMirror li,
.ProseMirror blockquote {
  line-height: var(--line-height, 1.5);
}

/* 引用样式 */
.ProseMirror blockquote {
  border-left: 4px solid hsl(var(--pf));
  padding-left: 1em;
  margin: 0.5em 0;
  color: hsl(var(--bc) / 0.7);
  font-style: italic;
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

/* 制表符样式 - 确保存储和显示 */
.ProseMirror {
  white-space: pre-wrap;
  tab-size: 4;
  -moz-tab-size: 4;
  -o-tab-size: 4;
}

/* Markdown 渲染样式 */
.prose h1, .prose h2, .prose h3, .prose h4, .prose h5, .prose h6 {
  font-weight: 700;
  line-height: 1.25;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}

/* Ensure underline and strikethrough are visible */
.ProseMirror u {
  text-decoration: underline;
}

.ProseMirror s {
  text-decoration: line-through;
}

.prose h1 {
  font-size: 2em;
  border-bottom: 1px solid hsl(var(--bc) / 0.2);
  padding-bottom: 0.3em;
}

.prose h2 {
  font-size: 1.5em;
  border-bottom: 1px solid hsl(var(--bc) / 0.15);
  padding-bottom: 0.3em;
}

/* 远程光标样式 */
.remote-cursor {
  position: absolute;
  pointer-events: none;
  z-index: 1000;
}

.remote-cursor-label {
  position: absolute;
  white-space: nowrap;
  font-size: 12px;
  font-weight: 500;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

@keyframes blink {
  0%, 49% {
    opacity: 1;
  }
  50%, 100% {
    opacity: 0;
  }
}

/* 编辑器容器需要相对定位 */
.ProseMirror {
  position: relative;
}

.prose h3 {
  font-size: 1.25em;
}

.prose ul, .prose ol {
  padding-left: 1.5em;
  margin: 1em 0;
}

.prose ul {
  list-style-type: disc;
}

.prose ol {
  list-style-type: decimal;
}

.prose li {
  margin: 0.25em 0;
}

.prose ul ul, .prose ol ul, .prose ul ol, .prose ol ol {
  margin: 0.25em 0;
}

.prose ul ul {
  list-style-type: circle;
}

.prose ul ul ul {
  list-style-type: square;
}

.prose strong {
  font-weight: 700;
}

.prose em {
  font-style: italic;
}

.prose blockquote {
  border-left: 4px solid hsl(var(--pf));
  padding-left: 1em;
  margin: 0.5em 0;
  color: hsl(var(--bc) / 0.7);
  font-style: italic;
}

.prose hr {
  border: none;
  border-top: 2px solid hsl(var(--bc) / 0.2);
  margin: 2em 0;
}

.prose pre {
  background-color: #1a1a1a;
  color: #e6edf3;
  padding: 0.75rem;
  border-radius: 0.375rem;
  overflow-x: auto;
  margin: 0.5em 0;
}

.prose code {
  background-color: hsl(var(--bc) / 0.15);
  color: hsl(var(--pc));
  padding: 0.2em 0.4em;
  border-radius: 0.25em;
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 0.875em;
}

.prose pre code {
  background: none;
  padding: 0;
  color: #e6edf3;
}

.prose a {
  color: hsl(var(--pf));
  text-decoration: underline;
}

.prose a:hover {
  text-decoration: none;
}

.prose table {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
}

.prose th, .prose td {
  border: 1px solid hsl(var(--bc) / 0.3);
  padding: 0.5em 1em;
  text-align: left;
}

.prose th {
  background-color: hsl(var(--bc) / 0.1);
  font-weight: 700;
}

.prose tr:nth-child(even) {
  background-color: hsl(var(--bc) / 0.05);
}
</style>
