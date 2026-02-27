<template>
  <div class="h-full flex">
    <div class="flex-1 flex flex-col">
      <div class="flex-1 flex min-h-0">
        <div class="flex-1 flex flex-col">
          <div v-if="channelViewMode === 'chat'"
            class="flex flex-col flex-1 min-h-0 chat-card relative p-3"
            :class="{ 'dragging': isDragging }"
            @dragover.prevent="handleDragOver"
            @dragleave.prevent="handleDragLeave"
            @drop.prevent="handleDrop"
          >
            <!-- 拖放覆盖层 -->
            <div v-if="isDragging" class="absolute inset-0 bg-primary/10 border-2 border-dashed border-primary rounded-lg z-10 flex items-center justify-center">
              <div class="text-center">
                <UploadCloud class="w-16 h-16 mx-auto mb-2 text-primary opacity-60" />
                <p class="text-lg font-medium text-primary">拖放文件到此处</p>
                <p class="text-sm text-base-content/60 mt-1">松开鼠标即可上传</p>
              </div>
            </div>

            <div ref="messagesContainerRef" class="flex-1 overflow-y-auto pr-1 min-h-0 scroll-smooth">
              <div v-if="initialLoading" class="flex flex-col items-center justify-center h-32 text-xs text-base-content/30">
                <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-neutral mb-2"></div>
                <div>加载中...</div>
              </div>
              <div v-else-if="loadingMessages && messages.length === 0" class="flex flex-col items-center justify-center h-32 text-xs text-base-content/30">
                <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-neutral mb-2"></div>
                <div>加载消息中...</div>
              </div>
              <div v-else-if="messages.length === 0" class="flex flex-col items-center justify-center h-32 text-sm text-base-content/40">
                <div class="text-lg mb-1 opacity-30">暂无消息</div>
                <div class="text-xs">发送第一条消息开始交流</div>
              </div>
              <div v-else v-for="msg in messages" :key="msg.id" :class="[
                msg.user_id === 0 ? 'system-message-banner' : 'flex items-start gap-2 p-2 rounded-lg transition-colors hover:bg-base-200',
                msg.is_highlighted ? 'bg-yellow-50/50 border border-yellow-200/50 shadow-sm' : ''
              ]">
                <!-- 系统消息横幅 -->
                <template v-if="msg.user_id === 0">
                  <div class="text-center py-2 px-4 bg-base-200/50 rounded-lg mx-4">
                    <span class="text-sm text-base-content/70">{{ msg.content }}</span>
                    <span class="text-xs text-base-content/40 ml-2">{{ new Date(msg.created_at).toLocaleTimeString() }}</span>
                  </div>
                </template>
                <!-- 普通消息 -->
                <template v-else>
                  <div class="avatar">
                    <div class="rounded-full w-9 h-9 bg-neutral text-neutral-content flex items-center justify-center text-sm font-medium shrink-0 overflow-hidden">
                      <img v-if="msg.user?.avatar" :src="getFileUrl(msg.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
                      <span v-else>{{ (msg.user?.nickname || msg.user?.username || '?').charAt(0) }}</span>
                    </div>
                  </div>
                  <div class="flex-1 min-w-0" @contextmenu.prevent="handleMessageContextMenu($event, msg)">
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-semibold truncate max-w-[140px] text-base-content">
                        {{ msg.user?.nickname || msg.user?.username || '未知用户' }}
                      </span>
                      <span class="text-[10px] text-base-content/40 shrink-0">
                        {{ new Date(msg.created_at).toLocaleTimeString() }}
                      </span>
                      <span v-if="msg.is_highlighted" class="badge badge-xs bg-yellow-500 text-white border-none">精华</span>
                    </div>
                    <div v-if="msg.content" class="text-sm whitespace-pre-wrap mt-0.5 text-base-content/90">
                      {{ msg.content }}
                    </div>
                  <div v-if="msg.attachment && msg.attachment.file_path" class="mt-1">
                    <div
                      v-if="isImage(msg.attachment)"
                      @click="openMediaPlayer(msg.attachment, 'image')"
                      class="cursor-pointer hover:opacity-90 inline-block transition-opacity"
                    >
                      <img
                        :src="getFileUrl(msg.attachment.file_path)"
                        class="max-h-48 rounded-lg border border-base-300 shadow-sm"
                        alt="attachment"
                      />
                    </div>
                    <div
                      v-else-if="isVideo(msg.attachment)"
                      @click="openMediaPlayer(msg.attachment, 'video')"
                      class="cursor-pointer hover:opacity-90 inline-block transition-opacity"
                    >
                      <video
                        :src="getFileUrl(msg.attachment.file_path)"
                        class="max-h-48 rounded-lg border border-base-300 shadow-sm"
                        @loadedmetadata="(e) => { const duration = e.target.duration; if (duration && !msg.attachment._duration) { msg.attachment._duration = duration; } }"
                      />
                      <div class="text-xs text-base-content/50 mt-1 flex items-center gap-2">
                        <span v-if="msg.attachment.file_size">{{ formatFileSize(msg.attachment.file_size) }}</span>
                        <span v-if="msg.attachment._duration">{{ formatDuration(msg.attachment._duration) }}</span>
                      </div>
                    </div>
                    <div
                      v-else-if="isAudio(msg.attachment)"
                      @click="openMediaPlayer(msg.attachment, 'audio')"
                      class="cursor-pointer inline-block transition-all"
                    >
                      <div class="flex items-center gap-2 p-2 rounded-lg border border-base-300 hover:bg-base-100 hover:border-primary transition-all">
                        <Music class="w-5 h-5 text-base-content/60" />
                        <div>
                          <p class="text-sm font-medium text-base-content">{{ msg.attachment.file_name || '音频文件' }}</p>
                          <p class="text-xs text-base-content/50">点击播放音频</p>
                        </div>
                      </div>
                    </div>
                    <a
                      v-else
                      :href="getFileUrl(msg.attachment.file_path)"
                      target="_blank"
                      class="inline-flex items-center gap-2 p-2 rounded-lg border border-base-300 hover:bg-base-100 hover:border-primary transition-all text-sm text-base-content/70 hover:text-base-content"
                    >
                      <File class="w-5 h-5 text-base-content/60" />
                      <span class="truncate max-w-[200px]">{{ msg.attachment.file_name || '附件' }}</span>
                    </a>
                  </div>
                </div>
              </template>
              </div>
            </div>
            <div class="pt-2 border-t border-base-300 space-y-2">
              <template v-if="authStore.isAuthenticated">
                <input
                  v-model="newMessage"
                  type="text"
                  class="input input-bordered input-sm w-full"
                  placeholder="在频道中发送消息..."
                  @keyup.enter="sendMessage"
                />
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2 flex-1">
                    <input
                      ref="messageFileInputRef"
                      type="file"
                      class="hidden"
                      @change="handleMessageFileSelected"
                    />
                    <button class="btn btn-ghost btn-sm hover:bg-base-200" @click="triggerMessageFileSelect" :disabled="isUploading">
                      <Paperclip class="w-4 h-4" />
                      <span class="ml-1">上传</span>
                    </button>
                    <span
                      v-if="selectedMessageFileName && !isUploading"
                      class="text-[11px] text-base-content/60 truncate max-w-[160px]"
                    >
                      {{ selectedMessageFileName }}
                    </span>
                    <div v-if="isUploading" class="flex items-center gap-2 flex-1">
                      <span class="text-[11px] text-primary font-medium">
                        上传中 {{ uploadProgress }}%
                      </span>
                      <div class="flex-1 max-w-[150px]">
                        <progress class="progress progress-primary w-full" :value="uploadProgress" max="100"></progress>
                      </div>
                    </div>
                  </div>
                  <button
                    class="btn btn-neutral btn-sm"
                    @click="sendMessage"
                    :disabled="sending || isUploading || (!newMessage.trim() && !selectedMessageFile)"
                  >
                    <span v-if="sending" class="loading loading-spinner loading-xs mr-1"></span>
                    {{ sending ? '发送中...' : '发送' }}
                  </button>
                </div>
              </template>
            </div>
          </div>

          <div v-else class="flex flex-col flex-1 min-h-0 p-3">
            <div class="flex items-center justify-between mb-3 pb-2 border-b border-base-300">
              <template v-if="authStore.isAuthenticated">
                <button
                  @click="showCreateNote = true"
                  class="btn btn-neutral btn-sm"
                >
                  <Plus class="w-4 h-4 mr-1" />
                  新建笔记
                </button>
              </template>
            </div>
            <div class="flex-1 overflow-y-auto space-y-2">
              <div v-if="initialLoading" class="flex flex-col items-center justify-center h-40 text-xs text-base-content/30">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-neutral mb-3"></div>
                <div>加载中...</div>
              </div>
              <div v-else-if="loading && notes.length === 0" class="flex flex-col items-center justify-center h-40 text-xs text-base-content/30">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-neutral mb-3"></div>
                <div>加载笔记中...</div>
              </div>
              <div v-else-if="notes.length === 0" class="flex flex-col items-center justify-center h-40 text-base-content/40">
                <FileText class="w-12 h-12 mb-3 opacity-20" />
                <span class="text-sm font-medium">暂无笔记</span>
                <span v-if="authStore.isAuthenticated" class="text-xs mt-1 opacity-60">点击"新建笔记"开始记录</span>
              </div>
              <div
                v-else
                v-for="note in notes"
                :key="note.id"
                class="p-3 rounded-lg border border-base-300 hover:border-base-content hover:shadow-md cursor-pointer transition-all bg-base-100 group"
                @click="openNote(note)"
                @contextmenu.prevent="handleNoteContextMenu($event, note)"
              >
                <div class="flex items-start gap-2">
                  <FileText class="w-5 h-5 text-base-content/30 shrink-0 mt-0.5 group-hover:text-base-content transition-colors" />
                  <div class="flex-1 min-w-0">
                    <div class="font-semibold text-sm truncate group-hover:text-base-content transition-colors">
                      {{ note.title || '无标题' }}
                    </div>
                    <div v-if="tagsFor(note).length" class="mt-1.5 flex flex-wrap gap-1">
                      <span v-for="tag in tagsFor(note)" :key="tag" class="badge badge-neutral badge-xs">{{ tag }}</span>
                    </div>
                    <div class="text-[10px] text-base-content/40 mt-1.5">
                      {{ note.updated_at ? new Date(note.updated_at).toLocaleString() : '' }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showCreateNote" class="modal modal-open" @click.self="closeNoteModal">
        <div class="modal-box">
          <h3 class="font-bold text-lg text-base-content">新建频道笔记</h3>
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
                  @keyup.enter.prevent="handleAddNoteTag"
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
            <button @click="closeNoteModal" class="btn" :disabled="creatingNote">取消</button>
            <button @click="createNote" class="btn btn-neutral" :disabled="creatingNote || !newNoteTitle.trim()">
              {{ creatingNote ? '创建中...' : '创建' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="showInvite" class="modal modal-open" @click.self="showInvite = false">
        <div class="modal-box">
          <h3 class="font-bold text-lg text-base-content">邀请成员</h3>
          <input v-model="inviteUsername" type="text" placeholder="输入用户名" class="input input-bordered w-full mt-4" />
          <div class="modal-action">
            <button @click="handleInvite" class="btn btn-neutral">发送邀请</button>
            <button @click="showInvite = false" class="btn">取消</button>
          </div>
        </div>
      </div>

      <div v-if="showFilesModal" class="modal modal-open" @click.self="showFilesModal = false">
        <div class="modal-box">
          <h3 class="font-bold text-lg text-base-content mb-2">频道文件</h3>
          <div class="space-y-3 text-sm max-h-80 overflow-y-auto">
            <div>
              <div class="font-semibold text-base-content mb-1">图片</div>
              <div v-if="attachmentsByType.images.length === 0" class="text-xs text-base-content/50">
                暂无图片
              </div>
              <ul v-else class="space-y-1">
                <li
                  v-for="img in attachmentsByType.images"
                  :key="img.id"
                  class="flex items-center justify-between gap-2"
                >
                  <a
                    :href="getFileUrl(img.url)"
                    target="_blank"
                    class="truncate max-w-[220px] hover:underline"
                  >
                    {{ img.name }}
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <div class="font-semibold text-base-content mb-1">视频</div>
              <div v-if="attachmentsByType.videos.length === 0" class="text-xs text-base-content/50">
                暂无视频
              </div>
              <ul v-else class="space-y-1">
                <li
                  v-for="v in attachmentsByType.videos"
                  :key="v.id"
                  class="flex items-center justify-between gap-2"
                >
                  <a
                    :href="getFileUrl(v.url)"
                    target="_blank"
                    class="truncate max-w-[220px] hover:underline"
                  >
                    {{ v.name }}
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <div class="font-semibold text-base-content mb-1">文件</div>
              <div v-if="attachmentsByType.files.length === 0" class="text-xs text-base-content/50">
                暂无文件
              </div>
              <ul v-else class="space-y-1">
                <li
                  v-for="f in attachmentsByType.files"
                  :key="f.id"
                  class="flex items-center justify-between gap-2"
                >
                  <a
                    :href="getFileUrl(f.url)"
                    target="_blank"
                    class="truncate max-w-[220px] hover:underline"
                  >
                    {{ f.name }}
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div class="modal-action">
            <button class="btn" @click="showFilesModal = false">关闭</button>
          </div>
        </div>
      </div>
    </div>

    <aside class="hidden lg:block w-64 border-l border-base-300 px-4 py-4 space-y-3">
      <h2 class="text-sm font-bold text-base-content mb-2">成员</h2>
      <div v-if="initialLoading" class="flex flex-col items-center justify-center h-24 text-xs text-base-content/30">
        <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-neutral mb-2"></div>
        <div>加载中...</div>
      </div>
      <div v-else-if="loading && members.length === 0" class="flex flex-col items-center justify-center h-24 text-xs text-base-content/30">
        <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-neutral mb-2"></div>
        <div>加载成员中...</div>
      </div>
      <div v-else-if="members === null || members.length === 0" class="flex flex-col items-center justify-center h-24 text-xs text-base-content/40">
        <div class="text-base mb-1 opacity-30">暂无成员</div>
      </div>
      <div v-else v-for="m in members" :key="m.id" class="flex items-center space-x-2 p-2 rounded-lg border border-transparent hover:border-base-300 hover:bg-base-100 transition-all cursor-pointer" @click="openMemberPanel(m)">
        <div class="avatar">
          <div class="bg-neutral text-neutral-content rounded-full w-9 h-9 flex items-center justify-center text-sm font-medium overflow-hidden">
            <img v-if="m.user?.avatar" :src="getFileUrl(m.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
            <span v-else>{{ m.user?.username[0] }}</span>
          </div>
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-base-content truncate">{{ m.user?.nickname || m.user?.username }}</p>
          <p class="text-xs text-base-content/50 capitalize">{{ m.role }}</p>
        </div>
      </div>
    </aside>

    <!-- 媒体播放器 -->
    <MediaPlayer
      :isOpen="showMediaPlayer"
      :mediaList="mediaList"
      :currentMedia="currentMedia"
      @close="closeMediaPlayer"
    />

    <!-- 笔记管理模态框 -->
    <ManageModal
      :open="showNoteManageModal"
      type="note"
      :item="editingNote"
      @close="closeNoteManageModal"
      @save="saveNoteManage"
    />

    <!-- 消息右键菜单 -->
    <div v-if="showMessageContextMenu"
         class="fixed z-50 bg-base-100 shadow-xl rounded-lg border border-base-300 py-1 min-w-[180px] text-sm"
         :style="{ top: `${messageContextMenuY}px`, left: `${messageContextMenuX}px` }"
         @click.stop>

      <div v-if="canDeleteMessage(selectedMessage)" class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">消息操作</div>

      <button v-if="canDeleteMessage(selectedMessage)"
              @click="deleteMessage"
              class="w-full text-left px-4 py-2 hover:bg-error/10 flex items-center gap-2 text-error hover:text-error/80">
        删除消息
      </button>

      <div v-if="canHighlightMessage(selectedMessage)" class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">管理操作</div>

      <button v-if="canHighlightMessage(selectedMessage)"
              @click="toggleHighlight"
              class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2">
        {{ selectedMessage?.is_highlighted ? '取消精华' : '设为精华' }}
      </button>

      <div class="divider my-0"></div>

      <button @click="copyMessageContent"
              class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2">
        复制内容
      </button>
    </div>

    <!-- 笔记右键菜单 -->
    <div v-if="showNoteContextMenu"
         class="fixed z-50 bg-base-100 shadow-xl rounded-lg border border-base-300 py-1 min-w-[180px] text-sm"
         :style="{ top: `${noteContextMenuY}px`, left: `${noteContextMenuX}px` }"
         @click.stop>

      <div class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">笔记操作</div>

      <button @click="openNote(selectedNoteForMenu)"
              class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2">
        打开
      </button>

      <div v-if="canManageNote(selectedNoteForMenu)" class="px-3 py-1.5 text-xs font-bold text-base-content/50 uppercase tracking-wider">管理操作</div>

      <button v-if="canManageNote(selectedNoteForMenu)"
              @click="openNoteManageModal"
              class="w-full text-left px-4 py-2 hover:bg-base-200 flex items-center gap-2">
        管理
      </button>

      <button v-if="canDeleteNote(selectedNoteForMenu)"
              @click="deleteChannelNote"
              class="w-full text-left px-4 py-2 hover:bg-error/10 flex items-center gap-2 text-error hover:text-error/80">
        删除
      </button>
    </div>

    <!-- 成员管理面板（向右展开） -->
    <transition name="slide-right">
      <div v-if="showMemberPanel" class="fixed inset-y-0 right-0 w-80 bg-base-200 shadow-2xl z-50 flex flex-col">
        <div class="p-4 border-b border-base-300 flex items-center justify-between">
          <h3 class="font-bold text-lg">成员管理</h3>
          <button @click="closeMemberPanel" class="btn btn-sm btn-circle btn-ghost">✕</button>
        </div>

        <div class="flex-1 overflow-y-auto p-4">
          <div class="card bg-base-100 mb-4">
            <div class="card-body p-4 text-center">
              <div class="avatar online mx-auto mb-2">
                <div class="w-16 h-16 bg-neutral text-neutral-content rounded-full flex items-center justify-center text-xl font-bold overflow-hidden">
                  <img v-if="selectedMember?.user?.avatar" :src="getFileUrl(selectedMember.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
                  <span v-else>{{ selectedMember?.user?.username?.charAt(0) || '?' }}</span>
                </div>
              </div>
              <h4 class="font-bold">{{ selectedMember?.user?.nickname || selectedMember?.user?.username }}</h4>
              <p class="text-sm text-base-content/60">@{{ selectedMember?.user?.username }}</p>
              <div class="badge badge-neutral mt-2">{{ selectedMember?.role }}</div>
            </div>
          </div>

          <div class="space-y-2">
            <h5 class="text-sm font-bold text-base-content/70 mb-2">操作</h5>

            <!-- 自己且非owner：退出频道 -->
            <button
              v-if="isSelf(selectedMember) && !canDissolveChannel(selectedMember)"
              @click="leaveChannel"
              class="w-full btn btn-sm btn-error text-white justify-start"
            >
              退出频道
            </button>

            <!-- Owner：解散频道 -->
            <button
              v-if="canDissolveChannel(selectedMember)"
              @click="dissolveChannel"
              class="w-full btn btn-sm btn-error text-white justify-start"
            >
              解散频道
            </button>

            <!-- 管理成员角色 -->
            <div v-if="canManageMember(selectedMember)" class="dropdown dropdown-top w-full">
              <div tabindex="0" role="button" class="w-full btn btn-sm justify-start btn-warning">
                更改角色
                <ChevronDown class="w-4 h-4 ml-auto" />
              </div>
              <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-100 rounded-box w-full">
                <li>
                  <a @click="changeMemberRoleTo('admin')" :class="{ 'active': selectedMember?.role === 'admin' }">
                    管理员
                  </a>
                </li>
                <li>
                  <a @click="changeMemberRoleTo('member')" :class="{ 'active': selectedMember?.role === 'member' }">
                    成员
                  </a>
                </li>
              </ul>
            </div>

            <!-- 移出成员 -->
            <button
              v-if="canRemoveMember(selectedMember)"
              @click="removeMember"
              class="w-full btn btn-sm btn-error text-white justify-start"
            >
              移出频道
            </button>

            <!-- 无操作权限 -->
            <button
              v-if="!isSelf(selectedMember) && !canManageMember(selectedMember) && !canRemoveMember(selectedMember) && !canDissolveChannel(selectedMember)"
              class="w-full btn btn-sm btn-ghost justify-start opacity-50"
              disabled
            >
              无操作权限
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 遮罩层 -->
    <div v-if="showMemberPanel" class="fixed inset-0 bg-black/30 z-40" @click="closeMemberPanel"></div>

    <!-- 移动端成员列表抽屉 -->
    <transition name="slide-up">
      <div v-if="showMembersDrawer" class="fixed inset-0 z-50 lg:hidden">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/50" @click="showMembersDrawer = false"></div>

        <!-- Drawer Content -->
        <div class="absolute bottom-0 left-0 right-0 bg-base-200 rounded-t-2xl max-h-[80vh] overflow-hidden">
          <!-- Handle -->
          <div class="flex justify-center py-2">
            <div class="w-12 h-1 bg-base-content/30 rounded-full"></div>
          </div>

          <!-- Header -->
          <div class="px-4 pb-2 border-b border-base-300 flex items-center justify-between">
            <h3 class="text-lg font-bold">成员 ({{ members.length }})</h3>
            <button @click="showMembersDrawer = false" class="btn btn-sm btn-circle btn-ghost">✕</button>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto p-4 space-y-2">
            <div v-if="members.length === 0" class="text-center py-8 text-base-content/50">
              暂无成员
            </div>
            <div
              v-for="m in members"
              :key="m.id"
              class="flex items-center space-x-2 p-2 rounded-lg border border-base-300 hover:bg-base-100 transition-all cursor-pointer"
              @click="openMemberPanel(m)"
            >
              <div class="avatar">
                <div class="bg-neutral text-neutral-content rounded-full w-9 h-9 flex items-center justify-center text-sm font-medium overflow-hidden">
                  <img v-if="m.user?.avatar" :src="getFileUrl(m.user.avatar)" alt="avatar" class="w-full h-full object-cover" />
                  <span v-else>{{ m.user?.username[0] }}</span>
                </div>
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-semibold text-base-content truncate">{{ m.user?.nickname || m.user?.username }}</p>
                <p class="text-xs text-base-content/50 capitalize">{{ m.role }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, computed, inject } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Paperclip, FileText, Plus, UploadCloud, Music, File, ChevronDown } from 'lucide-vue-next';
import { useAuthStore } from '../stores/auth';
import api from '../api/axios';
import eventBus from '../utils/eventBus';
import MediaPlayer from '../components/MediaPlayer.vue';
import ManageModal from '../components/ManageModal.vue';
import { getFileUrl } from '../utils/urlHelper';
import wsClient from '../utils/websocket';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const channel = ref(null);
const notes = ref([]);
const members = ref([]);
const showInvite = ref(false);
const inviteUsername = ref('');
const showCreateNote = ref(false);
const newNoteTitle = ref('');
const newNoteIsPublic = ref(false);
const newNoteTags = ref([]);

// 成员管理面板
const showMemberPanel = ref(false);
const selectedMember = ref(null);

// 移动端成员列表抽屉
const showMembersDrawer = ref(false);
const newNoteTagInput = ref('');
const creatingNote = ref(false);
const noteError = ref('');
const notification = inject('notification');
const channelViewMode = inject('channelViewMode', ref('chat'));
const channelHeaderBridge = inject('channelHeaderBridge', null);
const loading = ref(false);
const initialLoading = ref(true);
let fetchSeq = 0;

const messages = ref([]);
const newMessage = ref('');
const loadingMessages = ref(false);
const sending = ref(false);
const messageFileInputRef = ref(null);
const selectedMessageFile = ref(null);
const selectedMessageFileName = ref('');
const showFilesModal = ref(false);
const uploadProgress = ref(0);
const isUploading = ref(false);
const isDragging = ref(false);
const showMediaPlayer = ref(false);
const currentMedia = ref(null);
const mediaPlayerType = ref('');

// 右键菜单状态
const showMessageContextMenu = ref(false);
const messageContextMenuX = ref(0);
const messageContextMenuY = ref(0);
const selectedMessage = ref(null);

// 笔记右键菜单状态
const showNoteContextMenu = ref(false);
const noteContextMenuX = ref(0);
const noteContextMenuY = ref(0);
const selectedNoteForMenu = ref(null);

// 笔记管理模态框状态
const showNoteManageModal = ref(false);
const editingNote = ref(null);

// 消息列表容器引用（用于自动滚动）
const messagesContainerRef = ref(null);
const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const tagsFor = (note) => parseTags(note?.tags);

const fetchData = async () => {
  try {
    const seq = ++fetchSeq;
    if (!initialLoading.value) {
      loading.value = true;
    }
    const id = route.params.id;
    const [channelRes, noteRes] = await Promise.all([
      api.get(`/channels/${id}`),
      api.get('/notes', { params: { channel_id: id } }),
    ]);
    if (seq !== fetchSeq) return;
    channel.value = channelRes.data.channel;
    members.value = channelRes.data.members || [];
    notes.value = noteRes.data;
    await fetchMessages();
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '加载频道失败', 'error');
    }
  } finally {
    loading.value = false;
    initialLoading.value = false;
  }
};

