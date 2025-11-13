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
                  placeholder="Search bugs by title, description, or component..."
                  value={filters.search}
                  onChange={(e) => handleSearchChange(e.target.value)}
                  className="search-input"
                />
              </div>

              <div className="filter-controls">
                <select
                  value={filters.status}
                  onChange={(e) => handleStatusFilterChange(e.target.value)}
                  className="filter-select"
                >
                  <option value="all">All Statuses</option>
                  {getUniqueStatuses().map(status => (
                    <option key={status} value={status}>{status}</option>
                  ))}
                </select>

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

            <div className="bugs-stats">
              <div className="stat-card">
                <span className="stat-number">{filteredBugs.length}</span>
                <span className="stat-label">Total</span>
              </div>
              <div className="stat-card">
                <span className="stat-number">
                  {filteredBugs.filter(bug => bug.status === 'Open').length}
                </span>
                <span className="stat-label">Open</span>
              </div>
              <div className="stat-card">
                <span className="stat-number">
                  {filteredBugs.filter(bug => bug.status === 'In Progress').length}
                </span>
                <span className="stat-label">Progress</span>
              </div>
              <div className="stat-card">
                <span className="stat-number">
                  {filteredBugs.filter(bug => bug.status === 'Resolved').length}
                </span>
                <span className="stat-label">Resolved</span>
              </div>
            </div>
          </div>
        </div>

        <div className="bugs-container">
          <div className="bugs-grid">
            {filteredBugs.length > 0 ? (
              filteredBugs.map(bug => (
                <BugCard
                  key={bug.id}
                  bug={bug}
                  onClick={() => handleBugClick(bug)}
                />
              ))
            ) : (
              <div className="no-bugs-message">
                <h3>No bugs found</h3>
                <p>Try adjusting your search or filter criteria.</p>
              </div>
            )}
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
