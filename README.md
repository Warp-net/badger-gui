<iframe src="https://snapcraft.io/badger-gui/embedded?button=black" frameborder="0" width="100%" height="450px" style="border: 1px solid #CCC; border-radius: 2px;"></iframe>
# Badger DB GUI
![screenshot.png](screenshot.png)
A graphical user interface for managing [Badger](https://github.com/dgraph-io/badger) key-value database

## Features

- **Database Connection**: Open Badger databases with support for:
  - Custom database path selection
  - Encryption key support
  - Compression options (Snappy, ZSTD, None)
  - Custom key delimiter for nested key parsing

- **Data Management**: Full CRUD operations
  - List all keys with pagination
  - View key-value pairs
  - Add new entries
  - Update existing values
  - Delete entries
  - Search by key prefix

- **User Interface**:
  - Color-coded nested key visualization
  - Split-panel design for browsing and editing
  - Dark theme interface
  - Error handling with modal dialogs

## Download
 
 - [Release binaries](https://github.com/Warp-net/badger-gui/releases)
 - [Snapstore](https://snapcraft.io/badger-gui)

## Building

### Prerequisites

- Go 1.18 or higher
- Node.js 14 or higher
- Wails CLI (optional for development)

### Build Application

```bash
# For development (requires Wails CLI)
wails dev -m -nosyncgomod -tags webkit2_41

# For production
wails build -m -nosyncgomod -tags webkit2_41
```

## Usage

1. **Open Database**:
   - Launch the application
   - Click "Browse" to select your Badger database folder
   - (Optional) Enter decryption key if database is encrypted
   - Select compression type used by the database
   - Enter a delimiter character for nested keys (e.g., `/` for keys like `items/book/123`)
   - Click "Open Database"

2. **Manage Data**:
   - View all keys in the left panel
   - Click on a key to view its value
   - Use "Add New Entry" to create new key-value pairs
   - Click "Edit" to modify values
   - Click "Delete" to remove entries
   - Use the search box to filter keys by prefix
   - Click "Load More" for paginated results

3. **Key Format**:
   - Keys are displayed as color-coded blocks separated by your chosen delimiter
   - Example: `items/book/123` displays as [items][book][123] with different colors
   - Each level of the hierarchy gets a distinct color for easy visualization

## Architecture

- **Backend**: Go with Badger DB integration
- **Frontend**: Vue 3 with Vue Router
- **Communication**: Wails v2 binding for Go-JavaScript interop
- **Styling**: Tailwind CSS (inline utility classes)

## API Methods

The application exposes the following backend methods:

- `OpenDirectoryDialog()`: Opens a directory picker dialog
- `Call(AppMessage)`: Main RPC endpoint for database operations
  - `open`: Open database connection
  - `list`: List keys with optional pagination
  - `search`: Search keys with prefix filter and pagination
  - `get`: Retrieve value for a specific key
  - `set`: Create or update a key-value pair
  - `delete`: Remove a key-value pair

## Development

### Adding New Features

1. Update backend methods in `app.go`
2. Update frontend components in `frontend/src/`
3. Run `wails build -m -nosyncgomod -tags webkit2_41`

## License

See [LICENSE](LICENSE) file for details.
