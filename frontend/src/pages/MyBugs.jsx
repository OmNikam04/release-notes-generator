import { useState, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
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
import './MyBugs.css';

const MyBugs = () => {
  const dispatch = useDispatch();
  const { filteredBugs, selectedBug, loading, filters } = useSelector((state) => state.bugs);
  const { user } = useSelector((state) => state.auth);

  useEffect(() => {
    // Fetch bugs on component mount
    dispatch(fetchBugs());
  }, [dispatch]);

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
          <h1>My Bugs</h1>
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
        </div>
      </div>

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
