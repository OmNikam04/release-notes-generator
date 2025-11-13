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
  // Get all bugs with optional filters
  getBugs: async (filters = {}) => {
    console.log('[API] Fetching bugs from backend with filters:', filters);

    try {
      // Build query string from filters
      const queryParams = new URLSearchParams();
      if (filters.release) queryParams.append('release', filters.release);
      if (filters.status) queryParams.append('status', filters.status);
      if (filters.severity) queryParams.append('severity', filters.severity);
      if (filters.bug_type) queryParams.append('bug_type', filters.bug_type);
      if (filters.assigned_to) queryParams.append('assigned_to', filters.assigned_to);
      if (filters.manager_id) queryParams.append('manager_id', filters.manager_id);
      if (filters.page) queryParams.append('page', filters.page);
      if (filters.limit) queryParams.append('limit', filters.limit);

      const queryString = queryParams.toString();
      const endpoint = queryString ? `/bugs?${queryString}` : '/bugs';

      const response = await apiCall(endpoint, 'GET');

      console.log('[API] Bugs data from DB:', response.data);
      console.log('[API] Total bugs:', response.data.pagination?.total || response.data.bugs?.length);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bugs:', error.message);
      throw error;
    }
  },

  // Get single bug by UUID
  getBug: async (bugId) => {
    console.log('[API] Fetching bug by ID:', bugId);

    try {
      const response = await apiCall(`/bugs/${bugId}`, 'GET');

      console.log('[API] Bug data from DB:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bug:', error.message);
      throw error;
    }
  },

  // Get bug by Bugsby ID
  getBugByBugsbyId: async (bugsbyId) => {
    console.log('[API] Fetching bug by Bugsby ID:', bugsbyId);

    try {
      const response = await apiCall(`/bugs/bugsby/${bugsbyId}`, 'GET');

      console.log('[API] Bug data from DB:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bug by Bugsby ID:', error.message);
      throw error;
    }
  },

  // Update bug
  updateBug: async (bugId, updates) => {
    console.log('[API] Updating bug:', bugId, 'with data:', updates);

    try {
      const response = await apiCall(`/bugs/${bugId}`, 'PUT', updates);

      console.log('[API] Bug updated successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to update bug:', error.message);
      throw error;
    }
  },

  // Delete bug
  deleteBug: async (bugId) => {
    console.log('[API] Deleting bug:', bugId);

    try {
      const response = await apiCall(`/bugs/${bugId}`, 'DELETE');

      console.log('[API] Bug deleted successfully');
      return response;
    } catch (error) {
      console.error('[API] Failed to delete bug:', error.message);
      throw error;
    }
  },
};

export default { authAPI, bugsAPI };

