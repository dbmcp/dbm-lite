/**
 * v-ellipsis: 响应式文本缩写指令
 * 功能：
 *   1. 自动给元素添加单行省略号样式（overflow: hidden + text-overflow: ellipsis）
 *   2. 使用 ResizeObserver 动态监测元素宽度变化
 *   3. 当检测到文本溢出时，自动设置 title 属性，实现鼠标悬停显示完整文本
 *   4. 未溢出时移除 title，避免不必要的提示
 *
 * 用法：
 *   <span v-ellipsis>很长的文本内容</span>
 *   <span v-ellipsis="200">限制最大宽度 200px</span>
 *   <div v-ellipsis class="my-title">{{ title }}</div>
 *
 * 兼容：Chrome 64+ / Firefox 69+ / Safari 13.1+ / Edge 79+
 */
import type { Directive, DirectiveBinding } from 'vue'

interface EllipsisElement extends HTMLElement {
  __ellipsis_ro__?: ResizeObserver
  __ellipsis_original_text?: string
  __ellipsis_checker__?: () => void
}

/**
 * 检测文本是否溢出
 */
function isOverflow(el: HTMLElement): boolean {
  // 优先使用 scrollWidth vs clientWidth
  // 考虑到 border/padding，需加 1 像素容差
  return el.scrollWidth > el.clientWidth + 1
}

/**
 * 更新 title（溢出时才设置）
 */
function updateTitle(el: EllipsisElement) {
  const text = el.textContent?.trim() || ''
  // 保存原始文本供 tooltip 使用
  el.__ellipsis_original_text__ = text
  if (text && isOverflow(el)) {
    el.setAttribute('title', text)
  } else {
    el.removeAttribute('title')
  }
}

/**
 * 给元素设置省略号样式
 */
function applyStyles(el: HTMLElement, maxWidth?: string | number) {
  const style = el.style
  style.overflow = 'hidden'
  style.textOverflow = 'ellipsis'
  style.whiteSpace = 'nowrap'
  if (maxWidth !== undefined && maxWidth !== null && maxWidth !== '') {
    if (typeof maxWidth === 'number') {
      style.maxWidth = `${maxWidth}px`
    } else {
      style.maxWidth = String(maxWidth)
    }
  } else {
    // 默认使用父容器的 100%
    if (!style.maxWidth || style.maxWidth === 'none') {
      style.maxWidth = '100%'
    }
  }
  // 确保 inline-block / block 可正确生效
  if (style.display === 'inline' || !style.display) {
    style.display = 'inline-block'
  }
}

// 定义指令对象
const vEllipsis: Directive<EllipsisElement, string | number> = {
  mounted(el: EllipsisElement, binding: DirectiveBinding<string | number>) {
    // 应用样式
    applyStyles(el, binding.value)
    
    // 初始检测
    updateTitle(el)
    
    // ResizeObserver 监听元素尺寸变化
    const check = () => updateTitle(el)
    el.__ellipsis_checker__ = check
    
    if (typeof ResizeObserver !== 'undefined') {
      const ro = new ResizeObserver(check)
      ro.observe(el)
      el.__ellipsis_ro__ = ro
    }
    
    // MutationObserver 监听文本内容变化
    if (typeof MutationObserver !== 'undefined') {
      const mo = new MutationObserver(() => check())
      mo.observe(el, {
        childList: true,
        subtree: true,
        characterData: true
      })
      ;(el as any).__ellipsis_mo__ = mo
    }
    
    // 延迟再检测一次（确保 DOM 完全渲染
    setTimeout(check, 100)
  },
  
  updated(el: EllipsisElement, binding: DirectiveBinding<string | number>) {
    // 可能更新了内容或宽度限制
    applyStyles(el, binding.value)
    if (el.__ellipsis_checker__) {
      el.__ellipsis_checker__()
    }
  },
  
  unmounted(el: EllipsisElement) {
    if (el.__ellipsis_ro__) {
      el.__ellipsis_ro__.disconnect()
      el.__ellipsis_ro__ = undefined
    }
    if ((el as any).__ellipsis_mo__) {
      ;(el as any).__ellipsis_mo__.disconnect()
      ;(el as any).__ellipsis_mo__ = undefined
    }
  }
}

export default vEllipsis
