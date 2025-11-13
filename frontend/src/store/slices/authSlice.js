import { createSlice } from '@reduxjs/toolkit';

const initialState = {
  user: null,
  email: null,
  role: null, // 'developer' or 'manager'
  isAuthenticated: false,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    login: (state, action) => {
      state.email = action.payload.email;
      state.role = action.payload.role;
      state.isAuthenticated = true;
      state.user = {
        email: action.payload.email,
        role: action.payload.role,
      };
    },
    logout: (state) => {
      state.user = null;
      state.email = null;
      state.role = null;
      state.isAuthenticated = false;
    },
    setUser: (state, action) => {
      state.user = action.payload;
      state.email = action.payload.email;
      state.role = action.payload.role;
      state.isAuthenticated = true;
    },
  },
});

export const { login, logout, setUser } = authSlice.actions;
export default authSlice.reducer;

