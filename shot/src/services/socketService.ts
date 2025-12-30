import { ref, readonly } from 'vue'

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

class SocketService {
    private ws: WebSocket | null = null
    private heartbeatTimer: number | null = null
    private reconnectTimer: number | null = null
    
    public connected = ref(false)
    public username = ref('')
    public currentRoom = ref<RoomInfo | null>(null)
    public gameState = ref<GameState | null>(null)
    public gameStarted = ref(false)
    public gameOver = ref(false)
    public winner = ref('')
    
    private messageHandlers: Map<string, Function[]> = new Map()
    
    connect(username: string): Promise<void> {
        return new Promise((resolve, reject) => {
            this.username.value = username
            
            const wsUrl = `ws://localhost:8080/ws?username=${encodeURIComponent(username)}`
            this.ws = new WebSocket(wsUrl)
            
            this.ws.onopen = () => {
                console.log('WebSocket已连接')
                this.connected.value = true
                this.startHeartbeat()
                this.emit('connected', { username })
                resolve()
            }
            
            this.ws.onclose = () => {
                console.log('WebSocket已断开')
                this.connected.value = false
                this.stopHeartbeat()
                this.scheduleReconnect()
                this.emit('disconnected', {})
            }
            
            this.ws.onerror = (error) => {
                console.error('WebSocket错误:', error)
                reject(error)
            }
            
            this.ws.onmessage = (event) => {
                try {
                    const message: Message = JSON.parse(event.data)
                    this.handleMessage(message)
                } catch (error) {
                    console.error('消息解析错误:', error)
                }
            }
        })
    }
    
