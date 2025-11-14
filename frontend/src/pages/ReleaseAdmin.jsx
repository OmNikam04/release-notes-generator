import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { logout } from '../store/slices/authSlice';
import { syncAPI, authAPI, releaseNotesAPI, bugsAPI } from '../services/api';
import './ReleaseAdmin.css';

// Release mapping: release name -> blocking bug ID
const RELEASE_MAPPING = {
  'wifi-ooty-mustfix': '1270576',
  'wifi-ooty-ignore': '1270581',
  'wifi-nainital-mustfix': '1229583',
  'wifi-nainital-ignore': '1229588',
  'wifi-madurai-mustfix': '1156868',
  'wifi-madurai-ignore': '1156873',
};

const ReleaseAdmin = () => {
  const user = useSelector((state) => state.auth.user);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  // State for Sync Single Bug
  const [bugId, setBugId] = useState('');
  const [bugIdLoading, setBugIdLoading] = useState(false);
  const [bugIdResult, setBugIdResult] = useState(null);
  const [bugIdError, setBugIdError] = useState('');

  // State for Sync All Bugs from Release
  const [selectedRelease, setSelectedRelease] = useState('');
  const [releaseLoading, setReleaseLoading] = useState(false);
  const [releaseResult, setReleaseResult] = useState(null);
  const [releaseError, setReleaseError] = useState('');
  const [releaseCurrentPage, setReleaseCurrentPage] = useState(0);

  // Check authentication and authorization on component mount
  useEffect(() => {
    console.log('[ReleaseAdmin] Checking authentication and authorization...');

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

    // Check if user is a manager (only managers can access ReleaseAdmin)
    if (user.role !== 'manager') {
      console.log('[ReleaseAdmin] ❌ User is not a manager - redirecting to bugs page');
      console.log('[ReleaseAdmin] User role:', user.role);
      navigate('/bugs');
      return;
    }

    console.log('[ReleaseAdmin] ✅ User authenticated and authorized:', user.email, '(Manager)');
  }, [user, navigate]);

  // Handle Sync & Generate for Single Bug (combined operation)
  const handleSyncAndGenerateBug = async (e) => {
    e.preventDefault();
    if (!bugId.trim()) {
      setBugIdError('Please enter a Bug ID');
      return;
    }

    setBugIdLoading(true);
    setBugIdError('');
    setBugIdResult(null);

    try {
      console.log('[ReleaseAdmin] Syncing and generating release note for bug ID:', bugId);

      // Step 1: Sync the bug
      const syncResult = await syncAPI.syncBugById(bugId);
      console.log('[ReleaseAdmin] Bug synced successfully:', syncResult);

      // Step 2: Generate release note
      const bugUUID = syncResult.id;
      const generateResult = await releaseNotesAPI.generateReleaseNote(bugUUID);
      console.log('[ReleaseAdmin] Release note generated successfully:', generateResult);

      // Show combined result
      setBugIdResult({
        ...syncResult,
        releaseNoteGenerated: true,
        releaseNoteId: generateResult.data?.id
      });
      setBugId('');
    } catch (error) {
      const errorMsg = error.message || 'Failed to sync and generate release note';

      // Check if it's an authorization error
      if (errorMsg.includes('Unauthorized') || errorMsg.includes('401')) {
        console.log('[ReleaseAdmin] ❌ Authorization error - logging out');
        handleUnauthorized();
        return;
      }

      setBugIdError(errorMsg);
      console.error('[ReleaseAdmin] Error syncing and generating:', errorMsg);
    } finally {
      setBugIdLoading(false);
    }
  };

  // Handle Sync All Bugs from Release
  const handleSyncAllBugsFromRelease = async (e) => {
    e.preventDefault();
    if (!selectedRelease) {
      setReleaseError('Please select a release');
      return;
    }

    setReleaseLoading(true);
    setReleaseError('');
    setReleaseResult(null);

    try {
      const blockingBugId = RELEASE_MAPPING[selectedRelease];
      console.log('[ReleaseAdmin] Syncing all bugs from release:', selectedRelease, 'using blocking bug ID:', blockingBugId);

      // Build custom query using blocking bug ID
      const query = `blocks==${blockingBugId}`;

      const result = await syncAPI.customBugsbyQuery({
        query: query,
        limit: '100',
        sortBy: 'lastUpdateTime',
        order: 'desc'
      });

      setReleaseResult(result);
      setReleaseCurrentPage(0); // Reset to first page
      console.log('[ReleaseAdmin] All bugs synced successfully:', result);
    } catch (error) {
      const errorMsg = error.message || 'Failed to sync all bugs from release';

      // Check if it's an authorization error
      if (errorMsg.includes('Unauthorized') || errorMsg.includes('401')) {
        console.log('[ReleaseAdmin] ❌ Authorization error - logging out');
        handleUnauthorized();
        return;
      }

      setReleaseError(errorMsg);
      console.error('[ReleaseAdmin] Error syncing all bugs:', errorMsg);
    } finally {
      setReleaseLoading(false);
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



  return (
    <div className="release-admin-container">
      <div className="admin-header">
        <div className="header-top">
          <h1>🔧 Release Admin Panel</h1>
          <button
            className="back-button"
            onClick={() => navigate('/bugs')}
            title="Go back to bugs"
          >
            ← Back
          </button>
        </div>
        <p className="admin-subtitle">Manage Bugsby synchronization tasks</p>
        <p className="user-info">Logged in as: <strong>{user?.email}</strong> ({user?.role})</p>
      </div>

      <div className="sync-tasks-grid">
        {/* Task 1: Sync Single Bug */}
        <div className="sync-task-card">
          <div className="task-header">
            <h2>🐛 Sync Single Bug</h2>
            <p className="task-description">Fetch, sync and generate release note for a single bug by its Bugsby ID</p>
          </div>

          <form className="sync-form">
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
              type="button"
              onClick={handleSyncAndGenerateBug}
              disabled={bugIdLoading}
              className="sync-button primary"
            >
              {bugIdLoading ? '⏳ Syncing & Generating...' : '🔄 Sync & Generate'}
            </button>
          </form>

          {bugIdError && <div className="error-message">❌ {bugIdError}</div>}

          {bugIdResult && (
            <div className="result-box">
              <h3>✅ Sync & Generate Result</h3>
              <div className="result-content">
                <p><strong>Bug ID:</strong> {bugIdResult.bugsby_id}</p>
                <p><strong>Title:</strong> {bugIdResult.title}</p>
                <p><strong>Status:</strong> {bugIdResult.status}</p>
                <p><strong>Release:</strong> {bugIdResult.release}</p>
                <p><strong>Severity:</strong> {bugIdResult.severity}</p>
                <p><strong>Last Synced:</strong> {new Date(bugIdResult.last_synced_at).toLocaleString()}</p>
                {bugIdResult.releaseNoteGenerated && (
                  <p><strong style={{ color: '#22c55e' }}>✅ Release Note Generated</strong></p>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Task 2: Sync All Bugs from Release */}
        <div className="sync-task-card">
          <div className="task-header">
            <h2>📦 Sync All Bugs from Release</h2>
            <p className="task-description">Sync all bugs from a particular release (mustfix and ignore bugs)</p>
          </div>

          <form className="sync-form">
            <div className="form-group">
              <label htmlFor="releaseSelect">Select Release</label>
              <select
                id="releaseSelect"
                value={selectedRelease}
                onChange={(e) => setSelectedRelease(e.target.value)}
                disabled={releaseLoading}
                className="form-select"
              >
                <option value="">-- Choose a Release --</option>
                {Object.keys(RELEASE_MAPPING).map((release) => (
                  <option key={release} value={release}>
                    {release}
                  </option>
                ))}
              </select>
              <p className="form-hint">Select a release to sync all mustfix and ignore bugs</p>
            </div>

            <button
              type="button"
              onClick={handleSyncAllBugsFromRelease}
              disabled={releaseLoading || !selectedRelease}
              className="sync-button primary"
            >
              {releaseLoading ? '⏳ Syncing...' : '🔄 Sync All'}
            </button>
          </form>

          {releaseError && <div className="error-message">❌ {releaseError}</div>}

          {releaseResult && (
            <div className="result-box">
              <h3>✅ Sync Result</h3>
              <div className="result-content">
                {releaseResult.bugs && releaseResult.bugs.length === 0 ? (
                  <p className="no-results">No bugs found for this release</p>
                ) : (
                  <>
                    <p><strong>Total Bugs Found:</strong> {releaseResult.bugs ? releaseResult.bugs.length : 0}</p>

                    {releaseResult.bugs && releaseResult.bugs.length > 0 && (
                      <div className="bugs-pagination-container">
                        <div className="bugs-list-scrollable">
                          {releaseResult.bugs.slice(releaseCurrentPage * 5, (releaseCurrentPage + 1) * 5).map((bug, idx) => (
                            <div key={idx} className="bug-item">
                              <p className="bug-title">• {bug.title}</p>
                              <p className="bug-meta">ID: {bug.id} | Status: {bug.status} | Severity: {bug.severity}</p>
                            </div>
                          ))}
                        </div>

                        {releaseResult.bugs.length > 5 && (
                          <div className="pagination-controls">
                            <button
                              className="pagination-btn"
                              onClick={() => setReleaseCurrentPage(Math.max(0, releaseCurrentPage - 1))}
                              disabled={releaseCurrentPage === 0}
                            >
                              ← Previous
                            </button>
                            <span className="pagination-info">
                              Page {releaseCurrentPage + 1} of {Math.ceil(releaseResult.bugs.length / 5)}
                            </span>
                            <button
                              className="pagination-btn"
                              onClick={() => setReleaseCurrentPage(releaseCurrentPage + 1)}
                              disabled={(releaseCurrentPage + 1) * 5 >= releaseResult.bugs.length}
                            >
                              Next →
                            </button>
                          </div>
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

