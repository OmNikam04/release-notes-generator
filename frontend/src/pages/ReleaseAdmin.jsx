import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { logout } from '../store/slices/authSlice';
import { syncAPI, authAPI } from '../services/api';
import './ReleaseAdmin.css';

const ReleaseAdmin = () => {
  const user = useSelector((state) => state.auth.user);
  const navigate = useNavigate();
  const dispatch = useDispatch();
  
  // State for Sync by Bug ID
  const [bugId, setBugId] = useState('');
  const [bugIdLoading, setBugIdLoading] = useState(false);
  const [bugIdResult, setBugIdResult] = useState(null);
  const [bugIdError, setBugIdError] = useState('');
  
  // State for Query by Blocking Bugs
  const [blockingBugId, setBlockingBugId] = useState('');
  const [blockingLoading, setBlockingLoading] = useState(false);
  const [blockingResult, setBlockingResult] = useState(null);
  const [blockingError, setBlockingError] = useState('');

  // Check authentication on component mount
  useEffect(() => {
    console.log('[ReleaseAdmin] Checking authentication...');

    // Check if user is authenticated
    if (!user || !user.email) {
      console.log('[ReleaseAdmin] ❌ User not authenticated - redirecting to login');
      navigate('/');
      return;
    }

    // Check if token exists in localStorage
    const authToken = localStorage.getItem('authToken');
    if (!authToken) {
      console.log('[ReleaseAdmin] ❌ No auth token found - redirecting to login');
      navigate('/');
      return;
    }

    console.log('[ReleaseAdmin] ✅ User authenticated:', user.email);
  }, [user, navigate]);

  // Handle Sync by Bug ID
  const handleSyncBugById = async (e) => {
    e.preventDefault();
    if (!bugId.trim()) {
      setBugIdError('Please enter a Bug ID');
      return;
    }

    setBugIdLoading(true);
    setBugIdError('');
    setBugIdResult(null);

    try {
      console.log('[ReleaseAdmin] Syncing bug by ID:', bugId);
      const result = await syncAPI.syncBugById(bugId);
      setBugIdResult(result);
      setBugId('');
      console.log('[ReleaseAdmin] Bug synced successfully:', result);
    } catch (error) {
      const errorMsg = error.message || 'Failed to sync bug';

      // Check if it's an authorization error
      if (errorMsg.includes('Unauthorized') || errorMsg.includes('401')) {
        console.log('[ReleaseAdmin] ❌ Authorization error - logging out');
        handleUnauthorized();
        return;
      }

      setBugIdError(errorMsg);
      console.error('[ReleaseAdmin] Error syncing bug:', errorMsg);
    } finally {
      setBugIdLoading(false);
    }
  };

  // Handle unauthorized access
  const handleUnauthorized = async () => {
    console.log('[ReleaseAdmin] Handling unauthorized access');

    // Clear localStorage
    localStorage.removeItem('authToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('userEmail');
    localStorage.removeItem('userRole');
    localStorage.removeItem('userId');

    // Dispatch logout to Redux
    dispatch(logout());

    // Redirect to login
    console.log('[ReleaseAdmin] Redirecting to login page');
    navigate('/');
  };

  // Handle Query Blocking Bugs (using Custom Bugsby Query - Query 11)
  const handleQueryBlockingBugs = async (e) => {
    e.preventDefault();
    if (!blockingBugId.trim()) {
      setBlockingError('Please enter a Bug ID');
      return;
    }

    setBlockingLoading(true);
    setBlockingError('');
    setBlockingResult(null);

    try {
      console.log('[ReleaseAdmin] Querying bugs blocking bug ID:', blockingBugId);

      // Build custom query using Query 11 endpoint
      // Query format: blocks==<bugid> to find all bugs that are blocking this bug
      const query = `blocks==${blockingBugId}`;

      const result = await syncAPI.customBugsbyQuery({
        query: query,
        limit: '100',
        sortBy: 'lastUpdateTime',
        order: 'desc'
      });

      setBlockingResult(result);
      setBlockingBugId('');
      console.log('[ReleaseAdmin] Blocking bugs query executed successfully:', result);
    } catch (error) {
      const errorMsg = error.message || 'Failed to query blocking bugs';

      // Check if it's an authorization error
      if (errorMsg.includes('Unauthorized') || errorMsg.includes('401')) {
        console.log('[ReleaseAdmin] ❌ Authorization error - logging out');
        handleUnauthorized();
        return;
      }

      setBlockingError(errorMsg);
      console.error('[ReleaseAdmin] Error querying blocking bugs:', errorMsg);
    } finally {
      setBlockingLoading(false);
    }
  };

  return (
    <div className="release-admin-container">
      <div className="admin-header">
        <h1>🔧 Release Admin Panel</h1>
        <p className="admin-subtitle">Manage Bugsby synchronization tasks</p>
        <p className="user-info">Logged in as: <strong>{user?.email}</strong> ({user?.role})</p>
      </div>

      <div className="sync-tasks-grid">
        {/* Task 1: Sync by Bug ID */}
        <div className="sync-task-card">
          <div className="task-header">
            <h2>🐛 Sync Single Bug</h2>
            <p className="task-description">Fetch and sync a single bug by its Bugsby ID</p>
          </div>

          <form onSubmit={handleSyncBugById} className="sync-form">
            <div className="form-group">
              <label htmlFor="bugId">Bug ID (Bugsby ID)</label>
              <input
                id="bugId"
                type="text"
                placeholder="e.g., 1092263"
                value={bugId}
                onChange={(e) => setBugId(e.target.value)}
                disabled={bugIdLoading}
                className="form-input"
              />
            </div>

            <button
              type="submit"
              disabled={bugIdLoading}
              className="sync-button"
            >
              {bugIdLoading ? '⏳ Syncing...' : '🔄 Sync Bug'}
            </button>
          </form>

          {bugIdError && <div className="error-message">❌ {bugIdError}</div>}
          
          {bugIdResult && (
            <div className="result-box">
              <h3>✅ Sync Result</h3>
              <div className="result-content">
                <p><strong>Bug ID:</strong> {bugIdResult.bugsby_id}</p>
                <p><strong>Title:</strong> {bugIdResult.title}</p>
                <p><strong>Status:</strong> {bugIdResult.status}</p>
                <p><strong>Release:</strong> {bugIdResult.release}</p>
                <p><strong>Severity:</strong> {bugIdResult.severity}</p>
                <p><strong>Last Synced:</strong> {new Date(bugIdResult.last_synced_at).toLocaleString()}</p>
              </div>
            </div>
          )}
        </div>

        {/* Task 2: Query Blocking Bugs */}
        <div className="sync-task-card">
          <div className="task-header">
            <h2>🔗 Query Blocking Bugs</h2>
            <p className="task-description">Find all bugs that are blocking a specific bug</p>
          </div>

          <form onSubmit={handleQueryBlockingBugs} className="sync-form">
            <div className="form-group">
              <label htmlFor="blockingBugId">Bug ID (Bugsby ID) *</label>
              <input
                id="blockingBugId"
                type="text"
                placeholder="e.g., 1229583"
                value={blockingBugId}
                onChange={(e) => setBlockingBugId(e.target.value)}
                disabled={blockingLoading}
                className="form-input"
              />
              <p className="form-hint">Enter a bug ID to find all bugs blocking it</p>
            </div>

            <button
              type="submit"
              disabled={blockingLoading}
              className="sync-button"
            >
              {blockingLoading ? '⏳ Querying...' : '🔍 Find Blocking Bugs'}
            </button>
          </form>

          {blockingError && <div className="error-message">❌ {blockingError}</div>}

          {blockingResult && (
            <div className="result-box">
              <h3>✅ Query Result</h3>
              <div className="result-content">
                {blockingResult.bugs && blockingResult.bugs.length === 0 ? (
                  <p className="no-results">No bugs found blocking this bug</p>
                ) : (
                  <>
                    <p><strong>Bugs Found (This Page):</strong> {blockingResult.bugs ? blockingResult.bugs.length : 0}</p>
                    <p><strong>More Results Available:</strong> {blockingResult.has_next ? 'Yes' : 'No'}</p>

                    {blockingResult.bugs && blockingResult.bugs.length > 0 && (
                      <div className="bugs-list">
                        <p><strong>Blocking Bugs:</strong></p>
                        {blockingResult.bugs.slice(0, 5).map((bug, idx) => (
                          <div key={idx} className="bug-item">
                            <p className="bug-title">• {bug.title}</p>
                            <p className="bug-meta">ID: {bug.id} | Status: {bug.status} | Severity: {bug.severity}</p>
                          </div>
                        ))}
                        {blockingResult.bugs.length > 5 && (
                          <p className="bug-more">... and {blockingResult.bugs.length - 5} more bugs on this page</p>
                        )}
                        {blockingResult.has_next && (
                          <p className="bug-pagination">📄 Use cursor for pagination to see more results</p>
                        )}
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ReleaseAdmin;