    disconnect() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer)
            this.reconnectTimer = null
        }
        this.stopHeartbeat()
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this.connected.value = false
        this.currentRoom.value = null
        this.gameStarted.value = false
        this.gameOver.value = false
    }
    
    private startHeartbeat() {
        this.heartbeatTimer = window.setInterval(() => {
            this.send({ type: 'heartbeat' })
        }, 2000)
    }
    
    private stopHeartbeat() {
        if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer)
            this.heartbeatTimer = null
        }
    }
    
    private scheduleReconnect() {
        if (this.reconnectTimer) return
        this.reconnectTimer = window.setTimeout(() => {
            if (this.username.value) {
                console.log('尝试重新连接...')
                this.connect(this.username.value).catch(() => {})
            }
            this.reconnectTimer = null
        }, 5000)
    }
    
    send(message: Message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message))
        }
    }
    
    private handleMessage(message: Message) {
        console.log('收到消息:', message.type, message.payload)
        
        switch (message.type) {
            case 'heartbeat_reply':
                break
                
            case 'room_list':
                this.emit('roomList', message.payload.rooms)
                break
                
            case 'join_room_result':
                if (message.payload.success) {
                    this.currentRoom.value = message.payload.room
                    this.emit('joinedRoom', message.payload.room)
                } else {
                    this.emit('joinError', message.payload.message)
                }
                break
                
            case 'game_start':
                this.gameStarted.value = true
                this.gameOver.value = false
                this.currentRoom.value = message.payload
                this.initGameState()
                this.emit('gameStart', message.payload)
                break
                
            case 'game_state':
                this.updateGameState(message.payload)
                this.emit('gameStateUpdate', message.payload)
                break
                
            case 'player_action':
                this.handlePlayerAction(message.payload)
                this.emit('playerAction', message.payload)
                break
                
            case 'fire':
                this.handleFireAction(message.payload)
                this.emit('fireAction', message.payload)
                break
                
            case 'hit':
                this.handleHitAction(message.payload)
                this.emit('hitAction', message.payload)
                break
                
            case 'death':
                this.handleDeathAction(message.payload)
                this.emit('deathAction', message.payload)
                break
                
            case 'game_over':
                this.gameOver.value = true
                this.winner.value = message.payload.winner
                this.emit('gameOver', message.payload)
                break
                
            case 'error':
                this.emit('error', message.payload)
                break
        }
        
        const handlers = this.messageHandlers.get(message.type) || []
        handlers.forEach(handler => handler(message.payload))
    }
    
    private initGameState() {
        this.gameState.value = {
            hero1: { id: '', x: 50, y: 170, hp: 5, direction: 1, alive: true },
            hero2: { id: '', x: 730, y: 170, hp: 5, direction: -1, alive: true },
            bullets: [],
            status: 'playing'
        }
    }
    
    private updateGameState(state: GameState) {
        this.gameState.value = state
    }
    
    private handlePlayerAction(payload: any) {
        if (!this.gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const action = payload.action
        const value = payload.value
        const isHero1 = this.gameState.value.hero1.id === playerId || 
                       (!this.gameState.value.hero1.id && playerId === 'player1')
        const hero = isHero1 ? this.gameState.value.hero1 : this.gameState.value.hero2
        
        switch (action) {
            case 'move_y':
                hero.y = value
                break
            case 'direction':
                hero.direction = value
                break
        }
    }
    
    private handleFireAction(payload: any) {
        if (!this.gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const bulletId = payload.bullet_id || payload.bulletId
        const direction = payload.direction
        
        this.gameState.value.bullets.push({
            id: bulletId || Date.now().toString(),
            x: payload.x,
            y: payload.y,
            vx: direction * 7,
            ownerId: playerId
        })
    }
    
    private handleHitAction(payload: any) {
        if (!this.gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const targetId = payload.target_id || payload.targetId
        const remaining = payload.remaining
        const isHero1 = this.gameState.value.hero1.id === targetId || 
                       (!this.gameState.value.hero1.id && targetId === 'player1')
        const hero = isHero1 ? this.gameState.value.hero1 : this.gameState.value.hero2
        hero.hp = remaining
    }
    
    private handleDeathAction(payload: any) {
        if (!this.gameState.value) return
        
        // 同时支持下划线和驼峰命名的字段
        const playerId = payload.player_id || payload.playerId
        const isHero1 = this.gameState.value.hero1.id === playerId || 
                       (!this.gameState.value.hero1.id && playerId === 'player1')
        const hero = isHero1 ? this.gameState.value.hero1 : this.gameState.value.hero2
        hero.alive = false
    }
    
    on(type: string, handler: Function) {
        const handlers = this.messageHandlers.get(type) || []
        handlers.push(handler)
        this.messageHandlers.set(type, handlers)
    }
    
    off(type: string, handler: Function) {
        const handlers = this.messageHandlers.get(type) || []
        const index = handlers.indexOf(handler)
        if (index > -1) {
            handlers.splice(index, 1)
        }
    }
    
    private emit(type: string, payload: any) {
        const handlers = this.messageHandlers.get(type) || []
        handlers.forEach(handler => handler(payload))
    }
    
    createRoom(name: string, maxPlayers: number = 2) {
        this.send({
            type: 'create_room',
            payload: { name, max_players: maxPlayers }
        })
    }
    
    joinRoom(roomId: string) {
        this.send({
            type: 'join_room',
            payload: { room_id: roomId }
        })
    }
    
    getRoomList() {
        this.send({ type: 'room_list' })
    }
    
    startGame() {
        this.send({ type: 'start_game' })
    }
    
    sendPlayerAction(action: string, value: number) {
        this.send({
            type: 'player_action',
            payload: {
                player_id: this.username.value,
                action,
                value
            }
        })
    }
    
    sendFire(direction: number, x: number, y: number) {
        this.send({
            type: 'fire',
            payload: {
                player_id: this.username.value,
                direction,
                bullet_id: `bullet_${Date.now()}`,
                x,
                y
            }
        })
    }
    
    sendHit(targetId: string, damage: number, remaining: number) {
        this.send({
            type: 'hit',
            payload: {
                target_id: targetId,
                damage,
                remaining
            }
        })
    }
    
    sendDeath(playerId: string) {
        this.send({
            type: 'death',
            payload: { player_id: playerId }
        })
    }
    
    sendGameOver(winner: string, loser: string, duration: number) {
        this.send({
            type: 'game_over',
            payload: { winner, loser, duration }
        })
    }
}

export const socketService = new SocketService()
export default socketService