// 重新加载笔记列表（用于 WebSocket 实时更新）
const fetchNotes = async () => {
  try {
    const id = route.params.id;
    const res = await api.get('/notes', { params: { channel_id: id } });
    notes.value = res.data;
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '加载笔记失败', 'error');
    }
  }
};

// 重新加载频道信息（用于 WebSocket 实时更新）
const fetchChannelInfo = async () => {
  try {
    const id = route.params.id;
    const res = await api.get(`/channels/${id}`);
    channel.value = res.data.channel;
    members.value = res.data.members || [];
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '加载频道信息失败', 'error');
    }
  }
};

const attachmentsByType = computed(() => {
  const images = [];
  const videos = [];
  const files = [];

  for (const msg of messages.value) {
    const att = msg.attachment;
    if (!att || !att.file_path) continue;

    const id = att.id || `${msg.id}-${att.file_path}`;
    const name = att.file_name || '附件';
    const url = att.file_path;

    if (isImage(att)) {
      images.push({ id, name, url });
    } else if (isVideo(att)) {
      videos.push({ id, name, url });
    } else {
      files.push({ id, name, url });
    }
  }

  return { images, videos, files };
});

const fetchMessages = async () => {
  loadingMessages.value = true;
  try {
    const id = route.params.id;
    const res = await api.get(`/channels/${id}/messages`);
    messages.value = res.data || [];
    // 过滤掉过期的系统消息（5分钟前的）
    cleanupOldSystemMessages();
    // 加载消息后自动滚动到底部
    setTimeout(() => scrollToBottom(), 100);
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '加载消息失败', 'error');
    }
  } finally {
    loadingMessages.value = false;
  }
};

