<template>
  <div class="game-container">
    <div class="game-header">
      <div class="player-info p1-info">
        <span class="player-name">{{ player1Name || '玩家1' }}</span>
        <div class="hp-bar">
          <div class="hp-fill" :style="{ width: (p1.hp / maxHP * 100) + '%' }"></div>
        </div>
        <span class="hp-text">{{ p1.hp }}/{{ maxHP }}</span>
      </div>
      
      <div class="game-status">
        <span v-if="!socketService.gameStarted.value" class="waiting">
          等待游戏开始...
        </span>
        <span v-else-if="socketService.gameOver.value" class="game-over">
          🏆 {{ socketService.winner.value }} 获胜!
        </span>
        <span v-else class="playing">
          对战中
        </span>
      </div>
      
      <div class="player-info p2-info">
        <span class="player-name">{{ player2Name || '玩家2' }}</span>
        <div class="hp-bar">
          <div class="hp-fill" :style="{ width: (p2.hp / maxHP * 100) + '%' }"></div>
        </div>
        <span class="hp-text">{{ p2.hp }}/{{ maxHP }}</span>
      </div>
    </div>

    <canvas ref="gameCanvas" width="800" height="400"></canvas>

    <!-- 开始游戏按钮，只有房主可见 -->
    <div v-if="isHost && !socketService.gameStarted.value" class="start-game-container">
      <button @click="startGame" class="start-game-btn">开始游戏</button>
    </div>

    <div class="controls">
      <span>移动: W/S 或 ↑/↓ | 射击: D 或 ← | 离开: Esc</span>
    </div>

    <div v-if="socketService.gameOver.value" class="game-over-overlay">
      <div class="game-over-content">
        <h2>🎉 游戏结束</h2>
        <p>获胜者: {{ socketService.winner.value }}</p>
        <button @click="backToRooms" class="btn">返回房间</button>
      </div>
    </div>
    
    <div v-if="!socketService.connected.value" class="connection-error">
      <p>连接已断开，正在重连...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import socketService from '@/services/socketService'

const router = useRouter()
const gameCanvas = ref<HTMLCanvasElement | null>(null)
let ctx: CanvasRenderingContext2D | null = null
let animationId: number | null = null

const playerWidth = 20
const playerHeight = 60
const bulletSize = 5
const moveSpeed = 5
const bulletSpeed = 7
const maxHP = 5

const player1Name = ref('')
const player2Name = ref('')

const p1 = ref({
  x: 50,
  y: 200 - playerHeight / 2,
  hp: maxHP,
  color: '#00D2FF',
  bullets: [] as { x: number; y: number; vx: number }[]
})

const p2 = ref({
  x: 800 - 50 - playerWidth,
  y: 200 - playerHeight / 2,
  hp: maxHP,
  color: '#FF3D67',
  bullets: [] as { x: number; y: number; vx: number }[]
})

const keys: Record<string, boolean> = {}
let lastFireTime = 0
const fireCooldown = 500

// 判断当前玩家是否是房主
const isHost = computed(() => {
  const room = socketService.currentRoom.value
  return room && room.host === socketService.username.value
})

// 开始游戏方法
function startGame() {
  socketService.startGame()
}

onMounted(() => {
  if (!socketService.connected.value) {
    router.push('/rooms')
    return
  }
  
  if (gameCanvas.value) {
    ctx = gameCanvas.value.getContext('2d')
    setupEventListeners()
    setupSocketListeners()
    startGameLoop()
    initGameState()
  }
})

onUnmounted(() => {
  removeEventListeners()
  cleanupSocketListeners()
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
})

function setupEventListeners() {
  window.addEventListener('keydown', handleKeydown)
  window.addEventListener('keyup', handleKeyup)
}

function removeEventListeners() {
  window.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('keyup', handleKeyup)
}

