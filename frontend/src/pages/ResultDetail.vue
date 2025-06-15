<template>
  <div class="container mx-auto px-4 py-4">
    <div v-if="loading" class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>

    <div v-else-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
      <strong class="font-bold">Error!</strong>
      <span class="block sm:inline">{{ error }}</span>
    </div>

    <div v-else class="bg-white shadow-lg rounded-lg overflow-hidden">
      <!-- Header -->
      <div class="bg-gray-50 px-6 py-4 border-b">
        <div class="flex justify-between items-center">
          <div>
            <h1 class="text-2xl text-left font-bold text-gray-900">Test Result Details</h1>
            <p class="text-sm text-left text-gray-600">Created At: {{ result?.created_at ? new Date(result.created_at).toLocaleString() : 'N/A' }}</p>
          </div>
          <div class="flex items-center space-x-4">
            <button
              @click="fetchResultDetails"
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg class="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Refresh
            </button>
            <router-link
              to="/"
              class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg class="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to Dashboard
            </router-link>
          </div>
        </div>
      </div>

      <!-- Main Content -->
      <div class="p-6">
        <!-- Test Information -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div class="space-y-4">
            <div>
              <h3 class="text-sm font-medium text-gray-500">Site</h3>
              <p class="mt-1 text-lg text-gray-900">{{ result.site?.name }}</p>
            </div>
            <div>
              <h3 class="text-sm font-medium text-gray-500">Device</h3>
              <p class="mt-1 text-lg text-gray-900">{{ result.device?.name }}</p>
            </div>
            <div>
              <h3 class="text-sm font-medium text-gray-500">Browser</h3>
              <p class="mt-1 text-lg text-gray-900">{{ result.browser }}</p>
            </div>
          </div>
          <div class="space-y-4">
            <div>
              <h3 class="text-sm font-medium text-gray-500">Feature</h3>
              <p class="mt-1 text-lg text-gray-900">{{ result.feature?.name }}</p>
            </div>
            <div>
              <h3 class="text-sm font-medium text-gray-500">Duration</h3>
              <p class="mt-1 text-lg text-gray-900">{{ result?.duration ? result?.duration.toFixed(2) + 's' : 'N/A' }}</p>
            </div>
            <div>
              <h3 class="text-sm font-medium text-gray-500">Status</h3>
              <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
                :class="{
                  'bg-pass/10 text-pass': result?.status === 'passed',
                  'bg-warning/10 text-warning': result?.status === 'warning' || result?.status === 'processing',
                  'bg-fail/10 text-fail': result?.status === 'failed'
              }">
                {{ result.status }}
              </span>
            </div>
          </div>
        </div>

        <!-- Screenshots -->
        <div class="mt-8">
          <h2 class="text-xl font-semibold text-gray-900 mb-4">Error Log</h2>
          <div class="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-1 gap-6">
            <div class="bg-gray-50 rounded-lg overflow-hidden text-black py-4">
              {{ result.error_log || 'N/A' }}
            </div>
          </div>
        </div>

        <!-- Test Steps -->
        <div class="mt-8">
          <h2 class="text-xl font-semibold text-gray-900 mb-4">
            Test Steps
            <div class="text-sm text-red-500 mb-4">
                <small><i>*) Click the image to view in full width</i></small>
            </div>
          </h2>
          <div class="bg-gray-50 rounded-lg p-4">
            <template v-for="(step, index) in result.details" :key="index">
              <div class="flex items-start space-x-3 py-2">
                <span class="flex-shrink-0 w-8 h-8 flex items-center justify-center rounded-full bg-blue-100 text-blue-600">
                  {{ index + 1 }}
                </span>
                <div class="w-full">
                  <p class="text-gray-700 pb-4 text-left">{{ step.description }}</p>
                  <div class="w-full">
                    <Image :imageUrl="`${apiUrl}/${step.screenshot}`"></Image>
                  </div>
                </div>
              </div>
              <hr class="my-4">
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>

import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import Image from '../components/Image.vue'

const route = useRoute()
const result = ref(null)
const loading = ref(true)
const error = ref(null)
const apiUrl = ref(import.meta.env.VITE_API_URL)

const fetchResultDetails = async () => {
  try {
    loading.value = true
    const response = await axios.get(`/api/results/${route.params.id}`)
    result.value = response.data
  } catch (err) {
    error.value = 'Failed to load result details. Please try again later.'
    console.error('Error fetching result details:', err)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateString) => {
  if (!dateString) return 'N/A'
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  fetchResultDetails()
})
  
</script> 