// 自动滚动到消息列表底部
const scrollToBottom = () => {
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight;
  }
};

// 监听消息变化，自动滚动到底部
watch(messages, () => {
  // 使用 nextTick 确保在 DOM 更新后滚动
  setTimeout(() => scrollToBottom(), 50);
}, { deep: true });

const triggerMessageFileSelect = () => {
  if (messageFileInputRef.value) {
    messageFileInputRef.value.value = '';
    messageFileInputRef.value.click();
  }
};

const handleMessageFileSelected = (e) => {
  const file = e.target?.files?.[0];
  selectedMessageFile.value = file || null;
  selectedMessageFileName.value = file ? file.name : '';
};

const handleDrop = (e) => {
  e.preventDefault();
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (!file) return;
  selectedMessageFile.value = file;
  selectedMessageFileName.value = file.name;
  // 自动发送（只有文件，没有文本）
  if (!newMessage.value.trim()) {
    sendMessage();
  }
};

const handleDragOver = (e) => {
  e.preventDefault();
  isDragging.value = true;
};

const handleDragLeave = (e) => {
  e.preventDefault();
  // 只在真正离开聊天卡片时才隐藏拖放提示
  const rect = e.currentTarget.getBoundingClientRect();
  const x = e.clientX;
  const y = e.clientY;
  if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
    isDragging.value = false;
  }
};