function setupSocketListeners() {
  socketService.on('gameStart', (room: any) => {
    if (room.players && room.players.length >= 2) {
      player1Name.value = room.players[0]
      player2Name.value = room.players[1]
    }
    resetGame()
    initGameState()
  })
  
  socketService.on('gameStateUpdate', (state: any) => {
    updateGameState(state)
  })
  
  socketService.on('playerAction', (action: any) => {
    handleRemotePlayerAction(action)
  })
  
  socketService.on('fireAction', (action: any) => {
    handleRemoteFire(action)
  })
  
  socketService.on('hitAction', (action: any) => {
    handleRemoteHit(action)
  })
  
  socketService.on('deathAction', (action: any) => {
    handleRemoteDeath(action)
  })
  
  socketService.on('gameOver', (result: any) => {
    console.log('游戏结束:', result)
  })
  
  socketService.on('disconnected', () => {
    router.push('/rooms')
  })
}

function cleanupSocketListeners() {
  socketService.off('gameStart', () => {})
  socketService.off('gameStateUpdate', () => {})
  socketService.off('playerAction', () => {})
  socketService.off('fireAction', () => {})
  socketService.off('hitAction', () => {})
  socketService.off('deathAction', () => {})
  socketService.off('gameOver', () => {})
  socketService.off('disconnected', () => {})
}

function initGameState() {
  const username = socketService.username.value
  const room = socketService.currentRoom.value
  
  if (room && room.players) {
    player1Name.value = room.players[0] || ''
    player2Name.value = room.players[1] || ''
  }
  
  // 固定玩家1在左侧，玩家2在右侧，无论当前登录的用户是谁
  p1.value.x = 50
  p1.value.y = 200 - playerHeight / 2
  p1.value.hp = maxHP
  p1.value.bullets = []
  
  p2.value.x = 800 - 50 - playerWidth
  p2.value.y = 200 - playerHeight / 2
  p2.value.hp = maxHP
  p2.value.bullets = []
}

function resetGame() {
  p1.value.x = 50
  p1.value.y = 200 - playerHeight / 2
  p1.value.hp = maxHP
  p1.value.bullets = []
  
  p2.value.x = 800 - 50 - playerWidth
  p2.value.y = 200 - playerHeight / 2
  p2.value.hp = maxHP
  p2.value.bullets = []
}

function handleKeydown(e: KeyboardEvent) {
  keys[e.key] = true
  
  if (e.key === 'Escape') {
    backToRooms()
    return
  }
  
  if (!socketService.gameStarted.value || socketService.gameOver.value) return
  
  const now = Date.now()
  const isPlayer1 = player1Name.value === socketService.username.value
  const myPlayer = isPlayer1 ? p1.value : p2.value
  const opponent = isPlayer1 ? p2.value : p1.value
  
  if ((e.key === 'd' || e.key === 'D') && isPlayer1) {
    if (now - lastFireTime >= fireCooldown) {
      fire(myPlayer, 1)
      lastFireTime = now
    }
  }
  
  if (e.key === 'ArrowLeft' && !isPlayer1) {
    if (now - lastFireTime >= fireCooldown) {
      fire(myPlayer, -1)
      lastFireTime = now
    }
  }
}

function handleKeyup(e: KeyboardEvent) {
  keys[e.key] = false
}

function fire(player: typeof p1.value, direction: number) {
  if (player.hp <= 0) return
  
  const bullet = {
    x: direction === 1 ? player.x + playerWidth : player.x,
    y: player.y + playerHeight / 2,
    vx: direction * bulletSpeed
  }
  
  player.bullets.push(bullet)
  
  socketService.sendFire(direction, bullet.x, bullet.y)
}

function update() {
  if (!socketService.gameStarted.value || socketService.gameOver.value) return
  
  const isPlayer1 = player1Name.value === socketService.username.value
  const myPlayer = isPlayer1 ? p1.value : p2.value
  const opponent = isPlayer1 ? p2.value : p1.value
  
  let moved = false
  let yChanged = 0
  
  if (isPlayer1) {
    if (keys['w'] || keys['W']) {
      myPlayer.y -= moveSpeed
      yChanged = -moveSpeed
      moved = true
    }
    if (keys['s'] || keys['S']) {
      myPlayer.y += moveSpeed
      yChanged = moveSpeed
      moved = true
    }
  } else {
    if (keys['ArrowUp']) {
      myPlayer.y -= moveSpeed
      yChanged = -moveSpeed
      moved = true
    }
    if (keys['ArrowDown']) {
      myPlayer.y += moveSpeed
      yChanged = moveSpeed
      moved = true
    }
  }
  
  if (moved && yChanged !== 0) {
    myPlayer.y = Math.max(0, Math.min(400 - playerHeight, myPlayer.y))
    socketService.sendPlayerAction('move_y', myPlayer.y)
  }
  
  updateBullets(myPlayer, opponent, true)
  updateBullets(opponent, myPlayer, false)
  
  checkGameOver()
}

