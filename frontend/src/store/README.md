# Redux Store Structure

This directory contains the Redux store configuration and slices for state management.

## Store Structure

```
store/
├── store.js              # Redux store configuration
└── slices/
    ├── authSlice.js      # Authentication state
    ├── bugsSlice.js      # Bugs data and filtering
    └── releaseNotesSlice.js  # Release notes generation
```

## Slices

### 1. authSlice.js
Manages user authentication state.

**State:**
- `user`: User object with email and role
- `email`: User email
- `role`: User role ('developer' or 'manager')
- `isAuthenticated`: Boolean flag

**Actions:**
- `login({ email, role })`: Log in user
- `logout()`: Log out user
- `setUser(user)`: Set user data

### 2. bugsSlice.js
Manages bugs data, filtering, and selection.

**State:**
- `bugs`: Array of all bugs
- `filteredBugs`: Array of filtered bugs
- `selectedBug`: Currently selected bug
- `loading`: Loading state
- `error`: Error message
- `filters`: Object with search, status, and release filters

**Actions:**
- `setSelectedBug(bug)`: Select a bug
- `clearSelectedBug()`: Clear selected bug
- `setSearchFilter(value)`: Set search filter
- `setStatusFilter(value)`: Set status filter
- `setReleaseFilter(value)`: Set release filter
- `applyFilters()`: Apply all filters to bugs

**Async Thunks:**
- `fetchBugs()`: Fetch all bugs (TODO: Replace with API call)
- `updateBugStatus({ bugId, status })`: Update bug status (TODO: Replace with API call)

### 3. releaseNotesSlice.js
Manages release notes generation and approval workflow.

**State:**
- `releaseNotes`: Object mapping bugId to release note data
- `currentNote`: Currently active release note
- `loading`: Loading state
- `error`: Error message

**Actions:**
- `setCurrentNote(note)`: Set current release note
- `updateNoteContent({ bugId, content })`: Update note content
- `clearCurrentNote()`: Clear current note

**Async Thunks:**
- `generateReleaseNote({ bugId })`: Generate release note for a bug
- `regenerateReleaseNote({ bugId, feedback })`: Regenerate with feedback
- `sendToApproval({ bugId, content })`: Send to approval (developer)
- `approveReleaseNote({ bugId })`: Approve release note (manager)

## Usage Examples

### Using in Components

```javascript
import { useDispatch, useSelector } from 'react-redux';
import { login } from '../store/slices/authSlice';
import { fetchBugs, setSearchFilter } from '../store/slices/bugsSlice';

const MyComponent = () => {
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { filteredBugs, loading } = useSelector((state) => state.bugs);

  // Dispatch actions
  dispatch(login({ email: 'user@example.com', role: 'developer' }));
  dispatch(fetchBugs());
  dispatch(setSearchFilter('bug title'));

  return (
    // Component JSX
  );
};
```

## TODO: API Integration

All async thunks currently use mock data. Replace with actual API calls:

1. **authSlice**: Add login/logout API calls
2. **bugsSlice**: 
   - `fetchBugs()`: GET /api/bugs
   - `updateBugStatus()`: PATCH /api/bugs/:id/status
3. **releaseNotesSlice**:
   - `generateReleaseNote()`: POST /api/release-notes/generate
   - `regenerateReleaseNote()`: POST /api/release-notes/regenerate
   - `sendToApproval()`: POST /api/release-notes/approve
   - `approveReleaseNote()`: POST /api/release-notes/approve

## Persistence

Currently using localStorage for basic persistence:
- User email and role stored on login
- TODO: Implement proper token-based authentication
- TODO: Add Redux persist middleware for state persistence