const isImage = (attachment) => {
  if (!attachment || !attachment.file_path) return false;
  return /\.(png|jpe?g|gif|webp|bmp)$/i.test(attachment.file_path);
};

const isVideo = (attachment) => {
  if (!attachment || !attachment.file_path) return false;
  return /\.(mp4|webm|ogg|mov)$/i.test(attachment.file_path);
};

const isAudio = (attachment) => {
  if (!attachment || !attachment.file_path) return false;
  return /\.(mp3|wav|ogg|m4a|flac)$/i.test(attachment.file_path);
};

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes) return '';
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

// 格式化视频时长
const formatDuration = (seconds) => {
  if (!seconds || seconds <= 0) return '';
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
};

// 获取所有媒体文件列表
const mediaList = computed(() => {
  const list = [];
  for (const msg of messages.value) {
    const att = msg.attachment;
    if (!att || !att.file_path) continue;

    const type = isImage(att) ? 'image' : isVideo(att) ? 'video' : isAudio(att) ? 'audio' : null;
    if (type) {
      list.push({
        ...att,
        type
      });
    }
  }
  return list;
});

// 打开媒体播放器
const openMediaPlayer = (attachment, type) => {
  currentMedia.value = attachment;
  mediaPlayerType.value = type;
  showMediaPlayer.value = true;
};

