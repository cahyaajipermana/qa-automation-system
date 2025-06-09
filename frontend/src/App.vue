<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import axios from 'axios'

const sites = ref([])
const devices = ref([])
const features = ref([])
const testResults = ref([])
const lastUpdated = ref(new Date().toLocaleString())
const isLoading = ref(true)
const error = ref(null)
const showNewTestModal = ref(false)
const expandedResults = ref(new Set())
const apiUrl = ref(import.meta.env.VITE_API_URL)

// Pagination state
const pagination = ref({
    currentPage: 1,
    perPage: 10,
    total: 0,
    totalPages: 0,
    hasNext: false,
    hasPrev: false
})

const newTest = ref({
    site_id: null,
    device_id: null,
    feature_id: null
})

const filters = ref({
    site: '',
    device: '',
    feature: '',
    status: ''
})

// Computed properties for summary cards
const totalTests = computed(() => {
    return Array.isArray(testResults.value) ? testResults.value.length : 0
})

const passedTests = computed(() => {
    return Array.isArray(testResults.value)
        ? testResults.value.filter(r => r && r.status === 'passed').length
        : 0
})

const failedTests = computed(() => {
    return Array.isArray(testResults.value)
        ? testResults.value.filter(r => r && r.status === 'failed').length
        : 0
})

const warningTests = computed(() => {
    return Array.isArray(testResults.value)
        ? testResults.value.filter(r => r && r.status === 'warning').length
        : 0
})

const fetchData = async () => {
    isLoading.value = true
    error.value = null

    try {
        const response = await axios.get('/api/results', {
            params: {
                page: pagination.value.currentPage,
                limit: pagination.value.perPage,
                site_id: filters.value.site,
                device_id: filters.value.device,
                feature_id: filters.value.feature,
                status: filters.value.status
            }
        })

        // Handle test results data
        if (response.data && response.data.data) {
            testResults.value = response.data.data
            // Update pagination metadata if available
            if (response.data.meta) {
                pagination.value = {
                    currentPage: response.data.meta.current_page || 1,
                    perPage: response.data.meta.per_page || 10,
                    total: response.data.meta.total || 0,
                    totalPages: response.data.meta.total_pages || 0,
                    hasNext: response.data.meta.has_next || false,
                    hasPrev: response.data.meta.has_prev || false
                }
            }
        } else {
            // Handle legacy response format (array of results)
            if (Array.isArray(response.data)) {
                testResults.value = response.data
                pagination.value = {
                    currentPage: 1,
                    perPage: response.data.length,
                    total: response.data.length,
                    totalPages: 1,
                    hasNext: false,
                    hasPrev: false
                }
            } else {
                testResults.value = []
                pagination.value = {
                    currentPage: 1,
                    perPage: 10,
                    total: 0,
                    totalPages: 0,
                    hasNext: false,
                    hasPrev: false
                }
            }
        }

        lastUpdated.value = new Date().toLocaleString()
    } catch (err) {
        console.error('Error fetching data:', err)
        error.value = 'Failed to load test results. Please try again later.'
        testResults.value = []
        // Reset pagination on error
        pagination.value = {
            currentPage: 1,
            perPage: 10,
            total: 0,
            totalPages: 0,
            hasNext: false,
            hasPrev: false
        }
    } finally {
        isLoading.value = false
    }
}

const fetchDropdownData = async () => {
    isLoading.value = true
    error.value = null

    try {
        const [sitesRes, devicesRes, featuresRes, resultsRes] = await Promise.all([
            axios.get('/api/sites'),
            axios.get('/api/devices'),
            axios.get('/api/features')
        ])

        // Ensure we have arrays even if the API returns null or undefined
        sites.value = Array.isArray(sitesRes.data) ? sitesRes.data : []
        devices.value = Array.isArray(devicesRes.data) ? devicesRes.data : []
        features.value = Array.isArray(featuresRes.data) ? featuresRes.data : []

        lastUpdated.value = new Date().toLocaleString()
    } catch (err) {
        console.error('Error fetching data:', err)
        error.value = 'Failed to load test results. Please try again later.'
        // Initialize with empty arrays if the API calls fail
        sites.value = []
        devices.value = []
        features.value = []
    } finally {
        isLoading.value = false
    }
}

