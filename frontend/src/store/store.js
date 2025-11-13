import { configureStore } from '@reduxjs/toolkit';
import authReducer from './slices/authSlice';
import bugsReducer from './slices/bugsSlice';
import releaseNotesReducer from './slices/releaseNotesSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    bugs: bugsReducer,
    releaseNotes: releaseNotesReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['bugs/fetchBugs/fulfilled'],
      },
    }),
});

export default store;

