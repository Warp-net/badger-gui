<template>
  <div class="min-h-screen bg-gray-900 text-white">
    <!-- Header -->
    <div class="bg-gray-800 border-b border-gray-700 px-6 py-4">
      <div class="flex keys-center justify-between">
        <h1 class="text-2xl font-bold">Badger DB Data Manager</h1>
      </div>
    </div>

    <div class="flex h-[calc(100vh-73px)]">
      <!-- Left Sidebar -->
      <div class="w-96 bg-gray-800 border-r border-gray-700 flex flex-col">
        <div class="p-4 border-b border-gray-700">
          <div class="flex gap-2 mb-4">
            <input
                v-model="searchPrefix"
                type="text"
                placeholder="Search by prefix..."
                class="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-400 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
                @keyup.enter="searchKeys"
            />
            <button
                @click="searchKeys"
                class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            >
              Search
            </button>
            <button v-if="isInSearch"
                @click="loadKeys(false)"
                class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            >
              ⏎
            </button>
          </div>

          <button
              @click="showAddModal = true"
              class="w-full px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-green-500"
          >
            + Add New Entry
          </button>
        </div>

        <!-- Key List -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="loading" class="p-4 text-center text-gray-400">
            Loading...
          </div>
          <div v-else-if="keys.length === 0" class="p-4 text-center text-gray-400">
            No entries found
          </div>
          <div v-else>
            <div
                v-for="key in keys"
                :key="key"
                @click="selectKey(key)"
                class="p-3 border-b border-gray-700 cursor-pointer hover:bg-gray-700"
                :class="{ 'bg-gray-700': selectedKey === key }"
            >
              <div class="flex flex-wrap gap-1">
                <span
                    v-for="(part, index) in parseKey(key)"
                    :key="index"
                    class="inline-block px-2 py-1 text-xs rounded font-medium"
                    :style="{ backgroundColor: getColor(index), color: '#fff' }"
                >
                  {{ part }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="cursor && cursor !== 'end'" class="p-4 border-t border-gray-700">
          <button
              @click="loadMore"
              class="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
          >
            Load More
          </button>
        </div>
      </div>

      <!-- Right Panel -->
      <div class="flex-1 flex flex-col">
        <div v-if="!selectedKey" class="flex-1 flex keys-center justify-center text-gray-500">
          <div class="text-center">
            <p class="text-xl">Select a key to view its value</p>
          </div>
        </div>

        <div v-else class="flex-1 flex flex-col p-6">
          <div class="mb-4">
            <h2 class="text-lg font-semibold mb-2">Key</h2>
            <div class="flex flex-wrap gap-1 p-3 bg-gray-800 border border-gray-700 rounded">
                {{ selectedKey }}
            </div>
          </div>

          <div class="mb-4 flex-1 flex flex-col">
            <div class="flex keys-center justify-between mb-2">
              <h2 class="text-lg font-semibold">Value</h2>
              <div class="flex gap-2">
                <button
                    @click="editMode = !editMode"
                    class="px-3 py-1 bg-yellow-600 text-white rounded hover:bg-yellow-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-yellow-500"
                >
                  {{ editMode ? 'Cancel' : 'Edit' }}
                </button>
                <button
                    v-if="editMode"
                    @click="updateValue"
                    class="px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-green-500"
                >
                  Save
                </button>
                <button
                    @click="confirmDelete"
                    class="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-red-500"
                >
                  Delete
                </button>
              </div>
            </div>

            <textarea
                v-model="currentValue"
                :readonly="!editMode"
                class="flex-1 p-3 bg-gray-800 border border-gray-700 rounded font-mono text-sm focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
                :class="{ 'bg-gray-700': editMode }"
            ></textarea>
          </div>
        </div>
      </div>
    </div>

    <!-- Add/Set Modal -->
    <div v-if="showAddModal" class="fixed inset-0 bg-black flex keys-center justify-center z-50">
      <div class="bg-gray-800 rounded-lg p-6 max-w-2xl w-full mx-4 border border-gray-700">
        <h3 class="text-xl font-semibold mb-4">Add New Entry</h3>
        <form @submit.prevent="addEntry">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-200 mb-2">Key</label>
            <input
                v-model="newEntry.key"
                type="text"
                required
                placeholder="e.g., keys/book/123"
                class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-400 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            />
          </div>

          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-200 mb-2">Value</label>
            <textarea
                v-model="newEntry.value"
                required
                rows="6"
                placeholder="Enter value..."
                class="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-400 font-mono text-sm focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-blue-500"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2">
            <button
                type="button"
                @click="showAddModal = false"
                class="px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-gray-500"
            >
              Cancel
            </button>
            <button
                type="submit"
                class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-green-500"
            >
              Add
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Error Modal -->
    <ErrorModal :show="showError" :message="errorMessage" @close="showError = false" />

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black flex keys-center justify-center z-50">
      <div class="bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4 border border-gray-700">
        <h3 class="text-xl font-semibold mb-4">Confirm Delete</h3>
        <p class="mb-6 text-gray-300">Are you sure you want to delete this key?</p>
        <div class="flex justify-end gap-2">
          <button
              @click="showDeleteConfirm = false"
              class="px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-600 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-gray-500"
          >
            Cancel
          </button>
          <button
              @click="deleteKey"
              class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-red-500"
          >
            Delete
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Call } from '../wailsjs/go/main/App'
import ErrorModal from '../components/ErrorModal.vue'