// 关闭媒体播放器
const closeMediaPlayer = () => {
  showMediaPlayer.value = false;
  currentMedia.value = null;
  mediaPlayerType.value = '';
};

// 右键菜单处理
const handleMessageContextMenu = (event, message) => {
  if (!authStore.isAuthenticated) return;
  
  selectedMessage.value = message;
  
  // 计算菜单位置
  const menuWidth = 180;
  const menuHeight = 300;
  let x = event.clientX;
  let y = event.clientY;

  // 边界检查
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 10;
  }
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 10;
  }

  messageContextMenuX.value = x;
  messageContextMenuY.value = y;
  showMessageContextMenu.value = true;
};

// 关闭右键菜单
const closeMessageContextMenu = () => {
  showMessageContextMenu.value = false;
  selectedMessage.value = null;
};

// 笔记右键菜单处理
const handleNoteContextMenu = (event, note) => {
  if (!authStore.isAuthenticated) return;

  selectedNoteForMenu.value = note;

  // 计算菜单位置
  const menuWidth = 180;
  const menuHeight = 150;
  let x = event.clientX;
  let y = event.clientY;

  // 边界检查
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 10;
  }
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 10;
  }

  noteContextMenuX.value = x;
  noteContextMenuY.value = y;
  showNoteContextMenu.value = true;
};

