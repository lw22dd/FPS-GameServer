<template>
  <div class="login-container">
    <div class="login-box">
      <h1>🎮 双人对战游戏</h1>
      <p class="subtitle">请输入用户名开始游戏</p>
      
      <div class="input-group">
        <input 
          type="text" 
          v-model="username" 
          placeholder="用户名（3-20字符，不能含空格）"
          @keyup.enter="login"
          maxlength="20"
        />
      </div>
      
      <div class="error-message" v-if="error">{{ error }}</div>
      
      <button @click="login" :disabled="!canLogin || loading" class="login-btn">
        {{ loading ? '登录中...' : '开始游戏' }}
      </button>
      
      <div class="tips">
        <p>💡 提示：用户名不能包含空格</p>
        <p>📡 服务器地址: http://localhost:8080</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import socketService from '@/services/socketService'

const router = useRouter()
const username = ref('')
const error = ref('')
const loading = ref(false)

const canLogin = computed(() => {
  const name = username.value.trim()
  return name.length >= 3 && name.length <= 20 && !name.includes(' ')
})

async function login() {
  error.value = ''
  
  const name = username.value.trim()
  if (name.length < 3) {
    error.value = '用户名至少3个字符'
    return
  }
  if (name.length > 20) {
    error.value = '用户名最多20个字符'
    return
  }
  if (name.includes(' ')) {
    error.value = '用户名不能包含空格'
    return
  }
  
  loading.value = true
  
  try {
    const response = await fetch('http://localhost:8080/user/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: name, password: '123456' })
    })
    
    const result = await response.json()
    
    if (!result.success) {
      if (result.message === '用户不存在') {
        await registerUser(name)
      } else {
        error.value = result.message
        loading.value = false
        return
      }
    }
    
    await socketService.connect(name)
    router.push('/rooms')
  } catch (err) {
    error.value = '连接服务器失败，请确保服务器已启动'
    console.error(err)
  } finally {
    loading.value = false
  }
}

async function registerUser(name: string) {
  const response = await fetch('http://localhost:8080/user/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: name, password: '123456', email: `${name}@game.local` })
  })
  
  const result = await response.json()
  if (!result.success) {
    error.value = result.message
    throw new Error(result.message)
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.login-box {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  text-align: center;
}

h1 {
  color: #e94560;
  margin-bottom: 10px;
  font-size: 28px;
}

.subtitle {
  color: #aaa;
  margin-bottom: 30px;
}

.input-group {
  margin-bottom: 20px;
}

input {
  width: 100%;
  padding: 15px;
  border: 2px solid #0f3460;
  border-radius: 10px;
  background: rgba(0, 0, 0, 0.3);
  color: white;
  font-size: 16px;
  box-sizing: border-box;
  transition: border-color 0.3s;
}

input:focus {
  outline: none;
  border-color: #e94560;
}

input::placeholder {
  color: #666;
}

.error-message {
  color: #ff4444;
  margin-bottom: 15px;
  font-size: 14px;
}

.login-btn {
  width: 100%;
  padding: 15px;
  background: linear-gradient(135deg, #e94560 0%, #0f3460 100%);
  border: none;
  border-radius: 10px;
  color: white;
  font-size: 18px;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(233, 69, 96, 0.4);
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.tips {
  margin-top: 30px;
  color: #666;
  font-size: 12px;
}

.tips p {
  margin: 5px 0;
}
</style>
