<template>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup>
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body,
#app {
  min-width: 0;
  min-height: 100%;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: #f5f7fa;
  overflow-x: hidden;
}

button,
input,
textarea,
select {
  font: inherit;
}

img,
svg {
  max-width: 100%;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 全局卡片样式 */
.card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

/* 页面标题 */
.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 24px;
}

/* 现代卡片式按钮 */
.btn-primary {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
}

/* 正确/错误颜色 */
.correct-color {
  color: #67c23a;
}

.wrong-color {
  color: #f56c6c;
}

@media (max-width: 640px) {
  .card {
    padding: 16px;
    border-radius: 10px;
  }

  .page-title {
    margin-bottom: 16px;
    font-size: 20px;
  }

  .el-message {
    width: calc(100vw - 32px) !important;
    min-width: 0 !important;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
