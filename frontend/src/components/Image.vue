<script setup>
import { ref } from 'vue'

defineProps({
    imageUrl: String,
})

const isFullWidth = ref(false)
</script>

<template>
    <div class="qa-image">
        <img 
            @click="isFullWidth = !isFullWidth" 
            :src="imageUrl" 
            class="rounded-lg shadow-md cursor-pointer"
            :class="{ 
                'max-w-full': isFullWidth,
                'max-w-lg': !isFullWidth
            }" 
        />
    </div>
</template>

<style scoped></style>
