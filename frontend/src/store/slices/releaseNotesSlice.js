import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';

// Async thunk for generating release note
export const generateReleaseNote = createAsyncThunk(
  'releaseNotes/generate',
  async ({ bugId }, { rejectWithValue }) => {
    try {
      // TODO: Replace with actual API call
      // const response = await fetch('/api/release-notes/generate', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ bugId }),
      // });
      // const data = await response.json();
      // return data;
      
      // For now, return mock data
      return {
        bugId,
        content: 'Fixed critical authentication bug that prevented users from logging in with valid credentials.',
        alternatives: [
          'Resolved login issue affecting user authentication with valid credentials.',
          'Corrected authentication service bug preventing successful user login.',
          'Fixed bug in authentication flow that blocked valid user logins.',
        ],
      };
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

// Async thunk for regenerating with feedback
export const regenerateReleaseNote = createAsyncThunk(
  'releaseNotes/regenerate',
  async ({ bugId, feedback }, { rejectWithValue }) => {
    try {
      // TODO: Replace with actual API call
      // const response = await fetch('/api/release-notes/regenerate', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ bugId, feedback }),
      // });
      // const data = await response.json();
      // return data;
      
      // For now, return mock data
      return {
        bugId,
        content: `Improved release note based on feedback: ${feedback}`,
        alternatives: [
          'Alternative version 1 with feedback applied.',
          'Alternative version 2 with feedback applied.',
          'Alternative version 3 with feedback applied.',
        ],
      };
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

// Async thunk for sending to approval
export const sendToApproval = createAsyncThunk(
  'releaseNotes/sendToApproval',
  async ({ bugId, content }, { rejectWithValue }) => {
    try {
      // TODO: Replace with actual API call
      return { bugId, status: 'dev_approved' };
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

// Async thunk for manager approval
export const approveReleaseNote = createAsyncThunk(
  'releaseNotes/approve',
  async ({ bugId }, { rejectWithValue }) => {
    try {
      // TODO: Replace with actual API call
      return { bugId, status: 'approved' };
    } catch (error) {
      return rejectWithValue(error.message);
    }
  }
);

const initialState = {
  releaseNotes: {}, // { bugId: { content, alternatives, status, ... } }
  currentNote: null,
  loading: false,
  error: null,
};

const releaseNotesSlice = createSlice({
  name: 'releaseNotes',
  initialState,
  reducers: {
    setCurrentNote: (state, action) => {
      state.currentNote = action.payload;
    },
    updateNoteContent: (state, action) => {
      const { bugId, content } = action.payload;
      if (state.releaseNotes[bugId]) {
        state.releaseNotes[bugId].content = content;
      }
    },
    clearCurrentNote: (state) => {
      state.currentNote = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Generate release note
      .addCase(generateReleaseNote.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(generateReleaseNote.fulfilled, (state, action) => {
        state.loading = false;
        const { bugId, content, alternatives } = action.payload;
        state.releaseNotes[bugId] = {
          content,
          alternatives,
          status: 'pending',
        };
        state.currentNote = state.releaseNotes[bugId];
      })
      .addCase(generateReleaseNote.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
      })
      // Regenerate with feedback
      .addCase(regenerateReleaseNote.fulfilled, (state, action) => {
        const { bugId, content, alternatives } = action.payload;
        state.releaseNotes[bugId] = {
          ...state.releaseNotes[bugId],
          content,
          alternatives,
        };
        state.currentNote = state.releaseNotes[bugId];
      })
      // Send to approval
      .addCase(sendToApproval.fulfilled, (state, action) => {
        const { bugId, status } = action.payload;
        if (state.releaseNotes[bugId]) {
          state.releaseNotes[bugId].status = status;
        }
      })
      // Approve
      .addCase(approveReleaseNote.fulfilled, (state, action) => {
        const { bugId, status } = action.payload;
        if (state.releaseNotes[bugId]) {
          state.releaseNotes[bugId].status = status;
        }
      });
  },
});

export const { setCurrentNote, updateNoteContent, clearCurrentNote } = releaseNotesSlice.actions;
export default releaseNotesSlice.reducer;

