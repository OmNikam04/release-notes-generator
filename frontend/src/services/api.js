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
      // Handle different error types
      let errorMessage = data.message || 'API request failed';

      if (response.status === 401) {
        errorMessage = 'Unauthorized: ' + (data.message || 'Please login again');
      } else if (response.status === 403) {
        errorMessage = 'Forbidden: ' + (data.message || 'You do not have permission');
      } else if (response.status === 404) {
        errorMessage = 'Not found: ' + (data.message || 'Resource not found');
      } else if (response.status === 500) {
        errorMessage = 'Server error: ' + (data.message || 'Please try again later');
      }

      console.error(`[API Error] ${method} ${endpoint} - Status: ${response.status} - Message: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    return data;
  } catch (error) {
    console.error(`[API Error] [${method} ${endpoint}]:`, error.message);
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
  // Get all bugs with optional filters (from /bugs endpoint)
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

  // Get release notes with bugs (Kanban view) - GET /release-notes
  getReleaseNotes: async (filters = {}) => {
    console.log('[API] Fetching release notes (Kanban view) with filters:', filters);

    try {
      // Build query string from filters
      const queryParams = new URLSearchParams();
      if (filters.status) queryParams.append('status', filters.status);
      if (filters.release) queryParams.append('release', filters.release);
      if (filters.component) queryParams.append('component', filters.component);
      if (filters.assigned_to_me) queryParams.append('assigned_to_me', filters.assigned_to_me);
      if (filters.manager_id) queryParams.append('manager_id', filters.manager_id);
      if (filters.page) queryParams.append('page', filters.page);
      if (filters.limit) queryParams.append('limit', filters.limit);
      if (filters.sort_by) queryParams.append('sort_by', filters.sort_by);
      if (filters.sort_order) queryParams.append('sort_order', filters.sort_order);

      const queryString = queryParams.toString();
      const endpoint = queryString ? `/release-notes?${queryString}` : '/release-notes';

      const response = await apiCall(endpoint, 'GET');

      console.log('[API] Release notes data from DB:', response.data);
      console.log('[API] Total release notes:', response.data.pagination?.total || response.data.release_notes?.length);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch release notes:', error.message);
      throw error;
    }
  },

  // Get bugs by status (for Kanban columns)
  getBugsByStatus: async (status, filters = {}) => {
    console.log('[API] Fetching bugs by status:', status, 'with filters:', filters);

    try {
      const queryParams = new URLSearchParams();
      queryParams.append('status', status);
      if (filters.release) queryParams.append('release', filters.release);
      if (filters.severity) queryParams.append('severity', filters.severity);
      if (filters.bug_type) queryParams.append('bug_type', filters.bug_type);
      if (filters.assigned_to) queryParams.append('assigned_to', filters.assigned_to);
      if (filters.manager_id) queryParams.append('manager_id', filters.manager_id);
      if (filters.page) queryParams.append('page', filters.page);
      if (filters.limit) queryParams.append('limit', filters.limit);

      const queryString = queryParams.toString();
      const endpoint = `/bugs?${queryString}`;

      const response = await apiCall(endpoint, 'GET');

      console.log('[API] Bugs with status', status, ':', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch bugs by status:', error.message);
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

// Release Notes API calls
export const releaseNotesAPI = {
  // Get release notes by status (for Kanban columns)
  getByStatus: async (status, filters = {}) => {
    console.log('[API] Fetching release notes by status:', status, 'with filters:', filters);

    try {
      const queryParams = new URLSearchParams();
      queryParams.append('status', status);
      if (filters.release) queryParams.append('release', filters.release);
      if (filters.component) queryParams.append('component', filters.component);
      if (filters.assigned_to_me) queryParams.append('assigned_to_me', filters.assigned_to_me);
      if (filters.manager_id) queryParams.append('manager_id', filters.manager_id);
      if (filters.page) queryParams.append('page', filters.page);
      if (filters.limit) queryParams.append('limit', filters.limit);
      if (filters.sort_by) queryParams.append('sort_by', filters.sort_by);
      if (filters.sort_order) queryParams.append('sort_order', filters.sort_order);

      const queryString = queryParams.toString();
      const endpoint = `/release-notes?${queryString}`;

      const response = await apiCall(endpoint, 'GET');

      console.log('[API] Release notes with status', status, ':', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch release notes by status:', error.message);
      throw error;
    }
  },

  // Get all release notes (Kanban view)
  getAll: async (filters = {}) => {
    console.log('[API] Fetching all release notes with filters:', filters);

    try {
      const queryParams = new URLSearchParams();
      if (filters.release) queryParams.append('release', filters.release);
      if (filters.component) queryParams.append('component', filters.component);
      if (filters.assigned_to_me) queryParams.append('assigned_to_me', filters.assigned_to_me);
      if (filters.manager_id) queryParams.append('manager_id', filters.manager_id);
      if (filters.page) queryParams.append('page', filters.page);
      if (filters.limit) queryParams.append('limit', filters.limit);
      if (filters.sort_by) queryParams.append('sort_by', filters.sort_by);
      if (filters.sort_order) queryParams.append('sort_order', filters.sort_order);

      const queryString = queryParams.toString();
      const endpoint = queryString ? `/release-notes?${queryString}` : '/release-notes';

      const response = await apiCall(endpoint, 'GET');

      console.log('[API] All release notes:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch all release notes:', error.message);
      throw error;
    }
  },

  // Get release note by bug ID (includes AI metadata and alternatives)
  getReleaseNoteByBugId: async (bugId) => {
    console.log('[API] Fetching release note for bug:', bugId);

    try {
      const response = await apiCall(`/release-notes/bug/${bugId}`, 'GET');
      console.log('[API] Release note fetched successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to fetch release note:', error.message);
      throw error;
    }
  },

  // Generate release note for a bug
  generateReleaseNote: async (bugId) => {
    console.log('[API] Generating release note for bug:', bugId);

    try {
      const response = await apiCall('/release-notes/generate', 'POST', { bug_id: bugId });
      console.log('[API] Release note generated successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to generate release note:', error.message);
      throw error;
    }
  },

  // Update release note (developer approval)
  updateReleaseNote: async (releaseNoteId, content, status) => {
    console.log('[API] Updating release note:', releaseNoteId, 'with status:', status);

    try {
      const response = await apiCall(`/release-notes/${releaseNoteId}`, 'PUT', {
        content,
        status
      });
      console.log('[API] Release note updated successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to update release note:', error.message);
      throw error;
    }
  },

  // Approve or reject release note (manager only)
  approveReleaseNote: async (releaseNoteId, action, correctedContent = null, feedback = '') => {
    console.log('[API] Approving/rejecting release note:', releaseNoteId, 'action:', action);
    if (correctedContent) {
      console.log('[API] Manager provided corrected content');
    }
    if (feedback) {
      console.log('[API] Manager provided feedback:', feedback);
    }

    try {
      const body = {
        action,
      };

      // Only include corrected_content if it's provided and different from original
      if (correctedContent) {
        body.corrected_content = correctedContent;
      }

      // Only include feedback if it's provided
      if (feedback) {
        body.feedback = feedback;
      }

      const response = await apiCall(`/release-notes/${releaseNoteId}/approve`, 'POST', body);
      console.log('[API] Release note approval processed:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to process approval:', error.message);
      throw error;
    }
  },
};

// Bugsby Sync API calls (Manager only)
export const syncAPI = {
  // Sync a single bug by Bugsby ID
  syncBugById: async (bugsby_id) => {
    console.log('[API] Syncing single bug by ID:', bugsby_id);

    try {
      const response = await apiCall(`/bugsby/sync/${bugsby_id}`, 'POST');
      console.log('[API] Bug synced successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to sync bug:', error.message);
      throw error;
    }
  },

  // Sync all bugs for a release
  syncRelease: async (filters = {}) => {
    console.log('[API] Syncing release with filters:', filters);

    try {
      const body = {
        release: filters.release,
      };

      if (filters.status) body.status = filters.status;
      if (filters.severity) body.severity = filters.severity;
      if (filters.bug_type) body.bug_type = filters.bug_type;
      if (filters.component) body.component = filters.component;

      const response = await apiCall('/bugsby/sync', 'POST', body);
      console.log('[API] Release synced successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to sync release:', error.message);
      throw error;
    }
  },

  // Get sync status for a release
  getSyncStatus: async (release) => {
    console.log('[API] Getting sync status for release:', release);

    try {
      const queryParams = new URLSearchParams();
      queryParams.append('release', release);

      const response = await apiCall(`/bugsby/status?${queryParams.toString()}`, 'GET');
      console.log('[API] Sync status retrieved:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to get sync status:', error.message);
      throw error;
    }
  },

  // Custom Bugsby Query (Query 11 - Testing endpoint, no auth required)
  customBugsbyQuery: async (queryData) => {
    console.log('[API] Executing custom Bugsby query:', queryData);

    try {
      const body = {
        query: queryData.query,
        limit: queryData.limit || '100',
        sortBy: queryData.sortBy || 'lastUpdateTime',
        order: queryData.order || 'desc',
        source: queryData.source || 'mysql',
        textQueryMode: queryData.textQueryMode || 'default'
      };

      const response = await apiCall('/bugsby-api/bugs/query', 'POST', body);
      console.log('[API] Custom query executed successfully:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to execute custom query:', error.message);
      throw error;
    }
  },

  // Get Bugs by Assignee (Query 10 - Testing endpoint, no auth required)
  getBugsByAssignee: async (email, limit = 100) => {
    console.log('[API] Getting bugs for assignee:', email);

    try {
      const queryParams = new URLSearchParams();
      queryParams.append('limit', limit);

      const response = await apiCall(`/bugsby-api/bugs/assignee/${email}?${queryParams.toString()}`, 'GET');
      console.log('[API] Bugs retrieved for assignee:', response.data);
      return response.data;
    } catch (error) {
      console.error('[API] Failed to get bugs by assignee:', error.message);
      throw error;
    }
  },
};

export default { authAPI, bugsAPI, releaseNotesAPI, syncAPI };

