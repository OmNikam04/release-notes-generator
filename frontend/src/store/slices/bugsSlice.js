import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { dummyBugs } from '../../data/dummyBugs';

// Async thunk for fetching bugs
export const fetchBugs = createAsyncThunk(
  'bugs/fetchBugs',
  async (bugsData = null, { rejectWithValue }) => {
    try {
      // If bugsData is provided (from API), use it
      if (bugsData) {
        console.log('[Redux] Setting bugs from API data:', bugsData);
        return bugsData;
      }

      // Otherwise, use dummy data (fallback)
      console.log('[Redux] Using dummy data as fallback');
      return dummyBugs;
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

// Async thunk for updating bug status
export const updateBugStatus = createAsyncThunk(
  'bugs/updateBugStatus',
  async ({ bugId, status }, { rejectWithValue }) => {
    try {
      // TODO: Replace with actual API call
      // const response = await fetch(`/api/bugs/${bugId}/status`, {
      //   method: 'PATCH',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ status }),
      // });
      // const data = await response.json();
      // return data;
      
      // For now, return mock data
      return { bugId, status };
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

const initialState = {
  bugs: [],
  filteredBugs: [],
  selectedBug: null,
  loading: false,
  error: null,
  filters: {
    search: '',
    status: 'all',
    release: 'all',
  },
};

const bugsSlice = createSlice({
  name: 'bugs',
  initialState,
  reducers: {
    setSelectedBug: (state, action) => {
      state.selectedBug = action.payload;
    },
    clearSelectedBug: (state) => {
      state.selectedBug = null;
    },
    setSearchFilter: (state, action) => {
      state.filters.search = action.payload;
    },
    setStatusFilter: (state, action) => {
      state.filters.status = action.payload;
    },
    setReleaseFilter: (state, action) => {
      state.filters.release = action.payload;
    },
    applyFilters: (state) => {
      let filtered = state.bugs;

      // Apply search filter
      if (state.filters.search) {
        const searchLower = state.filters.search.toLowerCase();
        filtered = filtered.filter(bug =>
          bug.title.toLowerCase().includes(searchLower) ||
          bug.description.toLowerCase().includes(searchLower) ||
          bug.component.toLowerCase().includes(searchLower) ||
          bug.bugsby_id.toLowerCase().includes(searchLower)
        );
      }

      // Apply status filter
      if (state.filters.status !== 'all') {
        filtered = filtered.filter(bug =>
          bug.status.toLowerCase() === state.filters.status.toLowerCase()
        );
      }

      // Apply release filter
      if (state.filters.release !== 'all') {
        filtered = filtered.filter(bug =>
          bug.release === state.filters.release
        );
      }

      state.filteredBugs = filtered;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch bugs
      .addCase(fetchBugs.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchBugs.fulfilled, (state, action) => {
        state.loading = false;
        state.bugs = action.payload;
        state.filteredBugs = action.payload;
      })
      .addCase(fetchBugs.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
      })
      // Update bug status
      .addCase(updateBugStatus.fulfilled, (state, action) => {
        const { bugId, status } = action.payload;
        const bugIndex = state.bugs.findIndex(bug => bug.id === bugId);
        if (bugIndex !== -1) {
          state.bugs[bugIndex].status = status;
        }
      });
  },
});

export const {
  setSelectedBug,
  clearSelectedBug,
  setSearchFilter,
  setStatusFilter,
  setReleaseFilter,
  applyFilters,
} = bugsSlice.actions;

export default bugsSlice.reducer;

