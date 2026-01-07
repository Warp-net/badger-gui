/**
 * Data type detection and conversion utilities
 * Converts various data formats to human-readable strings
 */

/**
 * Check if a string is valid hexadecimal
 */
function isHexString(str) {
  if (typeof str !== 'string') return false
  // Remove potential 0x prefix
  const cleanStr = str.startsWith('0x') ? str.slice(2) : str
  return /^[0-9A-Fa-f]+$/.test(cleanStr) && cleanStr.length % 2 === 0
}

/**
 * Check if a string is valid base64
 */
function isBase64String(str) {
  if (typeof str !== 'string') return false
  // Base64 regex pattern
  const base64Regex = /^[A-Za-z0-9+/]+={0,2}$/
  return base64Regex.test(str) && str.length % 4 === 0
}

/**
 * Check if data appears to be binary (contains non-printable characters)
 */
function isBinaryData(str) {
  if (typeof str !== 'string') return false
  // Check for null bytes or other non-printable characters
  for (let i = 0; i < Math.min(str.length, 100); i++) {
    const code = str.charCodeAt(i)
    // Check for control characters except newline, tab, and carriage return
    if (code < 32 && code !== 9 && code !== 10 && code !== 13) {
      return true
    }
    // Check for common binary patterns
    if (code > 127) {
      return true
    }
  }
  return false
}

/**
 * Check if string looks like a media file metadata (e.g., image headers)
 */
function isMediaData(str) {
  if (typeof str !== 'string') return false
  // Common media file signatures
  const mediaSignatures = [
    '\xFF\xD8\xFF', // JPEG
    '\x89PNG', // PNG
    'GIF87a', 'GIF89a', // GIF
    'RIFF', // WAV, AVI
    'ID3', // MP3
    '\x00\x00\x00\x18ftypmp4', // MP4
    '\x00\x00\x00\x1Cftypisom', // MP4
    'fLaC', // FLAC
    'OggS' // OGG
  ]
  
  const start = str.substring(0, 20)
  return mediaSignatures.some(sig => start.includes(sig))
}

/**
 * Convert binary data to hexadecimal string
 */
function binaryToHex(str) {
  let hex = ''
  for (let i = 0; i < str.length; i++) {
    const byte = str.charCodeAt(i).toString(16).padStart(2, '0')
    hex += byte
  }
  return '0x' + hex
}

/**
 * Convert hexadecimal string to regular string
 */
function hexToString(hex) {
  try {
    // Remove 0x prefix if present
    const cleanHex = hex.startsWith('0x') ? hex.slice(2) : hex
    let str = ''
    for (let i = 0; i < cleanHex.length; i += 2) {
      const charCode = parseInt(cleanHex.substr(i, 2), 16)
      // Only add printable characters
      if (charCode >= 32 && charCode <= 126) {
        str += String.fromCharCode(charCode)
      }
    }
    return str || hex // Return original if conversion fails
  } catch (e) {
    return hex
  }
}

/**
 * Convert base64 to string
 */
function base64ToString(base64) {
  try {
    return atob(base64)
  } catch (e) {
    return base64 // Return original if decoding fails
  }
}

/**
 * Extract metadata from media data
 */
function extractMediaMetadata(str) {
  const start = str.substring(0, 20)
  
  if (start.includes('\xFF\xD8\xFF')) {
    return `[Image: JPEG, ${str.length} bytes]`
  } else if (start.includes('\x89PNG')) {
    return `[Image: PNG, ${str.length} bytes]`
  } else if (start.includes('GIF87a') || start.includes('GIF89a')) {
    return `[Image: GIF, ${str.length} bytes]`
  } else if (start.includes('RIFF')) {
    return `[Media: WAV/AVI, ${str.length} bytes]`
  } else if (start.includes('ID3')) {
    return `[Audio: MP3, ${str.length} bytes]`
  } else if (start.includes('ftypmp4') || start.includes('ftypisom')) {
    return `[Video: MP4, ${str.length} bytes]`
  } else if (start.includes('fLaC')) {
    return `[Audio: FLAC, ${str.length} bytes]`
  } else if (start.includes('OggS')) {
    return `[Audio: OGG, ${str.length} bytes]`
  }
  
  return `[Binary Data: ${str.length} bytes]`
}

/**
 * Convert byte number to human-readable string
 */
function bytesToString(bytes) {
  const num = parseInt(bytes)
  if (isNaN(num)) return bytes
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = num
  let unitIndex = 0
  
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

/**
 * Main function to detect data type and convert to appropriate string representation
 * @param {string} value - The value to convert
 * @returns {Object} - Object with displayValue and detectedType
 */
export function detectAndConvertDataType(value) {
  if (!value || typeof value !== 'string') {
    return {
      displayValue: value || '',
      detectedType: 'string',
      originalValue: value
    }
  }

  // Check for media data first (before binary check)
  if (isMediaData(value)) {
    return {
      displayValue: extractMediaMetadata(value),
      detectedType: 'media',
      originalValue: value
    }
  }

  // Check for binary data
  if (isBinaryData(value)) {
    return {
      displayValue: binaryToHex(value),
      detectedType: 'binary',
      originalValue: value
    }
  }

  // Check if it's a hex string that should be converted
  if (isHexString(value) && value.length > 10) {
    const converted = hexToString(value)
    if (converted !== value && converted.length > 0) {
      return {
        displayValue: converted,
        detectedType: 'hex',
        originalValue: value
      }
    }
  }

  // Check for base64
  if (isBase64String(value) && value.length > 20) {
    const converted = base64ToString(value)
    // Only use base64 conversion if it results in printable text
    if (!isBinaryData(converted) && converted !== value) {
      return {
        displayValue: converted,
        detectedType: 'base64',
        originalValue: value
      }
    }
  }

  // Check if it's a number representing bytes
  if (/^\d+$/.test(value) && parseInt(value) > 1024) {
    return {
      displayValue: `${value} (${bytesToString(value)})`,
      detectedType: 'bytes',
      originalValue: value
    }
  }

  // Default: return as-is
  return {
    displayValue: value,
    detectedType: 'string',
    originalValue: value
  }
}

/**
 * Get a badge label for the detected data type
 */
export function getDataTypeBadge(detectedType) {
  const badges = {
    'binary': { label: 'Binary → Hex', color: '#EF4444' },
    'hex': { label: 'Hex → String', color: '#3B82F6' },
    'base64': { label: 'Base64 → String', color: '#10B981' },
    'media': { label: 'Media', color: '#8B5CF6' },
    'bytes': { label: 'Bytes', color: '#F59E0B' },
    'string': { label: 'Text', color: '#6B7280' }
  }
  
  return badges[detectedType] || badges['string']
}
