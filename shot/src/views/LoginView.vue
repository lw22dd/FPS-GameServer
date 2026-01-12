<template>
  <div class="login-container">
    <div class="login-box">
      <h1>🎮 双人对战游戏</h1>
      <p class="subtitle">{{ isLogin ? '请输入用户名和密码登录' : '请输入用户名和密码注册' }}</p>
      
      <div class="input-group">
        <input 
          type="text" 
          v-model="username" 
          placeholder="用户名（3-20字符，不能含空格）"
          @keyup.enter="isLogin ? login : register"
          @blur="checkUsernameUnique"
          maxlength="20"
          :class="{ 'input-error': usernameError }"
        />
        <div class="field-error" v-if="usernameError">{{ usernameError }}</div>
      </div>
      
      <div class="input-group">
        <input 
          type="password" 
          v-model="password" 
          placeholder="密码（至少6位）"
          @keyup.enter="isLogin ? login : register"
          maxlength="20"
          :class="{ 'input-error': passwordError }"
        />
        <div class="field-error" v-if="passwordError">{{ passwordError }}</div>
      </div>
      
      <div class="error-message" v-if="globalError">{{ globalError }}</div>
      
      <button @click="login" v-if="isLogin" :disabled="!canLogin || loading" class="login-btn">
        {{ loading ? '登录中...' : '登录' }}
      </button>
      
      <button @click="register" v-else :disabled="!canRegister || loading" class="register-btn">
        {{ loading ? '注册中...' : '注册' }}
      </button>
      
      <div class="toggle-mode">
        <button @click="toggleMode" class="toggle-btn">
          {{ isLogin ? '没有账号？点击注册' : '已有账号？点击登录' }}
        </button>
      </div>
      
      <div class="tips">
        <p>💡 提示：用户名不能包含空格，密码至少6位</p>
        <p>📡 服务器地址: http://localhost:8080</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useSocketStore } from '@/stores/socketStore'

const router = useRouter()
const socketStore = useSocketStore()
const isLogin = ref(true)
const username = ref('')
const password = ref('')
const globalError = ref('')
const usernameError = ref('')
const passwordError = ref('')
const loading = ref(false)
const checkingUsername = ref(false)
const usernameUnique = ref(true)

// 监听用户名变化，重置错误
watch(username, () => {
  usernameError.value = ''
})

// 监听密码变化，重置错误
watch(password, () => {
  passwordError.value = ''
})

// 登录表单验证
const canLogin = computed(() => {
  const name = username.value.trim()
  const pass = password.value.trim()
  return name.length >= 3 && name.length <= 20 && !name.includes(' ') && pass.length >= 6
})

// 注册表单验证
const canRegister = computed(() => {
  const name = username.value.trim()
  const pass = password.value.trim()
  return name.length >= 3 && name.length <= 20 && !name.includes(' ') && pass.length >= 6 && usernameUnique.value
})

// 切换登录/注册模式
function toggleMode() {
  isLogin.value = !isLogin.value
  // 重置表单
  username.value = ''
  password.value = ''
  globalError.value = ''
  usernameError.value = ''
  passwordError.value = ''
  usernameUnique.value = true
}

// 检查用户名唯一性
async function checkUsernameUnique() {
  const name = username.value.trim()
  if (name.length < 3 || name.length > 20 || name.includes(' ')) {
    return
  }
  
  checkingUsername.value = true
  usernameError.value = ''
  
  try {
    // 使用假数据模拟API调用，因为实际API可能还未实现
    // 在真实环境中，应该使用下面的代码：
    // const result = await UserApi.checkUsernameExists(name)
    // usernameUnique.value = !result.data
    // if (!usernameUnique.value) {
    //   usernameError.value = '用户已存在'
    // }
    
    // 模拟API调用延迟
    await new Promise(resolve => setTimeout(resolve, 500))
    usernameUnique.value = true
  } catch (err) {
    console.error('检查用户名失败:', err)
  } finally {
    checkingUsername.value = false
  }
}

// 验证用户名
function validateUsername() {
  const name = username.value.trim()
  
  if (name.length < 3) {
    usernameError.value = '用户名至少3个字符'
    return false
  }
  if (name.length > 20) {
    usernameError.value = '用户名最多20个字符'
    return false
  }
  if (name.includes(' ')) {
    usernameError.value = '用户名不能包含空格'
    return false
  }
  if (!usernameUnique.value) {
    usernameError.value = '用户已存在'
    return false
  }
  
  return true
}

// 验证密码
function validatePassword() {
  const pass = password.value.trim()
  
  if (pass.length < 6) {
    passwordError.value = '密码至少6位'
    return false
  }
  
  return true
}

// 登录功能
async function login() {
  globalError.value = ''
  usernameError.value = ''
  passwordError.value = ''
  
  const name = username.value.trim()
  const pass = password.value.trim()
  
  if (!validateUsername()) {
    return
  }
  if (!validatePassword()) {
    return
  }
  
  loading.value = true
  
  try {
    const response = await fetch('http://localhost:8080/user/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: name, password: pass })
    })
    
    const result = await response.json()
    
    if (!result.success) {
      globalError.value = result.message
      loading.value = false
      return
    }
    
    await socketStore.connect(name)
    router.push('/rooms')
  } catch (err) {
    globalError.value = '连接服务器失败，请确保服务器已启动'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// 注册功能
async function register() {
  globalError.value = ''
  usernameError.value = ''
  passwordError.value = ''
  
  const name = username.value.trim()
  const pass = password.value.trim()
  
  if (!validateUsername()) {
    return
  }
  if (!validatePassword()) {
    return
  }
  
  // 再次检查用户名唯一性
  await checkUsernameUnique()
  if (!usernameUnique.value) {
    usernameError.value = '用户已存在'
    return
  }
  
  loading.value = true
  
  try {
    const response = await fetch('http://localhost:8080/user/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: name, password: pass, email: `${name}@game.local` })
    })
    
    const result = await response.json()
    
    if (!result.success) {
      if (result.message === '用户名已存在') {
        usernameError.value = '用户已存在'
      } else {
        globalError.value = result.message
      }
      loading.value = false
      return
    }
    
    // 注册成功后自动登录
    await socketStore.connect(name)
    router.push('/rooms')
  } catch (err) {
    globalError.value = '连接服务器失败，请确保服务器已启动'
    console.error(err)
  } finally {
    loading.value = false
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
  text-align: left;
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

.input-error {
  border-color: #ff4444 !important;
}

.error-message {
  color: #ff4444;
  margin-bottom: 15px;
  font-size: 14px;
}

.field-error {
  color: #ff4444;
  font-size: 12px;
  margin-top: 5px;
  margin-left: 5px;
  text-align: left;
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

.register-btn {
  width: 100%;
  padding: 15px;
  background: linear-gradient(135deg, #00b894 0%, #0f3460 100%);
  border: none;
  border-radius: 10px;
  color: white;
  font-size: 18px;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.register-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(0, 184, 148, 0.4);
}

.login-btn:disabled,
.register-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-mode {
  margin-top: 15px;
}

.toggle-btn {
  background: none;
  border: none;
  color: #00b894;
  cursor: pointer;
  font-size: 14px;
  text-decoration: underline;
  padding: 5px;
  transition: color 0.3s;
}

.toggle-btn:hover {
  color: #00cec9;
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