// 关闭笔记右键菜单
const closeNoteContextMenu = () => {
  showNoteContextMenu.value = false;
  selectedNoteForMenu.value = null;
};

// 检查是否可以管理笔记
const canManageNote = (note) => {
  if (!authStore.isAuthenticated || !note) return false;

  // 如果是笔记的所有者，可以管理
  if (note.owner_id === authStore.user?.id) return true;

  // 检查是否是频道管理员或所有者
  const member = members.value.find(m => m.user_id === authStore.user?.id);
  return member && (member.role === 'admin' || member.role === 'owner');
};

// 检查是否可以删除笔记
const canDeleteNote = (note) => {
  if (!authStore.isAuthenticated || !note) return false;

  // 如果是笔记的所有者，可以删除
  if (note.owner_id === authStore.user?.id) return true;

  // 检查是否是频道管理员或所有者
  const member = members.value.find(m => m.user_id === authStore.user?.id);
  return member && (member.role === 'admin' || member.role === 'owner');
};

// 删除频道笔记
const deleteChannelNote = async () => {
  if (!selectedNoteForMenu.value) return;

  if (!confirm(`确定要删除笔记 "${selectedNoteForMenu.value.title || '无标题'}" 吗？`)) return;

  try {
    await api.delete(`/notes/${selectedNoteForMenu.value.id}`);
    notes.value = notes.value.filter(n => n.id !== selectedNoteForMenu.value.id);
    if (notification) {
      notification.showNotification('笔记已删除', 'success');
    }
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '删除失败', 'error');
    }
  }

  closeNoteContextMenu();
};

// 打开笔记管理模态框
const openNoteManageModal = () => {
  if (!selectedNoteForMenu.value) return;
  editingNote.value = selectedNoteForMenu.value;
  showNoteManageModal.value = true;
  closeNoteContextMenu();
};

// 关闭笔记管理模态框
const closeNoteManageModal = () => {
  showNoteManageModal.value = false;
  editingNote.value = null;
};

// 保存笔记管理
const saveNoteManage = async (payload) => {
  if (!editingNote.value) return;
  try {
    const res = await api.put(`/notes/${editingNote.value.id}`, {
      title: payload.title,
      is_public: payload.is_public,
      tags: payload.tags
    });
    
    // 更新本地笔记列表
    const noteIndex = notes.value.findIndex(n => n.id === editingNote.value.id);
    if (noteIndex !== -1) {
      notes.value[noteIndex] = { ...notes.value[noteIndex], ...res.data };
    }
    
    // 通知Sidebar更新
    eventBus.emit('note-updated', {
      id: editingNote.value.id,
      title: payload.title,
      is_public: payload.is_public,
      tags: payload.tags
    });
    
    if (notification) {
      notification.showNotification('保存成功', 'success');
    }
    closeNoteManageModal();
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '保存失败', 'error');
    }
  }
};

// 检查是否可以删除消息
const canDeleteMessage = (message) => {
  if (!authStore.isAuthenticated || !message) return false;
  
  // 如果是自己的消息，可以删除
  if (message.user_id === authStore.user?.id) return true;
  
  // 检查是否是管理员或所有者
  const member = members.value.find(m => m.user_id === authStore.user?.id);
  return member && (member.role === 'admin' || member.role === 'owner');
};

// 检查是否可以设置精华
const canHighlightMessage = (message) => {
  if (!authStore.isAuthenticated || !message) return false;
  
  // 检查是否是管理员或所有者
  const member = members.value.find(m => m.user_id === authStore.user?.id);
  return member && (member.role === 'admin' || member.role === 'owner');
};

// 删除消息
const deleteMessage = async () => {
  if (!selectedMessage.value) return;

  // 保存消息 ID，防止 selectedMessage 在执行过程中被修改
  const messageId = selectedMessage.value.id;

  try {
    await api.delete(`/channels/${route.params.id}/messages/${messageId}`);

    // 从消息列表中移除消息
    const messageIndex = messages.value.findIndex(m => m.id === messageId);
    if (messageIndex !== -1) {
      messages.value.splice(messageIndex, 1);
    }

    if (notification) {
      notification.showNotification('消息已删除', 'success');
    }
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '删除失败', 'error');
    }
  }

  closeMessageContextMenu();
};

// 切换精华状态
const toggleHighlight = async () => {
  if (!selectedMessage.value) return;

  try {
    const res = await api.put(`/channels/${route.params.id}/messages/${selectedMessage.value.id}/highlight`);

    // 更新本地消息状态
    const messageIndex = messages.value.findIndex(m => m.id === selectedMessage.value.id);
    if (messageIndex !== -1) {
      messages.value[messageIndex].is_highlighted = res.data.is_highlighted;
    }

    if (notification) {
      notification.showNotification(res.data.message, 'success');
    }
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '操作失败', 'error');
    }
  }

  closeMessageContextMenu();
};

// 复制消息内容
const copyMessageContent = async () => {
  if (!selectedMessage.value) return;
  
  let content = '';
  if (selectedMessage.value.content) {
    content = selectedMessage.value.content;
  } else if (selectedMessage.value.attachment) {
    content = `[文件: ${selectedMessage.value.attachment.file_name}]`;
  }
  
  if (content) {
    try {
      await navigator.clipboard.writeText(content);
      if (notification) {
        notification.showNotification('内容已复制到剪贴板', 'success');
      }
    } catch (err) {
      if (notification) {
        notification.showNotification('复制失败', 'error');
      }
    }
  }
  
  closeMessageContextMenu();
};

