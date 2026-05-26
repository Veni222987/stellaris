<!-- 顶层壳：深空背景 + naive message/dialog provider（仅供 toast/confirm） -->
<template>
  <n-config-provider :theme="darkTheme" :theme-overrides="overrides">
    <n-message-provider>
      <n-dialog-provider>
        <div class="cosmos" />
        <div class="starfield" />
        <router-view />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'

const overrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#41e3c9',
    primaryColorHover: '#5cf0d8',
    primaryColorPressed: '#2fc7af',
    fontFamily: "'IBM Plex Sans', system-ui, sans-serif",
  },
}
</script>
