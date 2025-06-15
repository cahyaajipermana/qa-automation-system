import { createRouter, createWebHistory } from 'vue-router'
import ResultDetail from '../pages/ResultDetail.vue'
import Dashboard from '../pages/Dashboard.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/results/:id',
    name: 'ResultDetail',
    component: ResultDetail,
    props: true
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router 