export default {
  name: 'DataManager',
  components: {
    ErrorModal
  },
  setup() {
    const router = useRouter()
    const delimiter = ref(sessionStorage.getItem('delimiter') || '/')
    const keys = ref([])
    const selectedKey = ref(null)
    const currentValue = ref('')
    const editMode = ref(false)
    const loading = ref(false)
    const searchPrefix = ref('')
    const searchOffset = ref(0)
    const isInSearch = ref(false)
    const cursor = ref(null)
    const showError = ref(false)
    const errorMessage = ref('')
    const showAddModal = ref(false)
    const showDeleteConfirm = ref(false)
    const newEntry = ref({ key: '', value: '' })

    const colors = [
      '#3B82F6', // blue
      '#10B981', // green
      '#F59E0B', // amber
      '#EF4444', // red
      '#8B5CF6', // purple
      '#EC4899', // pink
      '#14B8A6', // teal
      '#F97316', // orange
    ]

    const getColor = (index) => {
      return colors[index % colors.length]
    }

    const parseKey = (key) => {
      key = key.startsWith(delimiter.value)
          ? key.slice(delimiter.value.length)
          : key;1
      return key.split(delimiter.value)
    }

    const parseResponse = (response) => {
        return JSON.parse(response.body)
    }

    const isOkResponse = (responseText) => {
      return responseText === 'ok'
    }

    const loadKeys = async (loadMore = false) => {
      loading.value = true
      searchOffset.value = 0
      isInSearch.value = false
      searchPrefix.value = ''

      try {
        const message = {
          type: 'list',
          body: JSON.stringify({
            limit: 20,
            cursor: loadMore ? cursor.value : null
          })
        }

        console.log('[Frontend] Sending list request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received list response:', response)
        
        if (response.type === 'list') {
          const data = parseResponse(response)
          console.log('[Frontend] Parsed list data:', data)
          
          if (loadMore) {
            keys.value = [...keys.value, ...(data.keys || [])]
          } else {
            keys.value = data.keys || []
          }
          
          cursor.value = data.cursor || null
        } else {
          const errorText = parseResponse(response)
          console.error('[Frontend] List operation failed:', errorText)
          errorMessage.value = errorText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error loading keys:', error)
        errorMessage.value = 'Failed to load keys: ' + (error.message || error)
        showError.value = true
      } finally {
        loading.value = false
      }
    }

    const loadMore = () => {
      loadKeys(true)
    }

    const searchKeys = async (searchMore = false) => {
      loading.value = true

      if (!isInSearch.value) {
        keys.value = []
      }
      isInSearch.value = true

      let offset = 0
      if (searchMore) {
        offset = searchOffset.value
      }
      try {
        const message = {
          type: 'search',
          body: JSON.stringify({
            prefix: searchPrefix.value,
            limit: 20,
            offset: offset
          })
        }

        console.log('[Frontend] Sending search request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received search response:', response)

        if (response.type === 'search') {
          const data = parseResponse(response)
          console.log('[Frontend] Parsed search data:', data)

          if (searchMore) {
            keys.value = [...keys.value, ...(data.keys || [])]
          } else {
            keys.value = data.keys || []
          }

          searchOffset.value += data.offset
        } else {
          const errorText = parseResponse(response)
          console.error('[Frontend] Search operation failed:', errorText)
          errorMessage.value = errorText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error searching keys:', error)
        errorMessage.value = 'Failed to search keys: ' + (error.message || error)
        showError.value = true
      } finally {
        loading.value = false
      }
    }

    // TODO
    const searchMore = () => {
      searchKeys(isInSearch)
    }

    const selectKey = async (key) => {
      selectedKey.value = key
      editMode.value = false
      
      try {
        const message = {
          type: 'get',
          body: JSON.stringify({
            key: key
          })
        }

        console.log('[Frontend] Sending get request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received get response:', response)
        
        if (response.type === 'get') {
          const data = parseResponse(response)
          console.log('[Frontend] Parsed get data:', data)
          currentValue.value = data.value
        } else {
          const errorText = parseResponse(response)
          console.error('[Frontend] Get operation failed:', errorText)
          errorMessage.value = errorText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error getting value:', error)
        errorMessage.value = 'Failed to get value: ' + (error.message || error)
        showError.value = true
      }
    }

    const updateValue = async () => {
      try {
        const message = {
          type: 'set',
          body: JSON.stringify({
            key: selectedKey.value,
            value: currentValue.value
          })
        }

        console.log('[Frontend] Sending set request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received set response:', response)
        
        const responseText = response.body
        console.log('[Frontend] Set response text:', responseText)
        if (isOkResponse(responseText)) {
          editMode.value = false
          await loadKeys()
        } else {
          console.error('[Frontend] Set operation failed:', responseText)
          errorMessage.value = responseText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error updating value:', error)
        errorMessage.value = 'Failed to update value: ' + (error.message || error)
        showError.value = true
      }
    }

    const confirmDelete = () => {
      showDeleteConfirm.value = true
    }

    const deleteKey = async () => {
      try {
        const message = {
          type: 'delete',
          body: JSON.stringify({
            key: selectedKey.value
          })
        }

        console.log('[Frontend] Sending delete request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received delete response:', response)
        
        const responseText = response.body
        console.log('[Frontend] Delete response text:', responseText)
        if (isOkResponse(responseText)) {
          showDeleteConfirm.value = false
          selectedKey.value = null
          currentValue.value = ''
          // Reload the list
          await loadKeys()
        } else {
          console.error('[Frontend] Delete operation failed:', responseText)
          errorMessage.value = responseText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error deleting key:', error)
        errorMessage.value = 'Failed to delete key: ' + (error.message || error)
        showError.value = true
      }
    }

    const addEntry = async () => {
      try {
        const message = {
          type: 'set',
          body: JSON.stringify({
            key: newEntry.value.key,
            value: newEntry.value.value
          })
        }

        console.log('[Frontend] Sending add entry request:', message)
        const response = await Call(message)
        console.log('[Frontend] Received add entry response:', response)
        
        const responseText = response.body
        console.log('[Frontend] Add entry response text:', responseText)
        if (isOkResponse(responseText)) {
          showAddModal.value = false
          newEntry.value = { key: '', value: '' }
          // Reload the list
          await loadKeys()
        } else {
          console.error('[Frontend] Add entry operation failed:', responseText)
          errorMessage.value = responseText
          showError.value = true
        }
      } catch (error) {
        console.error('[Frontend] Error adding entry:', error)
        errorMessage.value = 'Failed to add entry: ' + (error.message || error)
        showError.value = true
      }
    }

    onMounted(() => {
      loadKeys()
    })

    return {
      keys,
      selectedKey,
      currentValue,
      editMode,
      loading,
      searchPrefix,
      searchOffset,
      isInSearch,
      cursor,
      showError,
      errorMessage,
      showAddModal,
      showDeleteConfirm,
      newEntry,
      parseKey,
      getColor,
      loadKeys,
      searchKeys,
      loadMore,
      searchMore,
      selectKey,
      updateValue,
      confirmDelete,
      deleteKey,
      addEntry
    }
  }
}
</script>