const createNewTest = async () => {
    try {
        const response = await axios.post('/api/results', newTest.value)
            .then(response => {
                // Reset the form
                newTest.value = {
                    site_id: null,
                    device_id: null,
                    feature_id: null
                }
                // Close the modal
                showNewTestModal.value = false
                // Refresh the data
                
                setTimeout(() => {
                    refreshResults()
                }, 1000)
            })
            .catch(err => {
                console.error('Error creating test:', err)
                error.value = 'Failed to create new test. Please try again.'
            })
    } catch (err) {
        console.error('Error creating test:', err)
        error.value = 'Failed to create new test. Please try again.'
    }
}

const refreshResults = () => {
    fetchData()
}

const toggleResultDetails = (resultId) => {
    if (expandedResults.value.has(resultId)) {
        expandedResults.value.delete(resultId)
    } else {
        expandedResults.value.add(resultId)
    }
}

const changePage = (page) => {
    pagination.value.currentPage = page
    fetchData()
}

const applyFilters = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when filters are applied
    fetchData();
};

const handleSearch = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when search is applied
    fetchData();
};

const handleStatusFilter = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when status filter is applied
    fetchData();
};

const handleFeatureFilter = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when feature filter is applied
    fetchData();
};

const handleSiteFilter = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when site filter is applied
    fetchData();
};

const handleBrowserFilter = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when browser filter is applied
    fetchData();
};

const handleDeviceFilter = () => {
    pagination.value.currentPage = 1; // Reset to page 1 when device filter is applied
    fetchData();
};

onMounted(() => {
    fetchData()
    fetchDropdownData()
})
</script>

