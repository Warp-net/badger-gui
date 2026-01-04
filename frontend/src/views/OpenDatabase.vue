<template>
  <div class="min-h-screen bg-white flex items-center justify-center p-4">
    <div class="bg-white rounded-lg p-8 max-w-2xl w-full border border-gray-200">
      <h1 class="text-3xl font-bold text-gray-900 mb-8 text-center">Badger DB Manager</h1>

      <form @submit.prevent="openDatabase" class="space-y-6">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Database Folder Path
          </label>
          <div class="flex gap-2">
            <input
                v-model="form.path"
                type="text"
                placeholder="/path/to/badger/db"
                class="flex-1 px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            />
            <button
                type="button"
                @click="selectFolder"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            >
              Browse
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Decryption Key (Optional)
          </label>
          <input
              v-model="form.decryptionKey"
              type="password"
              placeholder="Enter decryption key if database is encrypted"
              class="w-full px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Compression
          </label>
          <select
              v-model="form.compression"
              class="w-full px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
          >
            <option value="none">None</option>
            <option value="snappy">Snappy</option>
            <option value="zstd">ZSTD</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            Key Delimiter (for nested key parsing)
          </label>
          <input
              v-model="form.delimiter"
              type="text"
              pattern="[^a-zA-Z0-9]"
              maxlength="1"
              placeholder="/ or : or -"
              title="Must be a special character (non-alphanumeric)"
              class="w-full px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
          />
          <p class="mt-1 text-sm text-gray-500">
            Use a special character like / : - _ to separate nested keys (e.g., items/book/123)
          </p>
        </div>

        <div class="pt-4">
          <button
              type="submit"
              :disabled="loading"
              class="w-full px-6 py-3 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-green-500"
          >
            {{ loading ? 'Opening Database...' : 'Open Database' }}
          </button>
        </div>
      </form>
    </div>

    <ErrorModal :show="showError" :message="errorMessage" @close="showError = false" />
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Call, OpenDirectoryDialog } from '../wailsjs/go/main/App'
import ErrorModal from '../components/ErrorModal.vue'

export default {
  name: 'OpenDatabase',
  components: { ErrorModal },
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const showError = ref(false)
    const errorMessage = ref('')

    const form = ref({
      path: '',
      decryptionKey: '',
      compression: 'none',
      delimiter: '/'
    })

    const selectFolder = async () => {
      const selectedPath = await OpenDirectoryDialog()
      if (typeof selectedPath === 'string') {
        form.value.path = selectedPath
      }
    }

    const openDatabase = async () => {
      loading.value = true
      try {
        const message = {
          type: 'open',
          body: JSON.stringify({
            path: form.value.path,
            decryption_key: form.value.decryptionKey,
            compression: form.value.compression,
            delimiter: form.value.delimiter
          })
        }

        const response = await Call(message)

        if (response?.type === 'open') {
          const responseText = response.body
          if (responseText === 'ok') {
            sessionStorage.setItem('delimiter', form.value.delimiter)
            await router.push('/manager')
            return
          }

          errorMessage.value = String(responseText ?? 'Unknown error')
          showError.value = true
          return
        }

        errorMessage.value = 'Unexpected response type'
        showError.value = true
      } catch (error) {
        errorMessage.value = 'Failed to open database: ' + String(error?.message || error)
        showError.value = true
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      loading,
      showError,
      errorMessage,
      selectFolder,
      openDatabase
    }
  }
}
</script>
