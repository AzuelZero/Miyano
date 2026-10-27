import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/home',
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/home',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/exercises',
      name: 'catalog',
      component: () => import('../views/CatalogView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/exercises/:id',
      name: 'exercise-detail',
      component: () => import('../views/ExerciseDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/routines',
      name: 'routines',
      component: () => import('../views/RoutinesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/routines/new',
      name: 'routine-new',
      component: () => import('../views/RoutineEditorView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/routines/:id',
      name: 'routine-detail',
      component: () => import('../views/RoutineDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/routines/:id/edit',
      name: 'routine-edit',
      component: () => import('../views/RoutineEditorView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/home',
    },
  ],
})

// Route guard: protected pages need a session; auth pages are guest-only.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return '/login'
  if (to.meta.guestOnly && auth.isAuthenticated) return '/home'
})

export default router