<template>
    <div class="min-h-screen bg-gray-100">
        <!-- Navigation -->
        <nav class="bg-white shadow-sm">
            <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="flex justify-between h-16">
                    <div class="flex">
                        <div class="flex-shrink-0 flex items-center">
                            <h1 class="text-xl font-bold text-gray-900">QA Automation Dashboard</h1>
                        </div>
                    </div>
                    <div class="flex items-center space-x-4">
                        <button @click="showNewTestModal = true"
                            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                            <svg class="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M12 4v16m8-8H4" />
                            </svg>
                            New Test
                        </button>
                        <!-- <span class="text-sm text-gray-500">Last Updated: {{ lastUpdated }}</span> -->
                    </div>
                </div>
            </div>
        </nav>

        <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
            <!-- Loading State -->
            <div v-if="isLoading" class="flex justify-center items-center h-64">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
            </div>

            <!-- Error State -->
            <div v-else-if="error" class="bg-red-50 border-l-4 border-red-400 p-4 mb-6">
                <div class="flex">
                    <div class="flex-shrink-0">
                        <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fill-rule="evenodd"
                                d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                                clip-rule="evenodd" />
                        </svg>
                    </div>
                    <div class="ml-3">
                        <p class="text-sm text-red-700">{{ error }}</p>
                    </div>
                </div>
            </div>

            <!-- Content -->
            <template v-else>
                <!-- Summary Cards -->
                <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-6">
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="px-4 py-5 sm:p-6">
                            <dt class="text-sm font-medium text-gray-500 truncate">Total Tests</dt>
                            <dd class="mt-1 text-3xl font-semibold text-gray-900">{{ totalTests }}</dd>
                        </div>
                    </div>
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="px-4 py-5 sm:p-6">
                            <dt class="text-sm font-medium text-gray-500 truncate">Passed Tests</dt>
                            <dd class="mt-1 text-3xl font-semibold text-pass">{{ passedTests }}</dd>
                        </div>
                    </div>
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="px-4 py-5 sm:p-6">
                            <dt class="text-sm font-medium text-gray-500 truncate">Failed Tests</dt>
                            <dd class="mt-1 text-3xl font-semibold text-fail">{{ failedTests }}</dd>
                        </div>
                    </div>
                    <div class="bg-white overflow-hidden shadow rounded-lg">
                        <div class="px-4 py-5 sm:p-6">
                            <dt class="text-sm font-medium text-gray-500 truncate">Warning Tests</dt>
                            <dd class="mt-1 text-3xl font-semibold text-warning">{{ warningTests }}</dd>
                        </div>
                    </div>
                </div>

                <!-- Filters -->
                <div class="bg-white shadow px-4 py-5 sm:rounded-lg sm:p-6 mb-6">
                    <div class="grid grid-cols-1 gap-4 sm:grid-cols-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 text-left">Site</label>
                            <select v-model="filters.site"
                                class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;" @change="handleSiteFilter">
                                <option value="">All Sites</option>
                                <option v-for="site in sites" :key="`site-${site.id}`" :value="site.id">{{ site.name }}</option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 text-left">Device</label>
                            <select v-model="filters.device"
                                class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;" @change="handleDeviceFilter">
                                <option value="">All Devices</option>
                                <option v-for="device in devices" :key="`device-${device.id}`" :value="device.id">{{ device.name }}</option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 text-left">Feature</label>
                            <select v-model="filters.feature"
                                class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;" @change="handleFeatureFilter">
                                <option value="">All Features</option>
                                <option v-for="feature in features" :key="`feature-${feature.id}`" :value="feature.id">{{ feature.name }}
                                </option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 text-left">Status</label>
                            <select v-model="filters.status"
                                class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;" @change="handleStatusFilter">
                                <option value="">All Status</option>
                                <option value="passed">Passed</option>
                                <option value="warning">Warning</option>
                                <option value="failed">Failed</option>
                            </select>
                        </div>
                    </div>
                </div>

                <!-- Test Results Table -->
                <div class="bg-white shadow overflow-hidden sm:rounded-md">
                    <div class="px-4 py-5 sm:px-6 flex justify-between items-center">
                        <h3 class="text-lg leading-6 font-medium text-gray-900">Test Results</h3>
                        <button @click="refreshResults"
                            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
                            Refresh Results
                        </button>
                    </div>
                    <!-- Table Header -->
                    <div class="bg-gray-50 border-b border-gray-200">
                        <div class="px-4 py-3 sm:px-6">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center space-x-4">
                                    <div class="w-24">
                                        <span class="text-sm font-medium text-gray-500">Status</span>
                                    </div>
                                    <div class="w-64">
                                        <span class="text-sm font-medium text-gray-500">Feature</span>
                                    </div>
                                    <div class="w-32">
                                        <span class="text-sm font-medium text-gray-500">Duration</span>
                                    </div>
                                    <div class="w-48">
                                        <span class="text-sm font-medium text-gray-500">Error Log</span>
                                    </div>
                                    <div class="w-48">
                                        <span class="text-sm font-medium text-gray-500">Created At</span>
                                    </div>
                                </div>
                                <div class="w-32">
                                    <span class="text-sm font-medium text-gray-500">Action</span>
                                </div>
                            </div>
                        </div>
                    </div>
                    <ul role="list" class="divide-y divide-gray-200">
                        <li v-for="result in testResults" :key="result?.id || Math.random()">
                            <div class="px-4 py-4 sm:px-6">
                                <div class="flex items-center justify-between">
                                    <div class="flex items-center space-x-4">
                                        <div class="w-24">
                                            <span :class="{
                                                'bg-pass/10 text-pass': result?.status === 'passed',
                                                'bg-warning/10 text-warning': result?.status === 'warning' || result?.status === 'processing',
                                                'bg-fail/10 text-fail': result?.status === 'failed'
                                            }" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full items-center justify-center" style="width: 90px;">
                                                {{ result?.status || 'Unknown' }}
                                            </span>
                                        </div>
                                        <div class="w-64">
                                            <div class="text-sm font-medium text-gray-900">{{ result?.feature?.name || 'N/A' }}</div>
                                            <div class="text-sm text-gray-500">
                                                {{ result?.site?.name || 'N/A' }} - 
                                                {{ result?.browser || 'N/A' }} - 
                                                {{ result?.device?.name || 'N/A' }}
                                            </div>
                                        </div>
                                        <div class="w-32">
                                            <p class="text-sm text-gray-500">
                                                {{ result?.duration ? result?.duration.toFixed(2) + 's' : 'N/A' }}
                                            </p>
                                        </div>
                                        <div class="w-48">
                                            <p class="text-sm text-gray-500">
                                                {{ result?.error_log ? result?.error_log : 'N/A' }}
                                            </p>
                                        </div>
                                        <div class="w-48">
                                            <p class="text-sm text-gray-500">
                                                {{ result?.created_at ? new Date(result.created_at).toLocaleString() : 'N/A' }}
                                            </p>
                                        </div>
                                    </div>
                                    <div class="w-32">
                                        <button 
                                            @click="toggleResultDetails(result.id)"
                                            class="inline-flex items-center px-3 py-1 border border-transparent text-sm leading-4 font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                        >
                                            {{ expandedResults.has(result.id) ? 'Hide Details' : 'Show Details' }}
                                        </button>
                                    </div>
                                </div>
                                <div v-if="expandedResults.has(result.id)" class="mt-4 text-left">
                                    <hr class="mb-4">
                                    <!-- <div v-if="result?.error_log" class="text-sm text-gray-500 mb-4">
                                        {{ result.error_log }}
                                    </div> -->
                                    <template v-if="result?.details?.length > 0">
                                        <div v-for="detail in result.details" :key="detail.id" class="mb-6">
                                            <div class="text-sm font-medium text-gray-700 mb-2">{{ detail.description || 'N/A' }}</div>
                                            <img :src="`${apiUrl}/${detail.screenshot}`" class="max-w-lg rounded-lg shadow-md" />
                                        </div>
                                    </template>
                                    <div v-else class="text-sm text-gray-500">
                                        No details available
                                    </div>
                                </div>
                            </div>
                        </li>
                    </ul>
                    
                    <!-- Pagination -->
                    <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
                        <div class="flex-1 flex justify-between sm:hidden">
                            <button 
                                @click="changePage(pagination.currentPage - 1)"
                                :disabled="!pagination.hasPrev"
                                class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Previous
                            </button>
                            <button 
                                @click="changePage(pagination.currentPage + 1)"
                                :disabled="!pagination.hasNext"
                                class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Next
                            </button>
                        </div>
                        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                            <div>
                                <p class="text-sm text-gray-700">
                                    Showing
                                    <span class="font-medium">{{ ((pagination.currentPage - 1) * pagination.perPage) + 1 }}</span>
                                    to
                                    <span class="font-medium">{{ Math.min(pagination.currentPage * pagination.perPage, pagination.total) }}</span>
                                    of
                                    <span class="font-medium">{{ pagination.total }}</span>
                                    results
                                </p>
                            </div>
                            <div>
                                <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                                    <button 
                                        @click="changePage(pagination.currentPage - 1)"
                                        :disabled="!pagination.hasPrev"
                                        class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        <span class="sr-only">Previous</span>
                                        <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                    <button 
                                        v-for="page in pagination.totalPages" 
                                        :key="page"
                                        @click="changePage(page)"
                                        :class="[
                                            page === pagination.currentPage
                                                ? 'z-10 bg-indigo-50 border-indigo-500 text-indigo-600'
                                                : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                                            'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
                                        ]"
                                    >
                                        {{ page }}
                                    </button>
                                    <button 
                                        @click="changePage(pagination.currentPage + 1)"
                                        :disabled="!pagination.hasNext"
                                        class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        <span class="sr-only">Next</span>
                                        <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                </nav>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </main>

        <!-- New Test Modal -->
        <div v-if="showNewTestModal" class="fixed z-10 inset-0 overflow-y-auto" aria-labelledby="modal-title"
            role="dialog" aria-modal="true">
            <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
                <!-- Background overlay -->
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true"
                    @click="showNewTestModal = false"></div>

                <!-- Modal panel -->
                <div
                    class="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
                    <div class="sm:flex sm:items-start">
                        <div class="mt-3 text-center sm:mt-0 sm:text-left w-full">
                            <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                                Create New Test
                            </h3>
                            <div class="mt-4">
                                <form @submit.prevent="createNewTest" class="space-y-4">
                                    <div>
                                        <label for="site" class="block text-sm font-medium text-gray-700">Site</label>
                                        <select id="site" v-model="newTest.site_id" required
                                            class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;">
                                            <option :value="null">Select a site</option>
                                            <option v-for="site in sites" :key="`new-site-${site.id}`" :value="Number(site.id)">{{ site.name }}</option>
                                        </select>
                                    </div>

                                    <div>
                                        <label for="device"
                                            class="block text-sm font-medium text-gray-700">Device</label>
                                        <select id="device" v-model="newTest.device_id" required
                                            class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;">
                                            <option :value="null">Select a device</option>
                                            <option v-for="device in devices" :key="`new-device-${device.id}`" :value="Number(device.id)">{{ device.name }}</option>
                                        </select>
                                    </div>

                                    <div>
                                        <label for="feature"
                                            class="block text-sm font-medium text-gray-700">Feature</label>
                                        <select id="feature" v-model="newTest.feature_id" required
                                            class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md" style="color: #000;">
                                            <option :value="null">Select a feature</option>
                                            <option v-for="feature in features" :key="`new-feature-${feature.id}`" :value="Number(feature.id)">{{ feature.name }}</option>
                                        </select>
                                    </div>

                                    <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                        <button type="submit"
                                            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
                                            Create Test
                                        </button>
                                        <button type="button"
                                            class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:w-auto sm:text-sm"
                                            @click="showNewTestModal = false">
                                            Cancel
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.logo {
    height: 6em;
    padding: 1.5em;
    will-change: filter;
    transition: filter 300ms;
}

.logo:hover {
    filter: drop-shadow(0 0 2em #646cffaa);
}

.logo.vue:hover {
    filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