function updateBullets(
  attacker: typeof p1.value,
  target: typeof p2.value,
  isLocal: boolean
) {
  for (let i = attacker.bullets.length - 1; i >= 0; i--) {
    const b = attacker.bullets[i]
    b.x += b.vx
    
    if (
      b.x > target.x &&
      b.x < target.x + playerWidth &&
      b.y > target.y &&
      b.y < target.y + playerHeight
    ) {
      attacker.bullets.splice(i, 1)
      target.hp--
      
      if (isLocal) {
        const targetId = target === p1.value ? player1Name.value : player2Name.value
        socketService.sendHit(targetId, 1, target.hp)
      }
      
      if (target.hp <= 0) {
        handleDeath(target === p1.value ? player1Name.value : player2Name.value)
      }
      continue
    }
    
    if (b.x < 0 || b.x > 800) {
      attacker.bullets.splice(i, 1)
    }
  }
}

function handleDeath(playerId: string) {
  socketService.sendDeath(playerId)
  
  const winner = playerId === player1Name.value ? player2Name.value : player1Name.value
  socketService.sendGameOver(winner, playerId, 0)
}

function checkGameOver() {
  if (p1.value.hp <= 0 || p2.value.hp <= 0) {
    if (p2.value.hp <= 0) {
      socketService.sendGameOver(player1Name.value, player2Name.value, 0)
    } else {
      socketService.sendGameOver(player2Name.value, player1Name.value, 0)
    }
  }
}

function handleRemotePlayerAction(action: any) {
  // 同时支持下划线和驼峰命名的字段
  const playerId = action.player_id || action.playerId
  const actionType = action.action
  const value = action.value
  
  // 确保player1Name和player2Name已经初始化
  if (!player1Name.value || !player2Name.value) {
    return
  }
  
  if (playerId === player2Name.value) {
    if (actionType === 'move_y') {
      p2.value.y = value
    }
  } else if (playerId === player1Name.value) {
    if (actionType === 'move_y') {
      p1.value.y = value
    }
  }
}

function handleRemoteFire(action: any) {
  // 同时支持下划线和驼峰命名的字段
  const playerId = action.player_id || action.playerId
  const direction = action.direction
  const x = action.x
  const y = action.y
  
  // 确保player1Name和player2Name已经初始化
  if (!player1Name.value || !player2Name.value) {
    return
  }
  
  const isPlayer1Fire = playerId === player1Name.value
  const player = isPlayer1Fire ? p1.value : p2.value
  
  player.bullets.push({
    x,
    y,
    vx: direction * bulletSpeed
  })
}

function handleRemoteHit(action: any) {
  // 同时支持下划线和驼峰命名的字段
  const targetId = action.target_id || action.targetId
  const remaining = action.remaining
  
  // 确保player1Name和player2Name已经初始化
  if (!player1Name.value || !player2Name.value) {
    return
  }
  
  if (targetId === player1Name.value) {
    p1.value.hp = remaining
  } else if (targetId === player2Name.value) {
    p2.value.hp = remaining
  }
}

function handleRemoteDeath(action: any) {
  // 同时支持下划线和驼峰命名的字段
  const playerId = action.player_id || action.playerId
  
  // 确保player1Name和player2Name已经初始化
  if (!player1Name.value || !player2Name.value) {
    return
  }
  
  if (playerId === player1Name.value) {
    p1.value.hp = 0
  } else if (playerId === player2Name.value) {
    p2.value.hp = 0
  }
}

function updateGameState(state: any) {
  if (!state) return
  
  if (state.hero1) {
    if (state.hero1.alive !== undefined) {
      if (!state.hero1.alive && p1.value.hp > 0) {
        p1.value.hp = 0
      }
    }
    if (state.hero1.hp !== undefined && state.hero1.hp < p1.value.hp) {
      p1.value.hp = state.hero1.hp
    }
  }
  
  if (state.hero2) {
    if (state.hero2.alive !== undefined) {
      if (!state.hero2.alive && p2.value.hp > 0) {
        p2.value.hp = 0
      }
    }
    if (state.hero2.hp !== undefined && state.hero2.hp < p2.value.hp) {
      p2.value.hp = state.hero2.hp
    }
  }
}

