// API service for backend communication
const API_BASE_URL = 'http://localhost:8080/api/v1';

// Helper function to make API calls
const apiCall = async (endpoint, method = 'GET', body = null) => {
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  // Add authorization token if it exists
  const token = localStorage.getItem('authToken');
  if (token) {
    options.headers.Authorization = `Bearer ${token}`;
  }

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || 'API request failed');
    }

    return data;
  } catch (error) {
    console.error(`API Error [${method} ${endpoint}]:`, error);
    throw error;
  }
};

// User/Auth API calls
export const authAPI = {
  // Login with email and role
  login: async (email, role) => {
    console.log('[API] Sending login request to backend:', { email, role });
    
    try {
      const response = await apiCall('/user/login', 'POST', { email, role });
      
      console.log('[API] Login response received from backend:', response);
      console.log('[API] User data from DB:', response.data.user);
      console.log('[API] JWT Token:', response.data.token);
      console.log('[API] Refresh Token:', response.data.refresh_token);
      
      return response.data;
    } catch (error) {
      console.error('[API] Login failed:', error.message);
      throw error;
    }
  },

  // Refresh access token
  refreshToken: async (refreshToken) => {
    console.log('[API] Refreshing access token');
    
    try {
      const response = await apiCall('/user/refresh', 'POST', { refresh_token: refreshToken });
      
      console.log('[API] Token refresh successful');
      console.log('[API] New JWT Token:', response.data.token);
      
      return response.data;
    } catch (error) {
      console.error('[API] Token refresh failed:', error.message);
      throw error;
    }
  },

  // Logout
  logout: async (refreshToken) => {
    console.log('[API] Sending logout request');
    
    try {
      const response = await apiCall('/user/logout', 'POST', { refresh_token: refreshToken });
      
      console.log('[API] Logout successful');
      return response;
    } catch (error) {
      console.error('[API] Logout failed:', error.message);
      throw error;
    }
  },

  // Get current user info
  getCurrentUser: async () => {
    console.log('[API] Fetching current user info');
    
    try {
      const response = await apiCall('/user/me', 'GET');
      
      console.log('[API] Current user data from DB:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch current user:', error.message);
      throw error;
    }
  },
};

// Bugs API calls
export const bugsAPI = {
  // Get all bugs
  getBugs: async () => {
    console.log('[API] Fetching bugs from backend');
    
    try {
      const response = await apiCall('/bugs/', 'GET');
      
      console.log('[API] Bugs data from DB:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bugs:', error.message);
      throw error;
    }
  },

  // Get single bug
  getBug: async (bugId) => {
    console.log('[API] Fetching bug:', bugId);
    
    try {
      const response = await apiCall(`/bugs/${bugId}`, 'GET');
      
      console.log('[API] Bug data from DB:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bug:', error.message);
      throw error;
    }
  },
};

export default { authAPI, bugsAPI };

