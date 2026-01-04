<template>
  <div v-if="show" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
      <div class="flex items-center mb-4">
        <svg class="w-6 h-6 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <h3 class="text-lg font-semibold text-gray-900">Error</h3>
      </div>
      <p class="text-gray-700 mb-6">{{ message }}</p>
      <div class="flex justify-end">
        <button
          @click="$emit('close')"
          class="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ErrorModal',
  props: {
    show: {
      type: Boolean,
      required: true
    },
    message: {
      type: String,
      default: 'An error occurred'
    }
  },
  emits: ['close']
}
</script>
