<template>
  <div class="rooms-container">
    <div class="rooms-header">
      <h1>🎮 游戏大厅</h1>
      <div class="user-info">
        <span>玩家: {{ username }}</span>
        <button @click="logout" class="logout-btn">退出</button>
      </div>
    </div>
    
    <div class="rooms-content">
      <div class="create-room-section">
        <h2>创建房间</h2>
        <div class="create-form">
          <input 
            v-model="newRoomName" 
            placeholder="房间名称"
            maxlength="30"
          />
          <button @click="createRoom" :disabled="!canCreateRoom">
            创建房间
          </button>
        </div>
      </div>
      
      <div class="room-list-section">
        <h2>房间列表</h2>
        <button @click="refreshRooms" class="refresh-btn">刷新列表</button>
        
        <div class="room-list">
          <div 
            v-for="room in rooms" 
            :key="room.id" 
            class="room-item"
            :class="{ 'full': room.players.length >= room.maxPlayers }"
          >
            <div class="room-info">
              <span class="room-name">{{ room.name }}</span>
              <span class="room-players">
                {{ room.players.length }}/{{ room.maxPlayers }} 玩家
              </span>
              <span class="room-host">房主: {{ room.host }}</span>
              <span class="room-status" :class="room.status">
                {{ getStatusText(room.status) }}
              </span>
            </div>
            <button 
              @click="joinRoom(room.id)"
              :disabled="room.players.length >= room.maxPlayers || room.status === 'playing'"
            >
              {{ room.status === 'playing' ? '游戏中' : '加入' }}
            </button>
          </div>
          
          <div v-if="rooms.length === 0" class="no-rooms">
            暂无房间，请创建一个
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="error" class="error-toast">
      {{ error }}
      <button @click="error = ''">关闭</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSocketStore, RoomInfo } from '@/stores/socketStore'

const router = useRouter()
const socketStore = useSocketStore()
const username = ref(socketStore.username)
const rooms = ref<RoomInfo[]>([])
const newRoomName = ref('')
const error = ref('')

const canCreateRoom = computed(() => {
  const name = newRoomName.value.trim()
  return name.length >= 2 && name.length <= 30
})

onMounted(() => {
  if (!socketStore.connected) {
    router.push('/')
    return
  }
  
  refreshRooms()
  
  socketStore.on('roomList', (roomList: RoomInfo[]) => {
    rooms.value = roomList
  })
  
  socketStore.on('joinedRoom', (room: RoomInfo) => {
    router.push('/game')
  })
  
  socketStore.on('joinError', (message: string) => {
    error.value = message
  })
})

onUnmounted(() => {
  socketStore.off('roomList', () => {})
  socketStore.off('joinedRoom', () => {})
  socketStore.off('joinError', () => {})
})

function refreshRooms() {
  socketStore.getRoomList()
}

function createRoom() {
  if (!canCreateRoom.value) return
  socketStore.createRoom(newRoomName.value.trim())
  newRoomName.value = ''
}

function joinRoom(roomId: string) {
  socketStore.joinRoom(roomId)
}

function logout() {
  socketStore.disconnect()
  router.push('/')
}

function getStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    'waiting': '等待中',
    'ready': '准备就绪',
    'playing': '游戏中',
    'ended': '已结束'
  }
  return statusMap[status] || status
}
</script>

<style scoped>
.rooms-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  padding: 20px;
}

.rooms-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  color: #e94560;
  margin: 0;
}

h2 {
  color: #0f9b8e;
  margin: 0 0 15px 0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 15px;
  color: white;
}

.logout-btn {
  padding: 8px 16px;
  background: #e94560;
  border: none;
  border-radius: 5px;
  color: white;
  cursor: pointer;
}

.rooms-content {
  max-width: 1200px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 20px;
  padding: 20px;
}

.create-room-section,
.room-list-section {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 15px;
  padding: 20px;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.create-form input {
  padding: 12px;
  border: 2px solid #0f3460;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.3);
  color: white;
  font-size: 14px;
}

.create-form input:focus {
  outline: none;
  border-color: #e94560;
}

.create-form button {
  padding: 12px;
  background: linear-gradient(135deg, #0f9b8e 0%, #0f3460 100%);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 16px;
  cursor: pointer;
  transition: transform 0.2s;
}

.create-form button:hover:not(:disabled) {
  transform: translateY(-2px);
}

.create-form button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn {
  padding: 8px 16px;
  background: #0f3460;
  border: none;
  border-radius: 5px;
  color: white;
  cursor: pointer;
  margin-bottom: 15px;
}

.room-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 10px;
  border: 1px solid #0f3460;
}

.room-item.full {
  opacity: 0.7;
}

.room-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.room-name {
  color: #e94560;
  font-weight: bold;
  font-size: 16px;
}

.room-players,
.room-host {
  color: #aaa;
  font-size: 12px;
}

.room-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 3px;
  display: inline-block;
  width: fit-content;
}

.room-status.waiting {
  background: #4CAF50;
}

.room-status.playing {
  background: #e94560;
}

.room-item button {
  padding: 8px 20px;
  background: #0f9b8e;
  border: none;
  border-radius: 5px;
  color: white;
  cursor: pointer;
}

.room-item button:disabled {
  background: #666;
  cursor: not-allowed;
}

.no-rooms {
  text-align: center;
  color: #666;
  padding: 40px;
}

.error-toast {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: #e94560;
  color: white;
  padding: 15px 25px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 15px;
}

.error-toast button {
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
}
</style>
