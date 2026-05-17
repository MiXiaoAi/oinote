import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'
import { Decoration, DecorationSet } from '@tiptap/pm/view'

/**
 * 远程光标 Tiptap 扩展
 * 使用 ProseMirror 装饰器实现多人协作时的光标和选区显示
 */
export const RemoteCursorExtension = Extension.create({
  name: 'remoteCursor',

  addOptions() {
    return {
      cursors: new Map(),
      onUpdate: null,
    }
  },

  addProseMirrorPlugins() {
    const extension = this

    return [
      new Plugin({
        key: new PluginKey('remoteCursor'),

        state: {
          init() {
            return DecorationSet.empty
          },

          apply(tr, oldState) {
            // 当远程光标更新时，或者文档内容变化时，重新计算装饰器
            const isCursorUpdate = tr.getMeta('remoteCursorUpdate')
            const docChanged = tr.docChanged
            
            if (!isCursorUpdate && !docChanged) {
              // 如果不是光标更新也不是文档变化，返回旧的装饰器
              return oldState
            }

            const cursors = extension.options.cursors
            if (!cursors || cursors.size === 0) {
              return DecorationSet.empty
            }

            const decorations = []
            const doc = tr.doc

            for (const [clientId, cursor] of cursors.entries()) {
              try {
                const { from, to, color, username, nickname } = cursor
                const displayName = nickname || username
                const docSize = doc.content.size

                if (docSize <= 1) continue

                const safeFrom = Math.min(Math.max(0, from), docSize - 1)
                const safeTo = Math.min(Math.max(0, to), docSize - 1)

                // 选区高亮
                if (safeFrom !== safeTo) {
                  const highlightDecoration = Decoration.inline(
                    safeFrom,
                    safeTo,
                    {
                      class: 'remote-selection',
                      style: `background-color: ${color}33;`,
                      'data-client-id': clientId,
                    }
                  )
                  decorations.push(highlightDecoration)
                }

                // 光标小部件
                const cursorDecoration = Decoration.widget(safeFrom, (view) => {
                  const el = document.createElement('span')
                  el.className = 'remote-cursor-widget'
                  el.setAttribute('data-client-id', clientId)
                  el.style.setProperty('--cursor-color', color)

                  const caret = document.createElement('span')
                  caret.className = 'remote-cursor-caret'
                  caret.style.backgroundColor = color

                  const label = document.createElement('span')
                  label.className = 'remote-cursor-label'
                  label.style.backgroundColor = color
                  label.textContent = displayName

                  el.appendChild(caret)
                  el.appendChild(label)

                  return el
                }, {
                  side: -1,
                  key: `cursor-${clientId}`,
                })
                decorations.push(cursorDecoration)
              } catch (e) {
                // 忽略单个光标的错误
              }
            }

            return DecorationSet.create(doc, decorations)
          },
        },

        props: {
          decorations(state) {
            return this.getState(state)
          },
        },
      }),
    ]
  },
})

export default RemoteCursorExtension
