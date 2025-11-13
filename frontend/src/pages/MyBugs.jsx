import { useState, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import BugCard from '../components/BugCard';
import BugDetailModal from '../components/BugDetailModal';
import {
  fetchBugs,
  setSelectedBug,
  clearSelectedBug,
  setSearchFilter,
  setStatusFilter,
  setReleaseFilter,
  applyFilters,
} from '../store/slices/bugsSlice';
import { logout } from '../store/slices/authSlice';
import { bugsAPI, authAPI } from '../services/api';
import './MyBugs.css';

const MyBugs = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { filteredBugs, selectedBug, filters } = useSelector((state) => state.bugs);
  const { user } = useSelector((state) => state.auth);
  const [apiLoading, setApiLoading] = useState(false);
  const [apiError, setApiError] = useState('');
  const [logoutLoading, setLogoutLoading] = useState(false);
  const [showError, setShowError] = useState(false);

  useEffect(() => {
    // Fetch bugs from backend API on component mount
    const fetchBugsFromAPI = async () => {
      setApiLoading(true);
      setApiError('');

      try {
        console.log('[MyBugs] Fetching bugs from backend API for user:', user?.email);

        // Fetch bugs from backend
        const response = await bugsAPI.getBugs();

        console.log('[MyBugs] Bugs fetched from backend:', response.bugs);
        console.log('[MyBugs] Total bugs:', response.pagination?.total || response.bugs?.length);
        console.log('[MyBugs] Pagination info:', response.pagination);

        // Transform backend data to match frontend format
        // Backend returns bugs with generated_note as null initially
        const transformedBugs = response.bugs.map(bug => ({
          ...bug,
          generated_note: bug.generated_note || null, // Explicitly set to null if not present
        }));

        console.log('[MyBugs] Transformed bugs:', transformedBugs);

        // Dispatch to Redux store
        dispatch(fetchBugs(transformedBugs));
      } catch (error) {
        console.error('[MyBugs] Failed to fetch bugs from API:', error.message);
        const errorMsg = error.message || 'Failed to load bugs';
        setApiError(errorMsg);
        setShowError(true);
        // Auto-hide error after 5 seconds
        setTimeout(() => setShowError(false), 5000);
        // Fallback to empty bugs list
        dispatch(fetchBugs([]));
      } finally {
        setApiLoading(false);
      }
    };

    fetchBugsFromAPI();
  }, [dispatch, user]);

  useEffect(() => {
    // Apply filters whenever filter values change
    dispatch(applyFilters());
  }, [filters.search, filters.status, filters.release, dispatch]);

  const handleSearchChange = (value) => {
    dispatch(setSearchFilter(value));
  };

  const handleStatusFilterChange = (value) => {
    dispatch(setStatusFilter(value));
  };

  const handleReleaseFilterChange = (value) => {
    dispatch(setReleaseFilter(value));
  };

  const handleBugClick = (bug) => {
    dispatch(setSelectedBug(bug));
  };

  const handleCloseModal = () => {
    dispatch(clearSelectedBug());
  };

  const handleLogout = async () => {
    try {
      setLogoutLoading(true);
      console.log('[MyBugs] Starting logout process');

      // Get refresh token from localStorage
      const refreshToken = localStorage.getItem('refreshToken');

      // Call logout API
      if (refreshToken) {
        console.log('[MyBugs] Calling logout API');
        await authAPI.logout(refreshToken);
        console.log('[MyBugs] Logout API successful');
      }

      // Clear localStorage
      console.log('[MyBugs] Clearing localStorage');
      localStorage.removeItem('authToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('userEmail');
      localStorage.removeItem('userRole');
      localStorage.removeItem('userId');

      // Dispatch logout action to Redux
      console.log('[MyBugs] Dispatching logout action to Redux');
      dispatch(logout());

      // Redirect to login page
      console.log('[MyBugs] Redirecting to login page');
      navigate('/');
    } catch (error) {
      console.error('[MyBugs] Logout failed:', error.message);
      // Still clear local state and redirect even if API call fails
      localStorage.removeItem('authToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('userEmail');
      localStorage.removeItem('userRole');
      localStorage.removeItem('userId');
      dispatch(logout());
      navigate('/');
    } finally {
      setLogoutLoading(false);
    }
  };

  const bugs = useSelector((state) => state.bugs.bugs);

  const getUniqueStatuses = () => {
    return [...new Set(bugs.map(bug => bug.status))];
  };

  const getUniqueReleases = () => {
    return [...new Set(bugs.map(bug => bug.release).filter(Boolean))];
  };

  // Group bugs by Kanban columns
  const getBugsByColumn = () => {
    const columns = {
      pending: [],
      dev_approved: [],
      approved: []
    };

    filteredBugs.forEach(bug => {
      if (bug.status === 'pending') {
        columns.pending.push(bug);
      } else if (bug.status === 'dev_approved') {
        columns.dev_approved.push(bug);
      } else if (bug.status === 'mgr_approved' || bug.status === 'approved') {
        columns.approved.push(bug);
      }
    });

    return columns;
  };

  const kanbanColumns = getBugsByColumn();

  return (
    <div className="my-bugs-page">
      <div className="page-header">
        <div className="header-left">
          <h1>🔥 NoteForge</h1>
          <p className="page-subtitle">
            Manage and track your assigned bugs. Generate release notes with AI assistance.
          </p>
        </div>
        <div className="header-right">
          <div className="user-info">
            <div className="user-avatar">👤</div>
            <div className="user-details">
              <span className="user-name">{user?.email || 'User'}</span>
              <span className="user-role">{user?.role ? user.role.charAt(0).toUpperCase() + user.role.slice(1) : 'Developer'}</span>
            </div>
          </div>
          <button
            className="logout-button"
            onClick={handleLogout}
            disabled={logoutLoading}
            title="Logout"
          >
            {logoutLoading ? '⏳' : '⚡'}
          </button>
        </div>
      </div>

      {/* Loading State */}
      {apiLoading && (
        <div className="loading-message">
          <p>Loading bugs from backend...</p>
        </div>
      )}

      {/* Error Alert - Corner Toast */}
      {showError && apiError && (
        <div className="error-toast">
          <div className="error-toast-content">
            <span className="error-icon">⚠️</span>
            <span className="error-text">{apiError}</span>
          </div>
          <button
            className="error-close-btn"
            onClick={() => setShowError(false)}
            title="Close"
          >
            ✕
          </button>
        </div>
      )}

      <div className="page-content">
        <div className="filters-section">
          <div className="filters-row">
            <div className="search-and-filters">
              <div className="search-bar">
                <input
                  type="text"
                  placeholder="Search bugs by title, description, or package..."
                  value={filters.search}
                  onChange={(e) => handleSearchChange(e.target.value)}
                  className="search-input"
                />
              </div>

              <div className="filter-controls">
                <select
                  value={filters.release}
                  onChange={(e) => handleReleaseFilterChange(e.target.value)}
                  className="filter-select"
                >
                  <option value="all">All Releases</option>
                  {getUniqueReleases().map(release => (
                    <option key={release} value={release}>{release}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Kanban Board */}
        <div className="kanban-board">
          {/* Column 1: Raw AI Generated / Pending */}
          <div className="kanban-column">
            <div className="column-header">
              <h3 className="column-title">
                <span className="column-icon">🤖</span>
                Raw AI Generated
              </h3>
              <span className="column-count">{kanbanColumns.pending.length}</span>
            </div>
            <div className="column-content">
              {kanbanColumns.pending.length > 0 ? (
                kanbanColumns.pending.map(bug => (
                  <BugCard
                    key={bug.id}
                    bug={bug}
                    onClick={() => handleBugClick(bug)}
                  />
                ))
              ) : (
                <div className="empty-column-message">
                  <p>No pending bugs</p>
                </div>
              )}
            </div>
          </div>

          {/* Column 2: Developer Approved */}
          <div className="kanban-column">
            <div className="column-header">
              <h3 className="column-title">
                <span className="column-icon">👨‍💻</span>
                Developer Approved
              </h3>
              <span className="column-count">{kanbanColumns.dev_approved.length}</span>
            </div>
            <div className="column-content">
              {kanbanColumns.dev_approved.length > 0 ? (
                kanbanColumns.dev_approved.map(bug => (
                  <BugCard
                    key={bug.id}
                    bug={bug}
                    onClick={() => handleBugClick(bug)}
                  />
                ))
              ) : (
                <div className="empty-column-message">
                  <p>No developer approved bugs</p>
                </div>
              )}
            </div>
          </div>

          {/* Column 3: Manager Approved / Release Note Approved */}
          <div className="kanban-column">
            <div className="column-header">
              <h3 className="column-title">
                <span className="column-icon">✅</span>
                Release Note Approved
              </h3>
              <span className="column-count">{kanbanColumns.approved.length}</span>
            </div>
            <div className="column-content">
              {kanbanColumns.approved.length > 0 ? (
                kanbanColumns.approved.map(bug => (
                  <BugCard
                    key={bug.id}
                    bug={bug}
                    onClick={() => handleBugClick(bug)}
                  />
                ))
              ) : (
                <div className="empty-column-message">
                  <p>No approved bugs</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {selectedBug && (
        <BugDetailModal
          bug={selectedBug}
          onClose={handleCloseModal}
        />
      )}
    </div>
  );
};

export default MyBugs;
