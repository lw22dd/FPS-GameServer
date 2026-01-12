import { defineStore } from 'pinia'
import { ref } from 'vue'
import CryptoUtil from '../utils/cryptoUtil'

export interface Message {
    type: string
    payload?: any
}

export interface RoomInfo {
    id: string
    name: string
    host: string
    players: string[]
    maxPlayers: number
    status: string
}

export interface GameState {
    hero1: HeroState
    hero2: HeroState
    bullets: BulletState[]
    status: string
}

export interface HeroState {
    id: string
    x: number
    y: number
    hp: number
    direction: number
    alive: boolean
}

export interface BulletState {
    id: string
    x: number
    y: number
    vx: number
    ownerId: string
}

export const useSocketStore = defineStore('socket', () => {
    // 状态
    const ws = ref<WebSocket | null>(null)
    const heartbeatTimer = ref<number | null>(null)
    const reconnectTimer = ref<number | null>(null)
    
    const connected = ref(false)
    const username = ref('')
    const currentRoom = ref<RoomInfo | null>(null)
    const gameState = ref<GameState | null>(null)
    const gameStarted = ref(false)
    const gameOver = ref(false)
    const winner = ref('')
    
    const messageHandlers = ref<Map<string, Function[]>>(new Map())
    
    // 连接WebSocket
    function connect(userName: string): Promise<void> {
        return new Promise((resolve, reject) => {
            username.value = userName
            
            const wsUrl = `ws://localhost:8080/ws?username=${encodeURIComponent(userName)}`
            ws.value = new WebSocket(wsUrl)
            
            ws.value.onopen = () => {
                console.log('WebSocket已连接')
                connected.value = true
                startHeartbeat()
                emit('connected', { userName })
                resolve()
            }
            
            ws.value.onclose = () => {
                console.log('WebSocket已断开')
                connected.value = false
                stopHeartbeat()
                scheduleReconnect()
                emit('disconnected', {})
            }
            
            ws.value.onerror = (error) => {
                console.error('WebSocket错误:', error)
                reject(error)
            }
            
            ws.value.onmessage = async (event) => {
                try {
                    const encrypted = event.data as string
                    const plaintext = await CryptoUtil.decrypt(encrypted)
                    const message: Message = JSON.parse(plaintext)
                    handleMessage(message)
                } catch (error) {
                    console.error('消息解析或解密错误:', error)
                }
            }
        })
    }
    
    // 断开连接
    function disconnect() {
        if (reconnectTimer.value) {
            clearTimeout(reconnectTimer.value)
            reconnectTimer.value = null
        }
        stopHeartbeat()
        if (ws.value) {
            ws.value.close()
            ws.value = null
        }
        connected.value = false
        currentRoom.value = null
        gameStarted.value = false
        gameOver.value = false
    }
    
    // 开始心跳
    function startHeartbeat() {
        heartbeatTimer.value = window.setInterval(() => {
            send({ type: 'heartbeat' })
        }, 2000) as unknown as number
    }
    
    // 停止心跳
    function stopHeartbeat() {
        if (heartbeatTimer.value) {
            clearInterval(heartbeatTimer.value)
            heartbeatTimer.value = null
        }
    }
    
    // 计划重连
    function scheduleReconnect() {
        if (reconnectTimer.value) return
        reconnectTimer.value = window.setTimeout(() => {
            if (username.value) {
                console.log('尝试重新连接...')
                connect(username.value).catch(() => {})
            }
            reconnectTimer.value = null
        }, 5000) as unknown as number
    }
    
    // 发送消息
    async function send(message: Message) {
        if (ws.value && ws.value.readyState === WebSocket.OPEN) {
            const plaintext = JSON.stringify(message)
            const encrypted = await CryptoUtil.encrypt(plaintext)
            ws.value.send(encrypted)
        }
    }
    
    // 处理消息
    function handleMessage(message: Message) {
        console.log('收到消息:', message.type, message.payload)
        
        switch (message.type) {
            case 'heartbeat_reply':
                break
                
            case 'room_list':
                emit('roomList', message.payload.rooms)
                break
                
            case 'join_room_result':
                if (message.payload.success) {
                    currentRoom.value = message.payload.room
                    emit('joinedRoom', message.payload.room)
                } else {
                    emit('joinError', message.payload.message)
                }
                break
                
            case 'game_start':
                gameStarted.value = true
                gameOver.value = false
                currentRoom.value = message.payload
                initGameState()
                emit('gameStart', message.payload)
                break
                
            case 'game_state':
                updateGameState(message.payload)
                emit('gameStateUpdate', message.payload)
                break
                
            case 'player_action':
                handlePlayerAction(message.payload)
                emit('playerAction', message.payload)
                break
                
            case 'fire':
                handleFireAction(message.payload)
                emit('fireAction', message.payload)
                break
                
            case 'hit':
                handleHitAction(message.payload)
                emit('hitAction', message.payload)
                break
                
            case 'death':
                handleDeathAction(message.payload)
                emit('deathAction', message.payload)
                break
                
            case 'game_over':
                gameOver.value = true
                winner.value = message.payload.winner
                emit('gameOver', message.payload)
                break
                
            case 'error':
                emit('error', message.payload)
                break
        }
        
        const handlers = messageHandlers.value.get(message.type) || []
        handlers.forEach(handler => handler(message.payload))
    }
    
    // 初始化游戏状态
    function initGameState() {
        gameState.value = {
            hero1: { id: '', x: 50, y: 170, hp: 5, direction: 1, alive: true },
            hero2: { id: '', x: 730, y: 170, hp: 5, direction: -1, alive: true },
            bullets: [],
            status: 'playing'
        }
    }
    
    // 更新游戏状态
    function updateGameState(state: GameState) {
        gameState.value = state
    }
    
    // 处理玩家动作
    function handlePlayerAction(payload: any) {
        if (!gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const action = payload.action
        const value = payload.value
        const isHero1 = gameState.value.hero1.id === playerId || 
                       (!gameState.value.hero1.id && playerId === 'player1')
        const hero = isHero1 ? gameState.value.hero1 : gameState.value.hero2
        
        switch (action) {
            case 'move_y':
                hero.y = value
                break
            case 'direction':
                hero.direction = value
                break
        }
    }
    
    // 处理射击动作
    function handleFireAction(payload: any) {
        if (!gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const bulletId = payload.bullet_id || payload.bulletId
        const direction = payload.direction
        
        gameState.value.bullets.push({
            id: bulletId || Date.now().toString(),
            x: payload.x,
            y: payload.y,
            vx: direction * 7,
            ownerId: playerId
        })
    }
    
    // 处理击中动作
    function handleHitAction(payload: any) {
        if (!gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const targetId = payload.target_id || payload.targetId
        const remaining = payload.remaining
        const isHero1 = gameState.value.hero1.id === targetId || 
                       (!gameState.value.hero1.id && targetId === 'player1')
        const hero = isHero1 ? gameState.value.hero1 : gameState.value.hero2
        hero.hp = remaining
    }
    
    // 处理死亡动作
    function handleDeathAction(payload: any) {
        if (!gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const isHero1 = gameState.value.hero1.id === playerId || 
                       (!gameState.value.hero1.id && playerId === 'player1')
        const hero = isHero1 ? gameState.value.hero1 : gameState.value.hero2
        hero.alive = false
    }
    
    // 注册消息处理器
    function on(type: string, handler: Function) {
        const handlers = messageHandlers.value.get(type) || []
        handlers.push(handler)
        messageHandlers.value.set(type, handlers)
    }
    
    // 移除消息处理器
    function off(type: string, handler: Function) {
        const handlers = messageHandlers.value.get(type) || []
        const index = handlers.indexOf(handler)
        if (index > -1) {
            handlers.splice(index, 1)
        }
    }
    
    // 触发消息
    function emit(type: string, payload: any) {
        const handlers = messageHandlers.value.get(type) || []
        handlers.forEach(handler => handler(payload))
    }
    
    // 创建房间
    function createRoom(name: string, maxPlayers: number = 2) {
        send({
            type: 'create_room',
            payload: { name, max_players: maxPlayers }
        })
    }
    
    // 加入房间
    function joinRoom(roomId: string) {
        send({
            type: 'join_room',
            payload: { room_id: roomId }
        })
    }
    
    // 获取房间列表
    function getRoomList() {
        send({ type: 'room_list' })
    }
    
    // 开始游戏
    function startGame() {
        send({ type: 'start_game' })
    }
    
    // 发送玩家动作
    function sendPlayerAction(action: string, value: number) {
        send({
            type: 'player_action',
            payload: {
                player_id: username.value,
                action,
                value
            }
        })
    }
    
    // 发送射击指令
    function sendFire(direction: number, x: number, y: number) {
        send({
            type: 'fire',
            payload: {
                player_id: username.value,
                direction,
                bullet_id: `bullet_${Date.now()}`,
                x,
                y
            }
        })
    }
    
    // 发送击中指令
    function sendHit(targetId: string, damage: number, remaining: number) {
        send({
            type: 'hit',
            payload: {
                target_id: targetId,
                damage,
                remaining
            }
        })
    }
    
    // 发送死亡指令
    function sendDeath(playerId: string) {
        send({
            type: 'death',
            payload: { player_id: playerId }
        })
    }
    
    // 发送游戏结束指令
    function sendGameOver(winner: string, loser: string, duration: number) {
        send({
            type: 'game_over',
            payload: { winner, loser, duration }
        })
    }
    
    return {
        // 状态
        connected,
        username,
        currentRoom,
        gameState,
        gameStarted,
        gameOver,
        winner,
        
        // 方法
        connect,
        disconnect,
        send,
        on,
        off,
        createRoom,
        joinRoom,
        getRoomList,
        startGame,
        sendPlayerAction,
        sendFire,
        sendHit,
        sendDeath,
        sendGameOver
    }
})