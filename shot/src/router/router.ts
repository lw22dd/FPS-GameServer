import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import RoomView from '@/views/RoomView.vue'
import ShootingGame from '@/components/ShootingGame.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', component: LoginView },
  { path: '/rooms', component: RoomView },
  { path: '/game', component: ShootingGame },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