const sendMessage = async () => {
  if (!newMessage.value.trim() && !selectedMessageFile.value) return;

  // 检查用户认证状态
  if (!authStore.isAuthenticated) {
    if (notification) {
      notification.showNotification('请先登录后再发送消息', 'error');
    }
    return;
  }

  sending.value = true;
  try {
    const id = route.params.id;
    let attachmentId = null;

    if (selectedMessageFile.value) {
      isUploading.value = true;
      uploadProgress.value = 0;

      const form = new FormData();
      form.append('file', selectedMessageFile.value);
      form.append('type', 'channel');
      form.append('channel_id', id);

      const uploadRes = await api.post('/upload', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
        timeout: 60000,
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total) {
            uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          }
        },
      });

      attachmentId = uploadRes.data?.id;
      isUploading.value = false;

      if (notification) {
        notification.showNotification('文件上传成功', 'success');
      }
    }

    const payload = {
      content: newMessage.value.trim(),
      attachment_id: attachmentId,
    };

    await api.post(`/channels/${id}/messages`, payload);

    newMessage.value = '';
    selectedMessageFile.value = null;
    selectedMessageFileName.value = '';

    if (notification) {
      notification.showNotification('消息发送成功', 'success');
    }

  } catch (err) {
      isUploading.value = false;
      uploadProgress.value = 0;
  
      if (notification) {
        let errorMessage = '发送失败';      if (err.code === 'ECONNABORTED') {
        errorMessage = '上传超时，请重试';
      } else if (err.response?.status === 413) {
        errorMessage = '文件太大';
      } else if (err.response?.status === 415) {
        errorMessage = '不支持的文件格式';
      } else if (err.response?.status === 500) {
        errorMessage = '服务器错误，请稍后重试';
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error;
      } else if (err.message) {
        errorMessage = `网络错误: ${err.message}`;
      }
      notification.showNotification(errorMessage, 'error');
    }
  } finally {
    sending.value = false;
  }
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
      tags: newNoteTags.value.join(','),
      channel_id: parseInt(route.params.id)
    });
    
    notes.value.unshift(res.data);
    closeNoteModal();
    
    // Navigate to the newly created note
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

const openNote = (note) => {
  router.push({ name: 'note-editor', params: { id: note.id } });
};

const handleInvite = async () => {
  try {
    await api.post('/channels/invite', {
      channel_id: parseInt(route.params.id),
      username: inviteUsername.value
    });
    if (notification) {
      notification.showNotification('已发送邀请');
    }
    showInvite.value = false;
    inviteUsername.value = '';
  } catch (err) {
    const errorMsg = '邀请失败: ' + (err.response?.data?.error || '未知错误');
    if (notification) {
      notification.showNotification(errorMsg, 'error');
    }
  }
};

// 成员管理方法
const openMemberPanel = (member) => {
  selectedMember.value = member;
  showMemberPanel.value = true;
};

const closeMemberPanel = () => {
  showMemberPanel.value = false;
  selectedMember.value = null;
};

const canManageMember = (member) => {
  if (!authStore.isAuthenticated || !member) return false;
  const currentMember = members.value.find(m => m.user_id === authStore.user?.id);
  if (!currentMember) return false;

  // 不能操作自己
  if (member.user_id === authStore.user?.id) return false;

  // 只有管理员和所有者可以管理成员
  if (currentMember.role !== 'admin' && currentMember.role !== 'owner') return false;

  // 不能管理所有者
  if (member.role === 'owner') return false;

  // 管理员不能管理其他管理员
  if (currentMember.role === 'admin' && member.role === 'admin') return false;

  return true;
};

const canRemoveMember = (member) => {
  if (!authStore.isAuthenticated || !member) return false;
  const currentMember = members.value.find(m => m.user_id === authStore.user?.id);
  if (!currentMember) return false;

  // 不能操作自己
  if (member.user_id === authStore.user?.id) return false;

  // 不能移出所有者
  if (member.role === 'owner') return false;

  // admin 只能移出普通成员
  if (currentMember.role === 'admin') {
    return member.role === 'member';
  }

  // owner 可以移出任何人（除了自己）
  if (currentMember.role === 'owner') {
    return true;
  }

  return false;
};

const isSelf = (member) => {
  if (!member) return false;
  return member.user_id === authStore.user?.id;
};

const canDissolveChannel = (member) => {
  if (!member) return false;
  const currentMember = members.value.find(m => m.user_id === authStore.user?.id);
  if (!currentMember) return false;

  // 只有 owner 点击自己时可以解散频道
  return currentMember.role === 'owner' && member.role === 'owner' && member.user_id === authStore.user?.id;
};

const leaveChannel = async () => {
  if (!selectedMember.value) return;

  if (!confirm('确定要退出该频道吗？')) return;

  try {
    await api.delete(`/channels/${route.params.id}/members/${authStore.user?.id}`);
    if (notification) {
      notification.showNotification('已退出频道', 'success');
    }
    router.push('/channels');
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '退出失败', 'error');
    }
  }
};

const dissolveChannel = async () => {
  if (!selectedMember.value) return;

  if (!confirm('确定要解散该频道吗？此操作不可恢复！')) return;

  try {
    await api.delete(`/channels/${route.params.id}`);
    if (notification) {
      notification.showNotification('频道已解散', 'success');
    }
    router.push('/channels');
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '解散失败', 'error');
    }
  }
};

const changeMemberRoleTo = async (newRole) => {
  if (!selectedMember.value) return;
  
  // 如果角色没有变化，不执行任何操作
  if (selectedMember.value.role === newRole) return;

  try {
    await api.put(`/channels/${route.params.id}/members/${selectedMember.value.user_id}`, {
      role: newRole
    });
    if (notification) {
      notification.showNotification('角色已更改', 'success');
    }
    fetchData(); // 重新加载数据
    closeMemberPanel();
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '更改失败', 'error');
    }
  }
};

