import { createRouter, createWebHashHistory } from 'vue-router'
import OpenDatabase from '../views/OpenDatabase.vue'
import DataManager from '../views/DataManager.vue'

const routes = [
  {
    path: '/',
    name: 'OpenDatabase',
    component: OpenDatabase
  },
  {
    path: '/manager',
    name: 'DataManager',
    component: DataManager
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