function draw() {
  if (!ctx) return
  
  ctx.clearRect(0, 0, 800, 400)
  
  ctx.fillStyle = '#1a1a2e'
  ctx.fillRect(0, 0, 800, 400)
  
  drawPlayer(p1.value)
  drawPlayer(p2.value)
  
  ctx.fillStyle = '#ffffff'
  p1.value.bullets.forEach(b => {
    ctx?.fillRect(b.x, b.y, bulletSize, bulletSize)
  })
  p2.value.bullets.forEach(b => {
    ctx?.fillRect(b.x, b.y, bulletSize, bulletSize)
  })
}

function drawPlayer(player: typeof p1.value) {
  if (!ctx) return
  
  ctx.fillStyle = player.color
  ctx.fillRect(player.x, player.y, playerWidth, playerHeight)
  
  if (player.hp <= 0) {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.5)'
    ctx.fillRect(player.x, player.y, playerWidth, playerHeight)
    
    ctx.strokeStyle = '#ff0000'
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(player.x, player.y)
    ctx.lineTo(player.x + playerWidth, player.y + playerHeight)
    ctx.moveTo(player.x + playerWidth, player.y)
    ctx.lineTo(player.x, player.y + playerHeight)
    ctx.stroke()
  }
}

function startGameLoop() {
  function loop() {
    if (socketService.gameStarted.value) {
      update()
    }
    draw()
    animationId = requestAnimationFrame(loop)
  }
  loop()
}

function backToRooms() {
  socketService.gameStarted.value = false
  socketService.gameOver.value = false
  router.push('/rooms')
}
</script>

<style scoped>
.game-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  padding: 20px;
  position: relative;
}

.game-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 800px;
  margin-bottom: 15px;
  padding: 15px 20px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.p1-info {
  color: #00D2FF;
}

.p2-info {
  color: #FF3D67;
}

.player-name {
  font-weight: bold;
  font-size: 14px;
  min-width: 60px;
}

.hp-bar {
  width: 100px;
  height: 10px;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 5px;
  overflow: hidden;
}

.hp-fill {
  height: 100%;
  background: linear-gradient(90deg, #00D2FF, #0f9b8e);
  transition: width 0.3s;
}

.p2-info .hp-fill {
  background: linear-gradient(90deg, #FF3D67, #e94560);
}

.hp-text {
  font-size: 12px;
  color: #aaa;
  min-width: 40px;
}

.game-status {
  text-align: center;
}

.game-status .waiting {
  color: #4CAF50;
  font-size: 18px;
}

.game-status .playing {
  color: #FF9800;
  font-size: 18px;
  animation: pulse 1s infinite;
}

.game-status .game-over {
  color: #FFD700;
  font-size: 24px;
  font-weight: bold;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

canvas {
  border: 4px solid #0f3460;
  border-radius: 10px;
  background: #1a1a2e;
  box-shadow: 0 0 30px rgba(15, 52, 96, 0.5);
}

.controls {
  margin-top: 15px;
  color: #666;
  font-size: 14px;
}

.game-over-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
}

.game-over-content {
  background: rgba(255, 255, 255, 0.1);
  padding: 40px 60px;
  border-radius: 20px;
  text-align: center;
  color: white;
}

.game-over-content h2 {
  color: #FFD700;
  margin-bottom: 20px;
}

.game-over-content p {
  font-size: 20px;
  margin-bottom: 30px;
}

.btn {
  padding: 12px 30px;
  background: linear-gradient(135deg, #e94560 0%, #0f3460 100%);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 16px;
  cursor: pointer;
  transition: transform 0.2s;
}

.btn:hover {
  transform: translateY(-2px);
}

.connection-error {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: #e94560;
  color: white;
  padding: 10px 20px;
  border-radius: 5px;
}

/* 开始游戏按钮样式 */
.start-game-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 10;
}

.start-game-btn {
  padding: 15px 40px;
  font-size: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.start-game-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
}

.start-game-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.4);
}
</style>