const removeMember = async () => {
  if (!selectedMember.value) return;

  if (!confirm(`确定要移出 ${selectedMember.value.user?.nickname || selectedMember.value.user?.username} 吗？`)) return;

  try {
    await api.delete(`/channels/${route.params.id}/members/${selectedMember.value.user_id}`);
    if (notification) {
      notification.showNotification('成员已移出', 'success');
    }
    fetchData(); // 重新加载数据
    closeMemberPanel();
  } catch (err) {
    if (notification) {
      notification.showNotification(err.response?.data?.error || '移出失败', 'error');
    }
  }
};

if (channelHeaderBridge) {
  channelHeaderBridge.openFiles = () => {
    showFilesModal.value = true;
  };
}

// 处理笔记更新事件
const handleNoteUpdated = (payload) => {
  if (!payload || !payload.id) return;
  const target = notes.value.find((note) => note.id === Number(payload.id));
  if (!target) return;
  if (typeof payload.title === 'string') target.title = payload.title;
  if (typeof payload.is_public === 'boolean') target.is_public = payload.is_public;
  if (typeof payload.tags === 'string') target.tags = payload.tags;
};

// WebSocket 消息处理（移到 onMounted 外部，避免重复定义）
const handleWsMessage = async (message) => {
  if (message.type === 'note') {
    if (message.action === 'create' || message.action === 'update') {
      // 重新加载频道笔记
      await fetchNotes();
    } else if (message.action === 'delete') {
      // 删除已删除的笔记（确保类型一致的比较）
      const deleteId = String(message.data.id);
      notes.value = notes.value.filter(n => String(n.id) !== deleteId);
    }
  } else if (message.type === 'channel') {
    if (message.action === 'update') {
      // 重新加载频道信息
      await fetchChannelInfo();
    }
  } else if (message.type === 'message') {
    if (message.action === 'create' && message.data) {
      // 新消息，添加到消息列表
      const messageId = message.data.id || message.data.message_id;
      if (messageId) {
        const messageExists = messages.value.some(m => String(m.id) === String(messageId));
        if (!messageExists) {
          messages.value.push(message.data);
        }
      }
    } else if (message.action === 'delete' && message.data) {
      // 删除消息
      const deleteId = message.data.id || message.data.message_id;
      if (deleteId) {
        messages.value = messages.value.filter(m => String(m.id) !== String(deleteId));
      }
    } else if (message.action === 'highlight' && message.data) {
      // 更新精华状态
      const msgId = message.data.id || message.data.message_id;
      if (msgId && message.data.is_highlighted !== undefined) {
        const msgIndex = messages.value.findIndex(m => String(m.id) === String(msgId));
        if (msgIndex !== -1) {
          messages.value[msgIndex].is_highlighted = message.data.is_highlighted;
        }
      }
    }
  }
};

// 设置 WebSocket 监听器的函数
let isListenersSetup = false;
const setupWebSocketListeners = () => {
  if (isListenersSetup) return;
  
  if (localStorage.getItem('userId')) {
    wsClient.on('note_create', handleWsMessage);
    wsClient.on('note_update', handleWsMessage);
    wsClient.on('note_delete', handleWsMessage);
    wsClient.on('channel_update', handleWsMessage);
    wsClient.on('message_create', handleWsMessage);
    wsClient.on('message_delete', handleWsMessage);
    wsClient.on('message_highlight', handleWsMessage);
    isListenersSetup = true;
  }
};

// 系统消息自动清理定时器
let systemMessageCleanupTimer = null;

// 清理过期的系统消息
const cleanupOldSystemMessages = () => {
  const now = Date.now();
  const fiveMinutesInMs = 5 * 60 * 1000;

  messages.value = messages.value.filter(msg => {
    if (msg.user_id === 0) {
      const messageTime = new Date(msg.created_at).getTime();
      if (isNaN(messageTime)) {
        return false;
      }
      const age = now - messageTime;
      return age < fiveMinutesInMs;
    }
    return true;
  });
};

onMounted(() => {
  fetchData();
  eventBus.on('note-updated', handleNoteUpdated);

  setupWebSocketListeners();
  wsClient.on('connected', setupWebSocketListeners);

  // 添加全局点击事件监听器来关闭右键菜单
  window.addEventListener('click', () => {
    closeMessageContextMenu();
    closeNoteContextMenu();
  });

  // 添加全局拖放事件监听器，防止文件被浏览器打开
  window.addEventListener('dragover', (e) => e.preventDefault());
  window.addEventListener('drop', (e) => e.preventDefault());

  // 每分钟清理一次过期的系统消息
  systemMessageCleanupTimer = setInterval(cleanupOldSystemMessages, 60000);
});

onBeforeUnmount(() => {
  eventBus.off('note-updated', handleNoteUpdated);

  // 清除 WebSocket 监听器
  wsClient.off('note_create', handleWsMessage);
  wsClient.off('note_update', handleWsMessage);
  wsClient.off('note_delete', handleWsMessage);
  wsClient.off('channel_update', handleWsMessage);
  wsClient.off('message_create', handleWsMessage);
  wsClient.off('message_delete', handleWsMessage);
  wsClient.off('message_highlight', handleWsMessage);
  wsClient.off('connected', setupWebSocketListeners);

  // 清除系统消息清理定时器
  if (systemMessageCleanupTimer) {
    clearInterval(systemMessageCleanupTimer);
  }

  // 重置标志位
  isListenersSetup = false;

  // 移除全局点击事件监听器
  window.removeEventListener('click', () => {
    closeMessageContextMenu();
    closeNoteContextMenu();
  });

  // 移除全局拖放事件监听器
  window.removeEventListener('dragover', (e) => e.preventDefault());
  window.removeEventListener('drop', (e) => e.preventDefault());
});

watch(
  () => route.params.id,
  () => {
    fetchData();
  }
);
</script>

<style scoped>
/* 系统消息横幅样式 */
.system-message-banner {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 16px;
  margin: 8px 0;
}

/* 拖放状态样式 */
.chat-card.dragging {
  background-color: oklch(var(--b3));
  transition: all 0.2s ease;
}

/* 滑动过渡动画 */
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.2s ease;
}

.slide-right-enter-from {
  transform: translateX(100%);
}

.slide-right-leave-to {
  transform: translateX(100%);
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.2s ease;
}

.slide-up-enter-from {
  transform: translateY(100%);
}

.slide-up-leave-to {
  transform: translateY(100%);
}
</style>
