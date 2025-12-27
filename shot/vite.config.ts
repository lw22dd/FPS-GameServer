import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';//注意这里报错时通过快速修复可以安装对应依赖
import tailwindcss from '@tailwindcss/vite';
// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()
    ,tailwindcss() ,
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'), // 将 @ 指向 src 目录
      //
    },
  },
})


