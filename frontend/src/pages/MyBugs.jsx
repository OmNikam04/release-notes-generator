import { useState, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import BugCard from '../components/BugCard';
import BugDetailModal from '../components/BugDetailModal';
import {
  fetchBugs,
  setSelectedBug,
  clearSelectedBug,
  clearBugs,
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

  // Individual column loading states
  const [columnLoading, setColumnLoading] = useState({
    ai_generated: false,
    dev_approved: false,
    mgr_approved: false
  });

  useEffect(() => {
    // Fetch bugs from backend API on component mount
    const fetchBugsFromAPI = async () => {
      setApiLoading(true);
      setApiError('');
      setColumnLoading({ ai_generated: true, dev_approved: true, mgr_approved: true });

      try {
        console.log('[MyBugs] Fetching Kanban columns from backend API for user:', user?.email, 'role:', user?.role);
        console.log('[MyBugs] Column 1 (AI Generated) - Loading...');
        console.log('[MyBugs] Column 2 (Dev Approved) - Loading...');
        console.log('[MyBugs] Column 3 (Manager Approved) - Loading...');

        // Determine filter based on user role
        // Developer: See bugs assigned to me
        // Manager: See bugs where I'm the manager (my team's bugs)
        const isManager = user?.role === 'manager';
        const filterParams = isManager
          ? { manager_id: true }      // Manager sees team's bugs
          : { assigned_to_me: true }; // Developer sees own bugs

        console.log('[MyBugs] Using filter:', isManager ? 'manager_id=true (team bugs)' : 'assigned_to_me=true (my bugs)');

        // Fetch each Kanban column separately based on status
        const [pendingResponse, devApprovedResponse, mgrApprovedResponse] = await Promise.all([
          bugsAPI.getReleaseNotes({ status: 'ai_generated', ...filterParams }),
          bugsAPI.getReleaseNotes({ status: 'dev_approved', ...filterParams }),
          bugsAPI.getReleaseNotes({ status: 'mgr_approved', ...filterParams })
        ]);

        console.log('[MyBugs] Column 1 (AI Generated) - Loaded:', pendingResponse.release_notes.length, 'bugs');
        console.log('[MyBugs] Column 2 (Dev Approved) - Loaded:', devApprovedResponse.release_notes.length, 'bugs');
        console.log('[MyBugs] Column 3 (Manager Approved) - Loaded:', mgrApprovedResponse.release_notes.length, 'bugs');

        // Combine all bugs from all columns
        const allBugs = [
          ...pendingResponse.release_notes,
          ...devApprovedResponse.release_notes,
          ...mgrApprovedResponse.release_notes
        ];

        // Transform backend data to match frontend format
        const transformedBugs = allBugs.map(note => ({
          ...note,
          generated_note: note.generated_note || null, // Explicitly set to null if not present
        }));

        console.log('[MyBugs] Total bugs from all columns:', transformedBugs.length);
        console.log('[MyBugs] Transformed bugs:', transformedBugs);

        // Dispatch to Redux store
        dispatch(fetchBugs(transformedBugs));
      } catch (error) {
        console.error('[MyBugs] ❌ AUTHORIZATION ERROR - Failed to fetch Kanban columns from API');
        console.error('[MyBugs] Error message:', error.message);
        console.error('[MyBugs] Full error:', error);

        const errorMsg = error.message || 'Failed to load bugs';
        console.log('[MyBugs] Setting error message:', errorMsg);

        setApiError(errorMsg);
        setShowError(true);

        console.log('[MyBugs] ✅ Error state updated - showError: true, apiError:', errorMsg);
        console.log('[MyBugs] Error toast should now be visible at bottom-right corner');

        // Auto-hide error after 10 seconds (longer duration for visibility)
        const errorTimeout = setTimeout(() => {
          console.log('[MyBugs] Auto-hiding error toast after 10 seconds');
          setShowError(false);
        }, 10000);

        // Fallback to empty bugs list
        dispatch(fetchBugs([]));

        // Cleanup timeout if component unmounts
        return () => clearTimeout(errorTimeout);
      } finally {
        setApiLoading(false);
        setColumnLoading({ ai_generated: false, dev_approved: false, mgr_approved: false });
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

      // Dispatch logout action to Redux (clears auth state)
      console.log('[MyBugs] Dispatching logout action to Redux');
      dispatch(logout());

      // Clear bugs data from Redux
      console.log('[MyBugs] Clearing bugs data from Redux');
      dispatch(clearBugs());

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
      dispatch(clearBugs());
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
  // Refresh Kanban board after approval
  const handleRefreshKanban = async () => {
    console.log('[MyBugs] Refreshing Kanban board after approval...');
    setColumnLoading({ ai_generated: true, dev_approved: true, mgr_approved: true });

    try {
      // Determine filter based on user role
      const isManager = user?.role === 'manager';
      const filterParams = isManager
        ? { manager_id: true }      // Manager sees team's bugs
        : { assigned_to_me: true }; // Developer sees own bugs

      // Fetch each Kanban column separately based on status
      const [pendingResponse, devApprovedResponse, mgrApprovedResponse] = await Promise.all([
        bugsAPI.getReleaseNotes({ status: 'ai_generated', ...filterParams }),
        bugsAPI.getReleaseNotes({ status: 'dev_approved', ...filterParams }),
        bugsAPI.getReleaseNotes({ status: 'mgr_approved', ...filterParams })
      ]);

      console.log('[MyBugs] Kanban refresh - Column 1 (AI Generated):', pendingResponse.release_notes.length, 'bugs');
      console.log('[MyBugs] Kanban refresh - Column 2 (Dev Approved):', devApprovedResponse.release_notes.length, 'bugs');
      console.log('[MyBugs] Kanban refresh - Column 3 (Manager Approved):', mgrApprovedResponse.release_notes.length, 'bugs');

      // Combine all bugs from all columns
      const allBugs = [
        ...pendingResponse.release_notes,
        ...devApprovedResponse.release_notes,
        ...mgrApprovedResponse.release_notes
      ];

      // Transform backend data to match frontend format
      const transformedBugs = allBugs.map(note => ({
        ...note,
        generated_note: note.generated_note || null,
      }));

      console.log('[MyBugs] Kanban refresh - Total bugs:', transformedBugs.length);

      // Dispatch to Redux store
      dispatch(fetchBugs(transformedBugs));

      // Close the modal after refresh
      dispatch(clearSelectedBug());
      console.log('[MyBugs] ✅ Kanban board refreshed successfully');
    } catch (error) {
      console.error('[MyBugs] ❌ Failed to refresh Kanban board:', error);
      setApiError('Failed to refresh Kanban board');
      setShowError(true);
    } finally {
      setColumnLoading({ ai_generated: false, dev_approved: false, mgr_approved: false });
    }
  };

  const getBugsByColumn = () => {
    const columns = {
      ai_generated: [],
      dev_approved: [],
      approved: []
    };

    filteredBugs.forEach(bug => {
      if (bug.status === 'ai_generated') {
        columns.ai_generated.push(bug);
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
          <h1>ReleaseNotegenerator</h1>
          <p className="page-subtitle">
            {user?.role === 'manager'
              ? 'Review and approve release notes from your team members.'
              : 'Manage and track your assigned bugs. Generate release notes with AI assistance.'
            }
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
            className="admin-button"
            onClick={() => navigate('/releaseadmin')}
            title="Release Admin Panel"
          >
            🔧
          </button>
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
          <p>⏳ Loading bugs from backend...</p>
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
          {/* Column 1: Raw AI Generated */}
          <div className="kanban-column">
            <div className="column-header">
              <h3 className="column-title">
                <span className="column-icon">🤖</span>
                Raw AI Generated
              </h3>
              <span className="column-count">{columnLoading.ai_generated ? '⏳' : kanbanColumns.ai_generated.length}</span>
            </div>
            <div className="column-content">
              {columnLoading.ai_generated ? (
                <div className="empty-column-message">
                  <p>Loading AI generated bugs...</p>
                </div>
              ) : kanbanColumns.ai_generated.length > 0 ? (
                kanbanColumns.ai_generated.map(bug => (
                  <BugCard
                    key={bug.id}
                    bug={bug}
                    onClick={() => handleBugClick(bug)}
                  />
                ))
              ) : (
                <div className="empty-column-message">
                  <p>No AI generated bugs</p>
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
              <span className="column-count">{columnLoading.dev_approved ? '⏳' : kanbanColumns.dev_approved.length}</span>
            </div>
            <div className="column-content">
              {columnLoading.dev_approved ? (
                <div className="empty-column-message">
                  <p>Loading developer approved bugs...</p>
                </div>
              ) : kanbanColumns.dev_approved.length > 0 ? (
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
              <span className="column-count">{columnLoading.mgr_approved ? '⏳' : kanbanColumns.approved.length}</span>
            </div>
            <div className="column-content">
              {columnLoading.mgr_approved ? (
                <div className="empty-column-message">
                  <p>Loading manager approved bugs...</p>
                </div>
              ) : kanbanColumns.approved.length > 0 ? (
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
          onApprovalChange={handleRefreshKanban}
        />
      )}
    </div>
  );
};

export default MyBugs;
